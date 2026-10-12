//go:build level

package main

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("=== 常量 vs 变量 比较 ===")
	fmt.Println("常量 0.1+0.2 == 0.3 :", 0.1+0.2 == 0.3)
	a, b := 0.1, 0.2
	fmt.Println("变量 a+b == 0.3     :", a+b == 0.3)
	var c, d float64 = 0.1, 0.2
	fmt.Println("显式 float64 c+d==0.3:", c+d == 0.3)

	fmt.Println("\n=== NewFromFloat 陷阱候选 ===")
	cases := []struct {
		name string
		v    float64
	}{
		{"0.1+0.7", func() float64 { x, y := 0.1, 0.7; return x + y }()},
		{"0.3-0.2", func() float64 { x, y := 0.3, 0.2; return x - y }()},
		{"1.0-0.9", func() float64 { x, y := 1.0, 0.9; return x - y }()},
		{"4.35*100", func() float64 { x, y := 4.35, 100.0; return x * y }()},
		{"0.1*3", func() float64 { x, y := 0.1, 3.0; return x * y }()},
		{"2.675*100", func() float64 { x, y := 2.675, 100.0; return x * y }()},
		{"1e18+1", func() float64 { return 1e18 + 1 }()},
		{"(0.1+0.2)*1e18", func() float64 { x, y := 0.1, 0.2; return (x + y) * 1e18 }()},
	}
	for _, c := range cases {
		dd := decimal.NewFromFloat(c.v)
		fmt.Printf("  %-14s float=%-22.17g decimal=%-24s StringFixed(8)=%s\n",
			c.name, c.v, dd.String(), dd.StringFixed(8))
	}

	fmt.Println("\n=== 正确写法 ===")
	dd := decimal.RequireFromString("0.1").Add(decimal.RequireFromString("0.7"))
	fmt.Println("NewFromString(\"0.1\").Add(\"0.7\") =", dd.String())

	fmt.Println("\n=== 1e18+1 ===")
	fmt.Println("float64 1e18+1 == 1e18 :", 1e18+1 == 1e18)
	fmt.Println("decimal 1e18+1 == 1e18 :", decimal.NewFromFloat(1e18).Add(decimal.NewFromInt(1)).Equal(decimal.NewFromFloat(1e18)))
	fmt.Println("decimal(1e18)+1 =", decimal.NewFromFloat(1e18).Add(decimal.NewFromInt(1)).String())

	fmt.Println("\n=== NewFromFloat 对 2^53 附近 ===")
	fmt.Println("NewFromFloat(9007199254740993) =", decimal.NewFromFloat(9007199254740993).String())
	fmt.Println("decimal.NewFromInt(9007199254740993) =", decimal.NewFromInt(9007199254740993).String())
}
