package helper

import "strings"

// Version 是导出常量：首字母大写，其他包可以直接引用。
const Version = "v1"

// Pair 演示字段级的可见性规则。
type Pair struct {
	Name   string // 导出字段，外部可读可写
	secret string // 非导出字段，只能在 helper 包内访问
}

// Secret 提供访问非导出字段的口子：这叫“受控暴露”，是 Go 里的封装手段。
func (p Pair) Secret() string { return p.secret }

// NewPair 构造一个 Pair（因为外部无法直接设置 secret 字段）。
func NewPair(name, secret string) Pair {
	return Pair{Name: name, secret: secret}
}

// PublicFunction is an example of a function that can be exported.
// In Go, any function or variable that starts with a capital letter is exported (public).
func PublicFunction(text string) string {
	return strings.ToUpper(text)
}

// privateFunction is not visible outside the `helper` package
// because it starts with a lowercase letter.
func privateFunction() string {
	return "this is private"
}
