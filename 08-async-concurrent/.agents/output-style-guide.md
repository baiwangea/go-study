# 控制台输出美化规范

## 设计原则
- 使用表情符号作为视觉锚点
- 使用分隔线区分不同部分
- 添加合适的缩进增强层次感
- 关键信息加粗或突出

## 标准表情符号

### 标题和分隔
- `🚀` - 文件主标题
- `----` (60个) - 示例分隔线
- `====` (60个) - 主要区块分隔

### 示例类型
- `📝` - 基础示例
- `🔤` - 字符串处理
- `🔢` - 数字处理
- `🏗️` - 结构体/对象处理
- `⚠️` - 错误处理
- `🐌` - 性能对比（慢）
- `⚡` - 性能优势（快）
- `🎯` - 实战案例
- `🧪` - 测试示例

### 数据流向
- `📥` - 输入数据
- `📤` - 输出结果
- `➡️` - 流程方向
- `🔄` - 循环/重复
- `⏱️` - 时间/耗时

### 状态指示
- `✅` - 成功/正确
- `❌` - 失败/错误
- `⚙️` - 配置/设置
- `📊` - 统计/数据
- `💰` - 金额/价格
- `💡` - 总结/要点
- `🔍` - 查看/检查

### 并发相关
- `🎭` - 并发任务
- `🔐` - 锁/互斥
- `📦` - 池化资源
- `🚦` - 限流/控制
- `🔁` - 重试机制
- `⛓️` - 管道/流水线

## 标准格式模板

### 文件头部
```go
fmt.Println("\n" + strings.Repeat("=", 60))
fmt.Println("🚀 A-XX: 功能名称")
fmt.Println(strings.Repeat("=", 60) + "\n")
```

### 示例开始
```go
fmt.Println("📝 示例 N：示例标题")
fmt.Println(strings.Repeat("-", 60))
```

### 数据输入输出
```go
fmt.Printf("  📥 输入: %v\n", input)
fmt.Printf("  📤 输出: %v\n", output)
fmt.Printf("  ⏱️  耗时: %v\n", duration)
```

### 成功/失败标记
```go
fmt.Printf("     [%d] ✅ 成功: %v\n", id, result)
fmt.Printf("     [%d] ❌ 失败: %v\n", id, err)
```

### 文件结尾
```go
fmt.Println(strings.Repeat("=", 60))
fmt.Println("💡 关键总结")
fmt.Println(strings.Repeat("=", 60))
fmt.Println("  ✅ 要点1")
fmt.Println("  ✅ 要点2")
fmt.Println(strings.Repeat("=", 60))
fmt.Println("➡️  下一关：XX_next_topic.go")
fmt.Println()
```

## 缩进规范
- 主标题：无缩进
- 一级内容：2个空格
- 二级内容：5个空格（3个额外）
- 列表项：5个空格 + 表情/符号

## 特殊场景

### 错误处理展示
```go
fmt.Println("  📊 结果:")
for i, r := range results {
    if r.Err != nil {
        fmt.Printf("     [%d] ❌ 错误: %v\n", i, r.Err)
    } else {
        fmt.Printf("     [%d] ✅ 结果: %v\n", i, r.Value)
    }
}
```

### 性能对比
```go
fmt.Printf("  ⏱️  并发耗时：%v\n", concurrentTime)
fmt.Printf("  ⏱️  串行耗时：%v\n", serialTime)
fmt.Printf("  ⚡ 速度提升：约 %.1f 倍！\n", speedup)
```

### 统计报告
```go
fmt.Printf("📊 统计报告\n")
fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
fmt.Printf("总数:       %d\n", total)
fmt.Printf("成功:       %d (%.1f%%)\n", success, percent)
fmt.Printf("失败:       %d\n", failed)
```

## 导入要求
所有使用此规范的文件必须导入：
```go
import (
    "context"
    "fmt"
    "strings"  // 必需：用于 strings.Repeat()
    "time"
    // ... 其他导入
)
```
