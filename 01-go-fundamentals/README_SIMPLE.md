# 01-go-fundamentals：直观学习模式

> **每个文件独立运行，无需导航框架，直接看代码学习！**

## 🎯 新的学习方式

### ✅ 现在（直观模式）
```bash
cd 01-go-fundamentals
go run functions/01_basic_func.go    # 直接运行，立即看效果
```

每个文件：
- ✓ 完整的 `package main` + `main()` 函数
- ✓ 清晰的注释说明（核心概念、思考题、通关标准）
- ✓ 格式化的输出（带标题、分隔线）
- ✓ 独立运行，无需理解框架

### ❌ 原来（框架模式）
```bash
go run . 1    # 需要理解 level 框架
```

## 📚 学习顺序（10 关）

### 函数（L1-01 ~ L1-06）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **L1-01** | `functions/01_basic_func.go` | 函数与返回值 | `go run functions/01_basic_func.go` |
| **L1-02** | `functions/02_multi_return.go` | 多返回值与 error | `go run functions/02_multi_return.go` |
| **L1-03** | `functions/03_variadic.go` | 变长参数 | `go run functions/03_variadic.go` |
| **L1-04** | `functions/04_closure.go` | 闭包 | `go run functions/04_closure.go` |
| **L1-05** | `functions/05_recursion.go` | 递归 | `go run functions/05_recursion.go` |
| **L1-06** | `functions/06_defer.go` | defer 延迟执行 | `go run functions/06_defer.go` |

### 包管理（L1-07 ~ L1-08）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **L1-07** | `packages/01_export.go` | 导出规则（首字母大小写） | `go run packages/01_export.go` |
| **L1-08** | `packages/02_import.go` | 导入方式 | `go run packages/02_import.go` |

### 接口（L1-09 ~ L1-10）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **L1-09** | `interfaces/01_define.go` | 接口定义 | `go run interfaces/01_define.go` |
| **L1-10** | `interfaces/02_implicit_impl.go` | 隐式实现 | `go run interfaces/02_implicit_impl.go` |
| **L1-11** | `interfaces/03_polymorphism.go` | 多态与类型断言 | `go run interfaces/03_polymorphism.go` |

## 🚀 快速开始

```bash
# 1. 进入目录
cd 01-go-fundamentals

# 2. 按顺序学习每一关（从函数开始）
go run functions/01_basic_func.go
go run functions/02_multi_return.go
go run functions/03_variadic.go
# ... 依此类推

# 3. 每个文件的结构都一样：
#    - 顶部：核心概念说明
#    - 中间：演示代码
#    - 底部：思考题和通关标准
```

## 💡 学习建议

### 每关的学习步骤
1. **先运行**：看输出结果
2. **看代码**：阅读注释和代码
3. **做实验**：根据思考题修改代码
4. **总结**：用一句话总结收获

### 重点关卡
如果你有 PHP/Java/Python 背景：
- **必学**：L1-02 (多返回值 error)、L1-04 (闭包)、L1-07/08 (包与可见性)、L1-09~11 (接口)
- **可快速浏览**：L1-01、L1-03、L1-05 (语法类似其他语言)

## 🎓 学完这 10 关你会掌握

1. **函数基础**：参数、返回值、多返回值、变长参数
2. **错误处理**：`(result, error)` 模式，不用 try/catch
3. **闭包**：捕获变量而非值
4. **包管理**：首字母大小写 = 唯一可见性规则
5. **接口**：隐式实现，方法集合，多态

## 🔧 关于 //go:build level

文件第一行的 `//go:build level` 是构建标签，作用：
- 允许同一目录有多个 `main()` 函数
- `go build ./...` 时会跳过这些文件（避免冲突）
- 单独运行 `go run 文件名.go` 时会忽略这个标签

**可以删除吗？** 可以！如果你只在一个目录保留一个学习文件，就不需要这个标签。

## 🎯 通关标准（整个模块）

能手写并解释：
1. 返回 `(结果, error)` 的函数
2. 接口存储不同实现
3. 指针修改生效的例子
4. 解释 nil 陷阱

## 🚀 下一站

完成本模块后，前往：
- **L1 指针部分**：`../02-go-pointers/` (L1-11~17)
- 学完整个 L1 后进入 **L2 并发模型**：`../03-go-concurrency/`

---

💡 **提示**：别只"看完"代码就进下一关，必须按思考题改代码验证！
