// Package notify 通知层：接口先行，Telegram / 邮件 / Webhook 都只是实现。
package notify

import "fmt"

type Notifier interface {
	Notify(msg string)
}

// Console 打印到标准输出，便于观察整条链路。
type Console struct{ Sent int }

func (c *Console) Notify(msg string) {
	c.Sent++
	fmt.Println("   📨 通知：" + msg)
}

// Telegram 是占位实现：接 L3-13 时把 token 从配置/环境变量读取，绝不写死在代码里。
type Telegram struct{ Token string }

func (t Telegram) Notify(msg string) {
	if t.Token == "" {
		fmt.Println("   ⚠️ TG_TOKEN 未设置，通知降级为本地打印：", msg)
		return
	}
	fmt.Println("   → 发送到 Telegram：", msg)
}
