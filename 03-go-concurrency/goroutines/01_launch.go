package goroutines

import (
	"fmt"
	"sync"
	"time"

	"go-study/go-concurrency/level"
)

// L2-01：go 关键字与「顺序不可预测」。
func L01() level.Level {
	return level.Level{
		ID:      "L2-01",
		Title:   "go 关键字与调度不确定性",
		Tags:    "goroutine · 调度 · 不要假设顺序",
		Pre:     "L1-04（闭包）",
		Goal:    "理解 go 语句是「提交任务后立即返回」，执行顺序由调度器决定，不保证任何顺序",
		Observe: "编号按 4→0 倒序完成（故意用 sleep 造出来的），但 [main] 那行与后台输出的相对顺序完全不保证",
		Questions: []string{
			"把循环里的 go func(){...}() 改成直接调用函数，顺序变成什么？为什么并发时不是这样？",
			"删掉结尾的 time.Sleep，主协程会立刻退出 —— 后台 goroutine 的输出还在吗？（这就是「不等」的坑）",
			"注意 i := i 这一行：Go 1.22 之前不写它会打印什么？为什么？",
		},
		Check: "能说清 go 语句的语义，并解释为什么「靠 sleep 等并发结果」在生产里是错的",
		Run: func() {
			var wg sync.WaitGroup // 这里先借用一下，L2-02 正式讲等待
			wg.Add(1)
			go func() {
				defer wg.Done()
				fmt.Println("  [goroutine] 我在后台跑")
			}()

			for i := 0; i < 5; i++ {
				i := i // Go 1.22 起循环变量每轮新建；旧版本必须手动复制，否则闭包共享同一个 i
				go func() {
					time.Sleep(time.Duration(5-i) * 10 * time.Millisecond) // 故意让完成顺序倒过来
					fmt.Printf("  [goroutine] 编号 %d 完成\n", i)
				}()
			}
			fmt.Println("  [main] 我已经走到这里了 —— go 语句不会等任何人")

			wg.Wait()
			time.Sleep(120 * time.Millisecond) // 借用等待收干净输出，下一关换成正规写法
		},
	}
}
