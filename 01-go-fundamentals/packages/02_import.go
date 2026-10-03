package packages

import (
	"fmt"

	h "go-study/go-fundamentals/packages/helper" // 别名导入：调用时写成 h.XXX

	"go-study/go-fundamentals/level"
)

// L07 第七关：导入形式与包名规则。
func L07() level.Level {
	return level.Level{
		ID:      "L1-07",
		Title:   "四种导入写法",
		Tags:    "import · 别名 · 点导入 · 空白导入",
		Pre:     "L1-06",
		Goal:    "分清普通导入、别名导入、点导入（.）与空白导入（_），知道各自什么时候用",
		Observe: "同一个 helper 包，本文件用 h.XXX 调用，上一关用 helper.XXX 调用",
		Questions: []string{
			"import _ \"database/sql\" 的作用是什么？为什么驱动包常被这样导入？（提示：只为触发 init）",
			"import . \"fmt\" 叫点导入，Println 可以不带包名调用 —— 为什么实际项目里几乎没人这么写？",
			"导入了却没使用，Go 会报错还是警告？这和「未使用的局部变量」是同一条规则吗？",
		},
		Check: "能说出别名导入的适用场景（包名冲突、调用太长），并解释 _ 导入只为执行 init",
		Run: func() {
			fmt.Println("  h.Version         =", h.Version)
			fmt.Println("  h.PublicFunction  =", h.PublicFunction("aliased import works"))
			p := h.NewPair("beta", "s3cr3t")
			fmt.Printf("  h.NewPair(...)    → Name=%q Secret()=%q\n", p.Name, p.Secret())
			fmt.Println("  ↑ 本文件把 helper 改名为 h 使用；包的身份没变，只是调用前缀变了")
		},
	}
}
