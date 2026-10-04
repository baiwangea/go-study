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

	"go-study/goframe/level"
)

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

// L3-13：Telegram 通知（成交/异常推送）。
func L13() level.Level {
	return level.Level{
		ID:      "L3-13",
		Title:   "TG 通知：把异常推给自己",
		Tags:    "Bot API · form 编码 · 两层成功",
		Pre:     "L3-02, L3-05",
		Goal:    "机器人跑在 VPS 上没人盯着，异常必须主动推到你手机 —— 这关打通 sendMessage",
		Observe: "未配 TG_BOT_TOKEN 时明确跳过并给出设置方法；配了则真实送达（HTTP 200 + ok=true）",
		Questions: []string{
			"为什么 HTTP 200 还不够，必须再看 body 的 ok 字段？（对照 L3-03 的统一响应中间件）",
			"通知失败该不该重试？无限重试会不会因为一次故障给你刷 500 条消息？（提示：冷却 + 每日配额）",
			"Token 从环境变量读而不是写进代码 —— 如果非要留在仓库里会怎样？（对照 L3-04 的配置策略）",
			"机器人里「下单失败」和「行情断线」的通知优先级该怎么区分？",
		},
		Check: "会调外部 HTTP API 并正确判定成功/失败，且通知失败不会拖垮主流程",
		Run: func() {
			token, chat := os.Getenv("TG_BOT_TOKEN"), os.Getenv("TG_CHAT_ID")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			fmt.Println("  先看失败路径（不依赖 token，随时可复现）：")
			if err := sendTelegram(ctx, "123456:demo", "1", "ping"); err != nil {
				fmt.Println("    ✓ 无效 token 被正确判为失败：", firstLine(err.Error()))
			}

			if token == "" || chat == "" {
				fmt.Println("\n  未配置 TG_BOT_TOKEN / TG_CHAT_ID，本关跳过真实发送。设置方法：")
				fmt.Println("    export TG_BOT_TOKEN=<BotFather 给的 token>")
				fmt.Println("    export TG_CHAT_ID=<你的 chat id（给 @userinfobot 发条消息即可拿到）>")
				fmt.Println("  配好后重跑 go run . 13 就会真实推送。")
				return
			}
			if err := sendTelegram(ctx, token, chat, "✅ go-study L3-13：通知链路已打通"); err != nil {
				fmt.Println("    ✗ 发送失败：", err)
				return
			}
			fmt.Println("    ✓ 已发送到你的 Telegram，去看一眼")
		},
	}
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
