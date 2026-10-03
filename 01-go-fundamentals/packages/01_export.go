package packages

import (
	"fmt"

	"go-study/go-fundamentals/level"
	"go-study/go-fundamentals/packages/helper"
)

// L06 第六关：包与导出规则（首字母大小写就是可见性）。
func L06() level.Level {
	return level.Level{
		ID:      "L1-06",
		Title:   "导出与封装",
		Tags:    "package · 可见性",
		Pre:     "L1-01",
		Goal:    "记住 Go 唯一的可见性规则：标识符首字母大写=对外可见，小写=包内私有",
		Observe: "helper.Version / PublicFunction / NewPair 都能调用；被注释掉的两行则会编译失败",
		Questions: []string{
			"把源码里注释掉的两行（helper.privateFunction() 与 _ = p.secret）取消注释，两条报错信息分别是什么？",
			"为什么外部能读 p.Name 却不能读 p.secret？想用 secret 只能走 p.Secret() —— 这种「受控暴露」和你熟悉的 getter/setter 有何异同？",
			"同一个目录里放两个 package 名不同的文件会怎样？（提示：一个目录 = 一个包）",
		},
		Check: "能说清「大写导出」这条唯一规则，并解释为什么需要 NewPair 这样的构造函数",
		Run: func() {
			fmt.Println("  helper.Version          =", helper.Version)
			fmt.Println("  helper.PublicFunction(...) =", helper.PublicFunction("hello from helper package"))

			p := helper.NewPair("alpha", "hidden-token")
			fmt.Printf("  p.Name = %q（导出字段，可直接访问）\n", p.Name)
			fmt.Printf("  p.Secret() = %q（非导出字段，只能走方法）\n", p.Secret())
			// 下面两行故意保持注释：取消注释会得到 undefined / cannot refer to unexported 错误
			// helper.privateFunction()
			// _ = p.secret
			fmt.Println("  源码里两行注释掉的调用是故意的：取消注释即可看到编译错误")
		},
	}
}
