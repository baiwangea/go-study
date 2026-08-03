package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// ==================== 任务类型常量 ====================
// 命名惯例: domain:action，便于按类型路由和排查

const (
	TypeEmailDelivery   = "email:deliver"   // 邮件投递任务（基础示例）
	TypeSmsNotification = "sms:notify"      // 短信通知任务（批量操作用）
	TypeDataSync        = "data:sync"       // 数据同步任务（定时任务用）
	TypeReportGenerate  = "report:generate" // 报表生成任务（唯一约束用）
	TypeOrderCancel     = "order:cancel"    // 订单超时取消任务（常见延迟任务）
)

// ==================== 邮件任务（基础示例） ====================

// EmailPayload 邮件任务负载
type EmailPayload struct {
	UserID     int    `json:"user_id"`
	TemplateID string `json:"template_id"`
}

// NewEmailTask 创建邮件投递任务
func NewEmailTask(userID int, templateID string) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailPayload{
		UserID:     userID,
		TemplateID: templateID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal email payload: %w", err)
	}
	return asynq.NewTask(TypeEmailDelivery, payload), nil
}

// ==================== 短信通知任务（批量操作用） ====================

// SmsPayload 短信任务负载
type SmsPayload struct {
	Phone   string `json:"phone"`
	Content string `json:"content"`
}

// NewSmsTask 创建短信通知任务
func NewSmsTask(phone, content string) (*asynq.Task, error) {
	payload, err := json.Marshal(SmsPayload{
		Phone:   phone,
		Content: content,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal sms payload: %w", err)
	}
	return asynq.NewTask(TypeSmsNotification, payload), nil
}

// BatchSmsTasks 批量创建短信通知任务
// 用于演示 Asynq 的批量入队能力
func BatchSmsTasks(phones []string, content string) ([]*asynq.Task, error) {
	tasks := make([]*asynq.Task, 0, len(phones))
	for _, phone := range phones {
		task, err := NewSmsTask(phone, content)
		if err != nil {
			return nil, fmt.Errorf("create sms task for %s: %w", phone, err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// ==================== 数据同步任务（定时任务用） ====================

// DataSyncPayload 数据同步任务负载
type DataSyncPayload struct {
	Source string `json:"source"` // 数据源，如 "mysql"、"mongodb"
	Target string `json:"target"` // 目标源，如 "elasticsearch"、"redis"
}

// NewDataSyncTask 创建数据同步任务（用于定时调度）
func NewDataSyncTask(source, target string) (*asynq.Task, error) {
	payload, err := json.Marshal(DataSyncPayload{
		Source: source,
		Target: target,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal data sync payload: %w", err)
	}
	return asynq.NewTask(TypeDataSync, payload), nil
}

// ==================== 报表生成任务（唯一约束用） ====================

// ReportPayload 报表生成任务负载
type ReportPayload struct {
	ReportType string `json:"report_type"` // 报表类型，如 "daily"、"weekly"、"monthly"
	Date       string `json:"date"`        // 报表日期，如 "2026-07-26"
}

// NewReportTask 创建报表生成任务
// TaskID 由业务参数组合而成，确保同类型同日期的报表任务全局唯一
// 这是"任务唯一约束"的核心：通过自定义 TaskID 实现幂等去重
func NewReportTask(reportType, date string) (*asynq.Task, error) {
	payload, err := json.Marshal(ReportPayload{
		ReportType: reportType,
		Date:       date,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal report payload: %w", err)
	}

	// 关键：用业务参数拼接出唯一的 TaskID
	// 好处：相同业务参数的任务重复入队时，Asynq 会拒绝重复的 TaskID
	// 从而避免同一报表被重复生成
	taskID := fmt.Sprintf("report:%s:%s", reportType, date)

	return asynq.NewTask(
		TypeReportGenerate,
		payload,
		asynq.TaskID(taskID), // 设置自定义任务ID，实现唯一约束
	), nil
}

// ==================== 订单超时取消任务（常见延迟任务） ====================

// OrderCancelPayload 订单取消任务负载
type OrderCancelPayload struct {
	OrderID string `json:"order_id"`
}

// NewOrderCancelTask 创建订单超时取消任务
// 常见场景：用户下单后，在支付超时窗口到期时自动取消订单
func NewOrderCancelTask(orderID string) (*asynq.Task, error) {
	payload, err := json.Marshal(OrderCancelPayload{
		OrderID: orderID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal order cancel payload: %w", err)
	}

	return asynq.NewTask(
		TypeOrderCancel,
		payload,
		asynq.TaskID(fmt.Sprintf("order:cancel:%s", orderID)),
	), nil
}
