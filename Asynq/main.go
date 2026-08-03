package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"go-study/Asynq/tasks"

	"github.com/hibiken/asynq"
)

const redisAddr = "127.0.0.1:6379"

func main() {
	// ============================================================
	//  第一部分：创建 Asynq 客户端（生产者）
	// ============================================================
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer client.Close()

	// ============================================================
	//  特性 1：优先级队列 + 延迟执行（基础示例回顾）
	// ============================================================
	log.Println("🎯 === [1/4] 优先级队列与延迟任务 ===")

	// 任务1: 重置密码邮件 - critical 队列（最高优先级）
	task1, err := tasks.NewEmailTask(1, "reset_password")
	if err != nil {
		log.Fatalf("❌ 创建任务失败(重置密码): %v", err)
	}
	info1, err := client.Enqueue(task1, asynq.Queue("critical"))
	if err != nil {
		log.Printf("⚠️  入队失败: %v", err)
	} else {
		log.Printf("✅ 重置密码邮件入队 [critical队列], ID: %s", info1.ID)
	}

	// 任务2: 欢迎邮件 - high 队列
	task2, err := tasks.NewEmailTask(2, "welcome")
	if err != nil {
		log.Fatalf("❌ 创建任务失败(欢迎邮件): %v", err)
	}
	info2, err := client.Enqueue(task2, asynq.Queue("high"))
	if err != nil {
		log.Printf("⚠️  入队失败: %v", err)
	} else {
		log.Printf("✅ 欢迎邮件入队 [high队列], ID: %s", info2.ID)
	}

	// 任务3: 营销邮件 - low 队列，延迟 5 秒执行
	task3, err := tasks.NewEmailTask(3, "marketing")
	if err != nil {
		log.Fatalf("❌ 创建任务失败(营销邮件): %v", err)
	}
	info3, err := client.Enqueue(task3, asynq.Queue("low"), asynq.ProcessIn(5*time.Second))
	if err != nil {
		log.Printf("⚠️  入队失败: %v", err)
	} else {
		log.Printf("✅ 营销邮件入队 [low队列,延迟5s], ID: %s", info3.ID)
	}

	// ============================================================
	//  特性 2：任务唯一约束（TaskID 去重）
	// ============================================================
	log.Println("📊 === [2/4] 任务唯一约束演示 ===")

	today := time.Now().Format("2006-01-02")

	// 第一次入队：生成日报表（应该成功）
	reportTask1, err := tasks.NewReportTask("daily", today)
	if err != nil {
		log.Fatalf("❌ 创建报表任务失败: %v", err)
	}
	infoR1, err := client.Enqueue(reportTask1, asynq.Queue("default"))
	if err != nil {
		log.Printf("❌ 日报表入队失败: %v", err)
	} else {
		log.Printf("✅ 日报表入队成功 [ID: %s]", infoR1.ID)
	}

	// 第二次入队：同样的日报表（应该失败，因为 TaskID 重复）
	reportTask2, err := tasks.NewReportTask("daily", today)
	if err != nil {
		log.Fatalf("❌ 创建报表任务失败: %v", err)
	}
	infoR2, err := client.Enqueue(reportTask2, asynq.Queue("default"))
	if err != nil {
		// 预期会失败：asynq: task id already exists
		log.Printf("🛡️  日报表重复入队被拦截 (预期行为): %v", err)
	} else {
		log.Printf("⚠️  意外: 日报表重复入队成功了 [ID: %s]", infoR2.ID)
	}

	log.Println("💡 原理：通过 asynq.TaskID() 设置自定义任务ID，相同 ID 的任务无法重复入队")

	// ============================================================
	//  特性 3：批量入队操作
	// ============================================================
	log.Println("📨 === [3/4] 批量入队操作演示 ===")

	// 模拟一批手机号（实际场景可能从数据库读取）
	phones := []string{
		"13800138001",
		"13800138002",
		"13800138003",
		"13800138004",
		"13800138005",
	}

	// 批量创建短信任务
	smsTasks, err := tasks.BatchSmsTasks(phones, "您的验证码是 888888，5分钟内有效")
	if err != nil {
		log.Fatalf("❌ 批量创建短信任务失败: %v", err)
	}

	// 批量入队
	// asynq 支持通过 EnqueueContext 批量操作，减少网络往返
	enqueueResults := make([]*asynq.TaskInfo, 0, len(smsTasks))
	for i, t := range smsTasks {
		info, err := client.Enqueue(t, asynq.Queue("default"))
		if err != nil {
			log.Printf("⚠️  短信任务[%d]入队失败: %v", i, err)
			continue
		}
		enqueueResults = append(enqueueResults, info)
	}
	log.Printf("✅ 批量入队完成: 成功 %d / 总共 %d", len(enqueueResults), len(smsTasks))

	log.Println("💡 原理：批量入队减少 Redis 网络往返，提升大规模任务投递效率")

	// ============================================================
	//  特性 4：定时任务调度器（Cron Scheduler）
	// ============================================================
	log.Println("⏰ === [4/4] 定时任务调度器启动 ===")

	// 创建调度器：用于周期性触发任务
	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: redisAddr},
		&asynq.SchedulerOpts{
			// 调度器的日志配置
			LogLevel: asynq.InfoLevel,
		},
	)
	defer scheduler.Shutdown()

	// --- 注册定时任务 1：每 10 秒执行一次数据同步（演示用，实际生产不会这么频繁）
	// Cron 表达式格式（6段）：秒 分 时 日 月 周
	// 标准 cron 是 5 段（分 时 日 月 周），Asynq 使用 robfig/cron/v3，支持 5 段或 6 段
	dataSyncTask, err := tasks.NewDataSyncTask("mysql", "redis")
	if err != nil {
		log.Fatalf("❌ 创建数据同步任务失败: %v", err)
	}

	// 每 10 秒执行一次：@every 10s
	// 解读：第10秒、20秒、30秒...每分钟内每10秒触发
	entryID1, err := scheduler.Register("@every 10s", dataSyncTask,
		asynq.Queue("high"),
		asynq.TaskID("scheduler:datasync:10s"), // 给定时任务也设唯一 ID，避免重复注册
	)
	if err != nil {
		log.Fatalf("❌ 注册定时任务失败: %v", err)
	}
	log.Printf("✅ 定时任务已注册 [数据同步-每10秒] entryID=%v", entryID1)

	// --- 注册定时任务 2：每天凌晨 2 点生成日报表（实际生产常用场景）
	dailyReportTask, err := tasks.NewReportTask("daily", time.Now().Format("2006-01-02"))
	if err != nil {
		log.Fatalf("❌ 创建日报表任务失败: %v", err)
	}

	// 每天凌晨 2 点：0 0 2 * * *
	// 解读：秒=0, 分=0, 时=2, 每天每月每周都触发
	entryID2, err := scheduler.Register("0 0 2 * * *", dailyReportTask,
		asynq.Queue("default"),
	)
	if err != nil {
		log.Printf("⚠️  日报表定时任务注册失败: %v", err)
	} else {
		log.Printf("✅ 定时任务已注册 [日报表-每天02:00] entryID=%v", entryID2)
	}

	// 启动调度器（异步运行，不阻塞）
	if err := scheduler.Start(); err != nil {
		log.Fatalf("❌ 启动调度器失败: %v", err)
	}
	log.Println("⏰ 调度器已启动，定时任务将按 cron 表达式触发...")
	log.Println("💡 原理：Scheduler 是独立组件，定期扫描 cron 任务，到点后自动将任务 Enqueue 到队列中")

	// ============================================================
	//  第二部分：创建 Asynq Worker 服务端（消费者）
	// ============================================================
	log.Println("🚀 === 启动 Asynq Worker 服务 ===")

	concurrency := runtime.NumCPU() * 10
	log.Printf("⚙️  CPU核心数: %d, Worker并发数: %d", runtime.NumCPU(), concurrency)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: concurrency,
			// 队列优先级权重：数值越大，被消费的概率越高
			Queues: map[string]int{
				"critical": 6, // 关键任务，最高权重
				"high":     3, // 高优先级
				"default":  2, // 默认
				"low":      1, // 低优先级
			},
			// 严格优先级模式（可选）：开启后 critical 队列空了才会处理 high，以此类推
			// StrictPriority: true,

			// 每个任务的默认超时时间（可选）
			// Timeout: 30 * time.Minute,

			// 任务处理失败后的默认最大重试次数（可选，默认 25 次）
			// RetryDelayFunc: func(n int, err error, t *asynq.Task) time.Duration {
			//     return asynq.DefaultRetryDelayFunc(n, err, t)
			// },
		},
	)

	// ============================================================
	//  注册任务处理器（ServeMux 模式，类似 net/http）
	// ============================================================
	mux := asynq.NewServeMux()

	// --- 中间件链（按注册顺序执行，洋葱模型）
	mux.Use(LoggingMiddleware)  // 日志中间件：记录耗时和状态
	mux.Use(RecoveryMiddleware) // Panic 恢复中间件：防止单个任务 panic 导致 Worker 崩溃

	// --- 注册各任务类型的处理器
	mux.HandleFunc(tasks.TypeEmailDelivery, handleEmailTask)   // 邮件投递
	mux.HandleFunc(tasks.TypeSmsNotification, handleSmsTask)   // 短信通知
	mux.HandleFunc(tasks.TypeDataSync, handleDataSyncTask)     // 数据同步
	mux.HandleFunc(tasks.TypeReportGenerate, handleReportTask) // 报表生成

	// ============================================================
	//  优雅关闭：监听系统信号
	// ============================================================
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("🛑 收到信号 %v，正在优雅关闭 Worker...", sig)
		// srv.Shutdown() 会等待正在执行的任务完成后再退出
		// 注意：在 Run 调用中用 Shutdown 方式略有不同，这里仅作演示
	}()

	// 启动 Worker（阻塞运行，直到收到终止信号）
	log.Println("🎬 Worker 开始运行，按 Ctrl+C 停止...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("💥 Worker 服务异常退出: %v", err)
	}

	log.Println("👋 Worker 已安全退出")
}

// ============================================================
//  任务处理器
// ============================================================

// handleEmailTask 邮件投递任务处理器
func handleEmailTask(ctx context.Context, t *asynq.Task) error {
	taskID, _ := asynq.GetTaskID(ctx)
	log.Printf("🔧 [邮件] 开始处理 - ID: %s", taskID)

	// 1. 幂等性检查（内存版，生产环境建议用 Redis SETNX）
	if isTaskExecuted(taskID) {
		log.Printf("🔄 [邮件] 任务已执行过，跳过 - ID: %s", taskID)
		return nil
	}

	// 2. 解析负载
	var payload tasks.EmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		log.Printf("❌ [邮件] 解析负载失败 - ID: %s, 错误: %v", taskID, err)
		return fmt.Errorf("unmarshal payload: %w", err)
	}
	log.Printf("📦 [邮件] 负载解析 - 用户ID: %d, 模板: %s", payload.UserID, payload.TemplateID)

	// 3. 模拟发邮件
	time.Sleep(500 * time.Millisecond)
	switch payload.TemplateID {
	case "welcome":
		log.Printf("📧 [邮件] 发送欢迎邮件给用户 %d", payload.UserID)
	case "reset_password":
		log.Printf("🔑 [邮件] 发送重置密码邮件给用户 %d", payload.UserID)
	case "marketing":
		log.Printf("🎁 [邮件] 发送营销邮件给用户 %d", payload.UserID)
	default:
		log.Printf("🔔 [邮件] 发送通知邮件给用户 %d", payload.UserID)
	}

	// 4. 标记已执行
	markTaskExecuted(taskID)

	log.Printf("✅ [邮件] 任务完成 - ID: %s", taskID)
	return nil
}

// handleSmsTask 短信通知任务处理器
func handleSmsTask(ctx context.Context, t *asynq.Task) error {
	taskID, _ := asynq.GetTaskID(ctx)

	// 解析负载
	var payload tasks.SmsPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		log.Printf("❌ [短信] 解析负载失败 - ID: %s, 错误: %v", taskID, err)
		return fmt.Errorf("unmarshal sms payload: %w", err)
	}

	// 模拟发短信（实际调用短信服务商 API）
	time.Sleep(200 * time.Millisecond)
	log.Printf("📱 [短信] 发送成功 - 手机号: %s, 内容: %s", payload.Phone, payload.Content)

	return nil
}

// handleDataSyncTask 数据同步任务处理器（由定时调度器触发）
func handleDataSyncTask(ctx context.Context, t *asynq.Task) error {
	taskID, _ := asynq.GetTaskID(ctx)

	// 解析负载
	var payload tasks.DataSyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		log.Printf("❌ [数据同步] 解析负载失败 - ID: %s, 错误: %v", taskID, err)
		return fmt.Errorf("unmarshal datasync payload: %w", err)
	}

	log.Printf("🔄 [数据同步] 开始同步 %s -> %s", payload.Source, payload.Target)

	// 模拟同步过程
	time.Sleep(1 * time.Second)

	log.Printf("✅ [数据同步] 同步完成 %s -> %s", payload.Source, payload.Target)
	return nil
}

// handleReportTask 报表生成任务处理器
// 演示了唯一约束 + 业务幂等的双重保障
func handleReportTask(ctx context.Context, t *asynq.Task) error {
	taskID, _ := asynq.GetTaskID(ctx)

	// 解析负载
	var payload tasks.ReportPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		log.Printf("❌ [报表] 解析负载失败 - ID: %s, 错误: %v", taskID, err)
		return fmt.Errorf("unmarshal report payload: %w", err)
	}

	log.Printf("📊 [报表] 开始生成 - 类型: %s, 日期: %s", payload.ReportType, payload.Date)

	// 模拟报表生成（耗时操作）
	time.Sleep(2 * time.Second)

	log.Printf("✅ [报表] 生成完成 - 类型: %s, 日期: %s", payload.ReportType, payload.Date)
	return nil
}

// ============================================================
//  中间件
// ============================================================

// LoggingMiddleware 日志中间件
// 记录每个任务的开始时间、结束时间、耗时和执行结果
func LoggingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		start := time.Now()
		taskID, _ := asynq.GetTaskID(ctx)
		log.Printf("⏱️  [中间件-日志] 任务开始 - 类型: %s, ID: %s", t.Type(), taskID)

		err := h.ProcessTask(ctx, t)

		duration := time.Since(start)
		if err != nil {
			log.Printf("❌ [中间件-日志] 任务失败 - 类型: %s, ID: %s, 耗时: %v, 错误: %v",
				t.Type(), taskID, duration, err)
		} else {
			log.Printf("✅ [中间件-日志] 任务完成 - 类型: %s, ID: %s, 耗时: %v",
				t.Type(), taskID, duration)
		}

		return err
	})
}

// RecoveryMiddleware Panic 恢复中间件
// 捕获任务处理中的 panic，防止单个任务异常导致整个 Worker 进程崩溃
// 这是生产环境必备的中间件
func RecoveryMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) (err error) {
		defer func() {
			if r := recover(); r != nil {
				taskID, _ := asynq.GetTaskID(ctx)
				log.Printf("💥 [中间件-恢复] 捕获 panic - 类型: %s, ID: %s, panic: %v",
					t.Type(), taskID, r)
				// 将 panic 转为 error，交给 asynq 的重试机制处理
				err = fmt.Errorf("task panic: %v", r)
			}
		}()

		return h.ProcessTask(ctx, t)
	})
}

// ============================================================
//  幂等性辅助（内存版，仅作演示）
// ============================================================

// taskExecuted 任务执行记录（内存 map）
// 生产环境建议用 Redis SET key value EX ttl NX 实现分布式幂等
var taskExecuted = make(map[string]bool)

// isTaskExecuted 检查任务是否已执行
func isTaskExecuted(taskID string) bool {
	return taskExecuted[taskID]
}

// markTaskExecuted 标记任务已执行
func markTaskExecuted(taskID string) {
	taskExecuted[taskID] = true
}
