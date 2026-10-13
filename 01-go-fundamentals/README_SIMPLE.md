# 01-go-fundamentals：直观学习模式

> **每个文件独立运行，直接看代码学习。** 控制台用表情把标题 / 示例 / 总结分开。

## 怎么跑

```bash
cd 01-go-fundamentals
go run interfaces/01_define.go    # 直接运行，立即看效果
```

每个文件：
- 完整的 `package main` + `main()`
- 顶部注释：核心概念、思考题、通关标准
- 控制台分区：🎯 标题、📝📌🔍 示例、💡 总结

```bash
make levels M=./01-go-fundamentals
make level  F=errors/01_error_chain.go
```

## 学习顺序（12 关）

### 接口（L1-08 ~ L1-10）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **L1-08** | `interfaces/01_define.go` | 接口 = 方法集合 | `go run interfaces/01_define.go` |
| **L1-09** | `interfaces/02_implicit_impl.go` | 隐式实现 | `go run interfaces/02_implicit_impl.go` |
| **L1-10** | `interfaces/03_polymorphism.go` | 多态与类型断言 | `go run interfaces/03_polymorphism.go` |

### 错误（L1-19 ~ L1-21）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **L1-19** | `errors/01_error_chain.go` | 错误链 Is/As/Unwrap | `go run errors/01_error_chain.go` |
| **L1-20** | `errors/02_custom_error.go` | 自定义错误与哨兵 | `go run errors/02_custom_error.go` |
| **L1-21** | `errors/03_panic_recover.go` | panic / recover | `go run errors/03_panic_recover.go` |

### 字符串（L1-22 ~ L1-23）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **L1-22** | `strings/01_bytes_rune.go` | string / []byte / rune | `go run strings/01_bytes_rune.go` |
| **L1-23** | `strings/02_builder_strconv.go` | Builder 与 strconv | `go run strings/02_builder_strconv.go` |

### 泛型（L1-24 ~ L1-25）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **L1-24** | `generics/01_type_params.go` | 类型参数 | `go run generics/01_type_params.go` |
| **L1-25** | `generics/02_constraints.go` | 约束与取舍 | `go run generics/02_constraints.go` |

### 组合（L1-26 ~ L1-27）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **L1-26** | `composition/01_embedding.go` | 结构体嵌入 | `go run composition/01_embedding.go` |
| **L1-27** | `composition/02_interface_embed.go` | 接口嵌入 | `go run composition/02_interface_embed.go` |

> L1-11~17 在 `../02-go-pointers/`。本目录没有 `functions/`、`packages/`。

## 学习建议

1. **先运行**：看带表情的分区输出
2. **看代码**：读注释和演示
3. **做实验**：按思考题改代码
4. **总结**：用一句话说出这关

## 关于 `//go:build level`

文件第一行的构建标签让同一目录可以有多个 `main()`：
- `go build ./...` 会跳过这些文件（避免冲突）
- `go run 文件名.go` 显式指定文件时仍能跑

## 通关标准（整个模块）

能手写并解释：
1. 接口隐式实现和类型断言
2. `%w` 错误链 + `errors.Is` / `As`
3. `len(string)` 是字节数，中文按 rune 切
4. 结构体嵌入不是继承

## 下一站

- **指针**：`../02-go-pointers/`（L1-11~17）
- 然后 **并发**：`../03-go-concurrency/`
