// Package notify 通知层：接口先行，Telegram / 邮件 / Webhook 都只是实现。
//
// 交易机器人的铁律：**通知挂了不能影响下单**。所以 Notifier 接口没有返回值 ——
// 所有实现都不返回错误、不 panic、不长时间阻塞，失败只在本地记一行日志。
// 需要拿到错误的调用方（比如测试）可以直接用 Telegram.Send。
package notify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Notifier interface {
	Notify(msg string)
}

// Console 打印到标准输出，便于观察整条链路。
type Console struct{ Sent int }

func (c *Console) Notify(msg string) {
	c.Sent++
	fmt.Println("   📨 通知：" + msg)
}

// Telegram 的默认参数。BaseURL 做成字段就是为了能注入：测试里用 httptest 起假 TG 服务，
// 绝不真的连 api.telegram.org。
const (
	DefaultBaseURL = "https://api.telegram.org"
	DefaultTimeout = 4 * time.Second // 单次请求 4s：够发一条消息，又不至于把管线拖死
)

// 环境变量：新名（与 04-goframe L3-13、根 README 一致）优先，旧名作为兼容别名。
const (
	EnvToken    = "TG_BOT_TOKEN" // 新名
	EnvTokenOld = "TG_TOKEN"     // 旧名，仅在新名未设置时回落
	EnvChatID   = "TG_CHAT_ID"
)

var (
	// ErrNoToken / ErrNoChatID 让调用方能明确区分「没配」和「发了但失败」。
	ErrNoToken  = errors.New(EnvToken + " 未设置")
	ErrNoChatID = errors.New(EnvChatID + " 未设置")
)

// TokenFromEnv 读 bot token：优先 TG_BOT_TOKEN，未设置时回落到旧的 TG_TOKEN。
func TokenFromEnv() string {
	if v := strings.TrimSpace(os.Getenv(EnvToken)); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv(EnvTokenOld))
}

// ChatIDFromEnv 读 chat id（没有历史别名）。
func ChatIDFromEnv() string { return strings.TrimSpace(os.Getenv(EnvChatID)) }

// NewTelegramFromEnv 从环境变量装配一个 Telegram 通知器，供 main 使用。
func NewTelegramFromEnv() Telegram {
	return Telegram{Token: TokenFromEnv(), ChatID: ChatIDFromEnv()}
}

// Telegram 真实调用 Bot API 的 sendMessage。
//
// Token / ChatID / BaseURL / Timeout / Client 全部可注入，测试不需要网络也不需要密钥。
type Telegram struct {
	Token   string
	ChatID  string
	BaseURL string        // 留空用 DefaultBaseURL
	Timeout time.Duration // 留空用 DefaultTimeout
	Client  *http.Client  // 留空则按 Timeout 新建
}

func (t Telegram) baseURL() string {
	if strings.TrimSpace(t.BaseURL) == "" {
		return DefaultBaseURL
	}
	return strings.TrimRight(t.BaseURL, "/")
}

func (t Telegram) timeout() time.Duration {
	if t.Timeout <= 0 {
		return DefaultTimeout
	}
	return t.Timeout
}

func (t Telegram) client() *http.Client {
	if t.Client != nil {
		return t.Client
	}
	return &http.Client{Timeout: t.timeout()}
}

// Notify 实现 Notifier：**永远不返回错误、不 panic、不阻塞主流程**。
//
//   - 缺 token / 缺 chat id → 降级为本地打印，并说明怎么配置；
//   - 发送失败 → 只记一行日志，主流程继续跑（下单绝不受影响）。
func (t Telegram) Notify(msg string) {
	switch {
	case t.Token == "":
		fmt.Println("   ⚠️ " + EnvToken + " 未设置，通知降级为本地打印：" + msg)
		fmt.Printf("      ↳ 配置方法：export %s=<BotFather 给的 token> %s=<你的 chat id（给 @userinfobot 发条消息即可拿到）>\n",
			EnvToken, EnvChatID)
		return
	case t.ChatID == "":
		fmt.Println("   ⚠️ " + EnvChatID + " 未设置，通知降级为本地打印：" + msg)
		fmt.Printf("      ↳ 配置方法：export %s=<你的 chat id（给 @userinfobot 发条消息即可拿到）>\n", EnvChatID)
		return
	}

	if err := t.Send(msg); err != nil {
		// 一行日志，带上原消息（否则复盘时不知道是哪条通知丢了），然后照常返回。
		fmt.Println("   ⚠️ Telegram 发送失败（不影响下单，管线继续）：" +
			firstLine(err.Error()) + " ｜ 原消息：" + msg)
		return
	}
	fmt.Println("   📨 已发送到 Telegram：" + msg)
}

// Send 真正 POST 到 Bot API 并返回错误，方便调用方与测试判定成败。
func (t Telegram) Send(msg string) error {
	return t.SendContext(context.Background(), msg)
}

// SendContext 是带 ctx 的发送：单次请求同时受 ctx 与 Timeout 约束，先到先算。
func (t Telegram) SendContext(ctx context.Context, msg string) error {
	if t.Token == "" {
		return ErrNoToken
	}
	if t.ChatID == "" {
		return ErrNoChatID
	}

	ctx, cancel := context.WithTimeout(ctx, t.timeout())
	defer cancel()

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", t.baseURL(), t.Token)
	form := url.Values{"chat_id": {t.ChatID}, "text": {msg}}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client().Do(req)
	if err != nil {
		return fmt.Errorf("网络层失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	// 两层成功：HTTP 200 只代表送达，真正的结果在 body 的 ok 字段里（对照 L3-03 / L3-13）。
	var out struct {
		OK   bool   `json:"ok"`
		Desc string `json:"description"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("HTTP %d 且响应无法解析: %s", resp.StatusCode, firstLine(string(body)))
	}
	if resp.StatusCode != http.StatusOK || !out.OK {
		return fmt.Errorf("Telegram 返回失败 http=%d ok=%v desc=%s", resp.StatusCode, out.OK, out.Desc)
	}
	return nil
}

// firstLine 取首行并截断，避免把整段 HTML/JSON 错误页打到终端上。
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if len(s) > 120 {
		return s[:120] + "…"
	}
	return s
}
