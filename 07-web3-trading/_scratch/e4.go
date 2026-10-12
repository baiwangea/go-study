//go:build level

package main

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("=== 常量折叠 vs 运行时 ===")
	fmt.Printf("字面量 0.1+0.2            = %.20f\n", 0.1+0.2)
	a, b := 0.1, 0.2
	fmt.Printf("变量   a+b                = %.20f\n", a+b)
	fmt.Println("字面量 decimal:", decimal.NewFromFloat(0.1+0.2).StringFixed(25))
	fmt.Println("变量   decimal:", decimal.NewFromFloat(a+b).StringFixed(25))
	fmt.Println("变量   decimal String():", decimal.NewFromFloat(a+b).String())
	fmt.Println("string 0.3    :", decimal.RequireFromString("0.3").StringFixed(25))
	fmt.Println("a+b == 0.3 ?", a+b == 0.3)

	fmt.Println("\n=== NewFromFloat 对单个小数 ===")
	fmt.Println("NewFromFloat(0.1) StringFixed(25) =", decimal.NewFromFloat(0.1).StringFixed(25))
	fmt.Println("NewFromFloat(0.1).String()        =", decimal.NewFromFloat(0.1).String())
	fmt.Println("NewFromString(\"0.1\") StringFixed(25) =", decimal.RequireFromString("0.1").StringFixed(25))
	fmt.Println("NewFromFloat(0.1).Equal(NewFromString(0.1)) =", decimal.NewFromFloat(0.1).Equal(decimal.RequireFromString("0.1")))

	fmt.Println("\n=== NewFromFloat 对带误差的计算结果 ===")
	v := 0.1 + 0.2
	fmt.Printf("v = %.20f\n", v)
	dv := decimal.NewFromFloat(v)
	fmt.Println("NewFromFloat(v).String()      =", dv.String())
	fmt.Println("NewFromFloat(v).StringFixed(25)=", dv.StringFixed(25))
	fmt.Println("diff vs 0.3 =", dv.Sub(decimal.RequireFromString("0.3")).String())

	fmt.Println("\n=== 再举一个更明显的 ===")
	p := 0.1
	q := 0.7
	sum := p + q
	fmt.Printf("0.1+0.7 = %.20f\n", sum)
	fmt.Println("NewFromFloat(0.1+0.7 runtime).String() =", decimal.NewFromFloat(sum).String())
	fmt.Println("NewFromFloat(0.1+0.7 runtime).StringFixed(20) =", decimal.NewFromFloat(sum).StringFixed(20))

	fmt.Println("\n=== map key ===")
	m := map[decimal.Decimal]string{}
	k1 := decimal.NewFromInt(1)
	m[k1] = "one"
	fmt.Println("用同一个变量查:", m[k1])
	fmt.Println("用新构造的查:", m[decimal.NewFromInt(1)], "(空 = 查不到)")
	fmt.Println("len(m) =", len(m))
	m[decimal.NewFromInt(1)] = "uno"
	fmt.Println("新构造的键写进去后 len(m) =", len(m), "→ 同一逻辑值占了两个槽位")
	for k, v := range m {
		fmt.Printf("   key=%s val=%s\n", k.String(), v)
	}

	fmt.Println("\n=== == vs Equal ===")
	x := decimal.RequireFromString("1.10")
	y := decimal.RequireFromString("1.1")
	fmt.Println("x == y :", x == y)
	fmt.Println("x.Equal(y):", x.Equal(y))
	fmt.Println("x.Cmp(y):", x.Cmp(y))
	fmt.Println("x.String():", x.String(), " y.String():", y.String())
}
