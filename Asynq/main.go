package main

import (
	"context"
	"encoding/json"
	"log"
	"runtime"
	"time"

	"go-study/Asynq/tasks"

	"github.com/hibiken/asynq"
)

const redisAddr = "127.0.0.1:6379"

func main() {
	// 1. 创建Asynq客户端
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer client.Close()

	// 2. 创建并批量入队多个任务
	log.Println("🎯 === 创建并入队任务 ===")

	// 任务1: 重置密码邮件 - 放入 critical 队列（最高优先级）
	task1, err := tasks.NewEmailTask(1, "reset_password")
	if err != nil {
		log.Fatalf("❌ 创建任务失败(重置密码): %v", err)
	}
	info1, _ := client.Enqueue(task1, asynq.Queue("critical"))
	log.Printf("✅ 重置密码邮件入队成功 🔑 [critical队列], ID: %s", info1.ID)

	// 任务2: 欢迎邮件 - 放入 high 队列（高优先级）
	task2, err := tasks.NewEmailTask(2, "welcome")
	if err != nil {
		log.Fatalf("❌ 创建任务失败(欢迎邮件): %v", err)
	}
	info2, _ := client.Enqueue(task2, asynq.Queue("high"))
	log.Printf("✅ 欢迎邮件入队成功 📧 [high队列], ID: %s", info2.ID)

	// 任务3: 通知邮件 - 放入 default 队列（默认优先级）
	task3, err := tasks.NewEmailTask(3, "notification")
	if err != nil {
		log.Fatalf("❌ 创建任务失败(通知邮件): %v", err)
	}
	info3, _ := client.Enqueue(task3, asynq.Queue("default"))
	log.Printf("✅ 通知邮件入队成功 🔔 [default队列], ID: %s", info3.ID)

	// 任务4: 营销邮件 - 放入 low 队列（低优先级，延迟5秒执行）
	task4, err := tasks.NewEmailTask(4, "marketing")
	if err != nil {
		log.Fatalf("❌ 创建任务失败(营销邮件): %v", err)
	}
	info4, _ := client.Enqueue(task4, asynq.Queue("low"), asynq.ProcessIn(5))
	log.Printf("✅ 营销邮件入队成功 🎉 [low队列,延迟5秒], ID: %s", info4.ID)

	log.Println("🎉 === 所有任务入队完成 ===")

	// 3. 创建Asynq服务端（Worker）
	// 并发数设置为 CPU 核心数的 10 倍，充分利用多核优势
	concurrency := runtime.NumCPU() * 10
	log.Printf("⚙️ CPU核心数: %d, 并发数: %d", runtime.NumCPU(), concurrency)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: concurrency,
			// 设置队列优先级: high > default > low
			Queues: map[string]int{
				"critical": 6, // 关键任务，权重最高
				"high":     3, // 高优先级
				"default":  2, // 默认
				"low":      1, // 低优先级
			},
		},
	)

	// 4. 注册任务处理器（使用中间件）
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeEmailDelivery, LoggingMiddleware(asynq.HandlerFunc(handleEmailTask)))

	// 5. 启动Worker服务
	log.Println("🚀 === 启动Asynq Worker服务 ===")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("💥 Worker服务启动失败: %v", err)
	}
}

func handleEmailTask(ctx context.Context, t *asynq.Task) error {
	taskID := asynq.GetTaskID(ctx)
	log.Printf("🔧 [任务处理] 开始处理任务 - 任务ID: %s, 任务类型: %s", taskID, t.Type())

	// 1. 幂等性检查：检查任务是否已执行过
	if isTaskExecuted(taskID) {
		log.Printf("🔄 [任务处理] 任务已执行过，跳过 - 任务ID: %s", taskID)
		return nil
	}

	// 2. 解析任务负载数据
	var payload struct {
		UserID     int    `json:"user_id"`
		TemplateID string `json:"template_id"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		log.Printf("❌ [任务处理] 解析负载失败 - 任务ID: %s, 错误: %v", taskID, err)
		return err
	}
	log.Printf("📦 [任务处理] 解析负载成功 - 任务ID: %s, 用户ID: %d, 模板ID: %s", taskID, payload.UserID, payload.TemplateID)

	// 3. 根据模板类型发送不同邮件
	err := sendEmail(payload.UserID, payload.TemplateID)
	if err != nil {
		log.Printf("💥 [任务处理] 发送邮件失败 - 任务ID: %s, 用户ID: %d, 错误: %v", taskID, payload.UserID, err)
		return err
	}

	// 4. 标记任务已执行（保证幂等性）
	markTaskExecuted(taskID)

	// 5. 任务处理成功
	log.Printf("✅ [任务处理] 任务完成 - 任务ID: %s, 用户ID: %d, 模板ID: %s", taskID, payload.UserID, payload.TemplateID)
	return nil
}

// sendEmail 根据模板ID发送邮件
func sendEmail(userID int, templateID string) error {
	log.Printf("📤 [邮件发送] 用户ID: %d, 模板: %s", userID, templateID)

	// 根据不同模板执行不同的发送逻辑
	switch templateID {
	case "welcome":
		log.Printf("🌟 [邮件发送] ✅ 发送欢迎邮件给用户 %d: 欢迎注册我们的服务！🎉", userID)
	case "notification":
		log.Printf("🔔 [邮件发送] ✅ 发送通知邮件给用户 %d: 您有新的系统通知。📢", userID)
	case "reset_password":
		log.Printf("🔑 [邮件发送] ✅ 发送重置密码邮件给用户 %d: 点击链接重置密码。🔗", userID)
	case "marketing":
		log.Printf("🎁 [邮件发送] ✅ 发送营销邮件给用户 %d: 最新优惠活动等你来！🎊", userID)
	default:
		log.Printf("❓ [邮件发送] ❌ 未知模板类型: %s, 用户ID: %d", templateID, userID)
	}

	return nil
}

// LoggingMiddleware 日志中间件 - 记录任务执行时间和状态
func LoggingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		start := time.Now()
		log.Printf("⏱️ [中间件] 任务开始执行 - 类型: %s", t.Type())

		err := h.ProcessTask(ctx, t)

		duration := time.Since(start)
		if err != nil {
			log.Printf("❌ [中间件] 任务执行失败 - 类型: %s, 耗时: %v, 错误: %v", t.Type(), duration, err)
		} else {
			log.Printf("✅ [中间件] 任务执行完成 - 类型: %s, 耗时: %v", t.Type(), duration)
		}

		return err
	})
}

// 任务执行记录（模拟幂等性存储）
var taskExecuted = make(map[string]bool)

// isTaskExecuted 检查任务是否已执行（幂等性检查）
func isTaskExecuted(taskID string) bool {
	return taskExecuted[taskID]
}

// markTaskExecuted 标记任务已执行
func markTaskExecuted(taskID string) {
	taskExecuted[taskID] = true
}
