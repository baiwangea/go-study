package selects

import (
	"errors"
	"fmt"
	"time"

	"go-study/go-concurrency/level"
)

// L2-08：超时控制 —— 下单不能无限等。
func L08() level.Level {
	// slowAPI 模拟一个响应不确定的交易所接口
	slowAPI := func(timeout time.Duration, latency time.Duration) error {
		ch := make(chan struct{})
		go func() {
			time.Sleep(latency)
			close(ch)
		}()

		select {
		case <-ch:
			return nil
		case <-time.After(timeout):
			return errors.New("context deadline exceeded（请求超时）")
		}
	}

	return level.Level{
		ID:      "L2-08",
		Title:   "select + time.After 超时",
		Tags:    "timeout · 快速失败",
		Pre:     "L2-07",
		Goal:    "给任何外部调用套上超时上限，宁可失败也不要无限等待",
		Observe: "接口耗时 100ms < 超时 200ms → 成功；接口耗时 500ms > 超时 200ms → 立刻返回错误",
		Questions: []string{
			"超时触发后，那个仍在 sleep 的 goroutine 怎么样了？（泄漏！这就是为什么真实项目要用 context.WithTimeout，见 L2-11）",
			"time.After 在 for 循环里反复调用会有什么副作用？（每次都会新建 timer，未触发前不会被 GC）",
			"机器人下单超时后应该立即重试，还是先查订单状态？为什么？（防止重复下单）",
		},
		Check: "能写出带超时的调用封装，并说清「超时返回 ≠ 后端没执行」这个坑",
		Run: func() {
			fmt.Println("  情况A：", mustLabel(slowAPI(200*time.Millisecond, 100*time.Millisecond)), "（接口 100ms，超时 200ms）")
			fmt.Println("  情况B：", mustLabel(slowAPI(200*time.Millisecond, 500*time.Millisecond)), "（接口 500ms，超时 200ms）")
			fmt.Println("  ↑ 情况B 只花了约 200ms 就失败返回，而不是等满 500ms —— 快速失败比死等安全")

			time.Sleep(600 * time.Millisecond) // 让两个后台 goroutine 收尾，避免污染下一关输出
		},
	}
}

func mustLabel(err error) string {
	if err != nil {
		return "失败：" + err.Error()
	}
	return "成功：拿到响应"
}
