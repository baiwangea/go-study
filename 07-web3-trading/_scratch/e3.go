//go:build level

package main

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("=== 累加各种费率 ===")
	for _, fee := range []float64{0.04, 0.07, 0.03, 0.0007, 0.0013, 0.0021, 0.1, 0.01, 0.001, 0.0001} {
		var s float64
		for i := 0; i < 10000; i++ {
			s += fee
		}
		exact := decimal.NewFromFloat(fee).Mul(decimal.NewFromInt(10000))
		d := decimal.NewFromFloat(s)
		diff := d.Sub(exact)
		fmt.Printf("  fee=%-8v sum=%.15f exact=%s diff=%s\n", fee, s, exact.String(), diff.String())
	}

	fmt.Println("=== decimal 基本 ===")
	d, _ := decimal.NewFromString("0.1")
	e, _ := decimal.NewFromString("0.2")
	fmt.Println("0.1+0.2 =", d.Add(e).String())
	fmt.Println("NewFromFloat(0.1) =", decimal.NewFromFloat(0.1).String())
	fmt.Println("NewFromFloat(0.1) full =", decimal.NewFromFloat(0.1).StringFixed(20))
	fmt.Println("NewFromString(0.1) full =", d.StringFixed(20))
	fmt.Printf("NewFromFloat(0.3).String()=%q\n", decimal.NewFromFloat(0.3).String())
	fmt.Printf("NewFromFloat(0.1).String()=%q\n", decimal.NewFromFloat(0.1).String())
	fmt.Println("NewFromFloatWithExponent(0.1,-1) =", decimal.NewFromFloatWithExponent(0.1, -1).String())
	fmt.Println("NewFromFloat(1e18) =", decimal.NewFromFloat(1e18).String())
	fmt.Println("NewFromFloat(0.1+0.2) =", decimal.NewFromFloat(0.1+0.2).String())

	fmt.Println("=== Cmp ===")
	x, _ := decimal.NewFromString("1.10")
	y, _ := decimal.NewFromString("1.1")
	fmt.Println("1.10 == 1.1 (==)", x == y, " Cmp:", x.Cmp(y), " Equal:", x.Equal(y))
	fmt.Println("x.String()=", x.String(), "y.String()=", y.String())

	fmt.Println("=== 除法与精度 ===")
	one := decimal.NewFromInt(1)
	three := decimal.NewFromInt(3)
	fmt.Println("1/3 =", one.Div(three).String())
	fmt.Println("1/3 DivRound 8 =", one.DivRound(three, 8).String())
	fmt.Println("DivisionPrecision =", decimal.DivisionPrecision)
	fmt.Println("1/3 float =", 1.0/3.0)

	fmt.Println("=== big.Int 互转 ===")
	bi := one.BigInt()
	fmt.Println("BigInt:", bi.String())
	fromBig := decimal.NewFromBigInt(bi, -8)
	fmt.Println("NewFromBigInt(1, -8) =", fromBig.String())

	fmt.Println("=== 值类型 map key ===")
	m := map[decimal.Decimal]string{}
	m[decimal.NewFromInt(1)] = "one"
	m[decimal.NewFromInt(2)] = "two"
	fmt.Println(m[decimal.NewFromInt(1)], m[decimal.NewFromInt(2)])
	_, ok := m[decimal.NewFromInt(3)]
	fmt.Println("key 3 exists:", ok)

	fmt.Println("=== Round 系列 ===")
	for _, s := range []string{"2.5", "3.5", "1.5", "0.5", "-2.5", "2.4", "2.6", "1.005", "0.125"} {
		v, _ := decimal.NewFromString(s)
		fmt.Printf("  %-6s Round(0)=%-6s RoundBank(0)=%-6s Truncate(0)=%-6s Floor(0)=%-6s Ceil(0)=%-6s\n",
			s, v.Round(0).String(), v.RoundBank(0).String(), v.Truncate(0).String(), v.Floor().String(), v.Ceil().String())
	}

	fmt.Println("=== StringFixed ===")
	v, _ := decimal.NewFromString("1.5")
	fmt.Println("StringFixed(8) =", v.StringFixed(8))
	fmt.Println("StringFixed(2) =", v.StringFixed(2))
	fmt.Println("StringFixed(0) =", v.StringFixed(0))
	fmt.Printf("String()=%q StringFixed(8)=%q\n", v.String(), v.StringFixed(8))
}
