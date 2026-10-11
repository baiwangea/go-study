//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 04-goframe && go run 13_telegram_notify.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L3-13：TG 通知：把异常推给自己
═══════════════════════════════════════════════════════════════════════

【核心概念】
    机器人跑在 VPS 上没人盯着，异常必须主动推到你手机 —— 这关打通 sendMessage。
    注意「两层成功」：HTTP 200 只代表送达，真正的结果在 body 的 ok 字段里。

【你会看到】
    ✓ 未配 TG_BOT_TOKEN 时明确跳过并给出设置方法；配了则真实送达（HTTP 200 + ok=true）

【思考题】（动手改代码验证，别空想）
    1. 为什么 HTTP 200 还不够，必须再看 body 的 ok 字段？（对照 L3-03 的统一响应中间件）
    2. 通知失败该不该重试？无限重试会不会因为一次故障给你刷 500 条消息？（提示：冷却 + 每日配额）
    3. Token 从环境变量读而不是写进代码 —— 如果非要留在仓库里会怎样？（对照 L3-04 的配置策略）
    4. 机器人里「下单失败」和「行情断线」的通知优先级该怎么区分？

【通关标准】
    会调外部 HTTP API 并正确判定成功/失败，且通知失败不会拖垮主流程

【前置】L3-02, L3-05      【标签】Bot API · form 编码 · 两层成功
═══════════════════════════════════════════════════════════════════════
*/

// sendTelegram 调用 Bot API 发消息。注意两层成功：HTTP 200 只代表送达，
// 真正的结果在 body 的 ok 字段里（对照 L3-03：err == nil 不等于成功）。
func sendTelegram(ctx context.Context, token, chatID, text string) error {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	form := url.Values{"chat_id": {chatID}, "text": {text}}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("网络层失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var out struct {
		OK  bool   `json:"ok"`
		Msg string `json:"description"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("HTTP %d 且响应无法解析: %s", resp.StatusCode, firstLine(string(body)))
	}
	if resp.StatusCode != http.StatusOK || !out.OK {
		return fmt.Errorf("Telegram 返回失败 http=%d ok=%v desc=%s", resp.StatusCode, out.OK, out.Msg)
	}
	return nil
}

// firstLine 取错误信息首行，避免把整段 HTML/JSON 打到终端上。
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if len(s) > 120 {
		return s[:120] + "…"
	}
	return s
}

func main() {
	fmt.Println("\n=== L3-13 TG 通知：把异常推给自己 ===")
	token, chat := os.Getenv("TG_BOT_TOKEN"), os.Getenv("TG_CHAT_ID")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ── 示例 1：失败路径（不依赖 token，随时可复现） ─────────────────────
	fmt.Println("\n【示例 1】失败路径：无效 token 被正确判为失败")
	fmt.Println("  先看失败路径（不依赖 token，随时可复现）：")
	if err := sendTelegram(ctx, "123456:demo", "1", "ping"); err != nil {
		fmt.Println("    ✓ 无效 token 被正确判为失败：", firstLine(err.Error()))
	}

	// ── 示例 2：真实发送（需要 TG_BOT_TOKEN / TG_CHAT_ID） ───────────────
	if token == "" || chat == "" {
		fmt.Println("\n  未配置 TG_BOT_TOKEN / TG_CHAT_ID，本关跳过真实发送。设置方法：")
		fmt.Println("    export TG_BOT_TOKEN=<BotFather 给的 token>")
		fmt.Println("    export TG_CHAT_ID=<你的 chat id（给 @userinfobot 发条消息即可拿到）>")
		fmt.Println("  配好后重跑 go run 13_telegram_notify.go 就会真实推送。")
		return
	}
	fmt.Println("\n【示例 2】真实发送")
	if err := sendTelegram(ctx, token, chat, "✅ go-study L3-13：通知链路已打通"); err != nil {
		fmt.Println("    ✗ 发送失败：", err)
		return
	}
	fmt.Println("    ✓ 已发送到你的 Telegram，去看一眼")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • 两层成功：HTTP 200 只是送达，业务结果在 body 的 ok 字段里")
	fmt.Println("  • token 从环境变量读，不进 Git；失败要有冷却与配额，避免故障时刷屏")
	fmt.Println("  • 通知失败只记录不 panic，绝不能拖垮主流程")
}
