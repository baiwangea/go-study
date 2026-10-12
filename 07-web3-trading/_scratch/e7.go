//go:build level

package main

import (
	"fmt"
	"math"

	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("=== Round(2) 精度案例 ===")
	for _, s := range []string{"1.005", "1.015", "2.675", "0.125", "1.045"} {
		d := decimal.RequireFromString(s)
		f, _ := d.Float64()
		fmt.Printf("  %-6s decimal Round(2)=%-8s RoundBank(2)=%-8s | float64 math.Round(x*100)/100=%.10f\n",
			s, d.Round(2).String(), d.RoundBank(2).String(), math.Round(f*100)/100)
	}

	fmt.Println("\n=== stepSize 对齐 decimal vs float64 ===")
	price := decimal.RequireFromString("7")
	budget := decimal.RequireFromString("0.7")
	step := decimal.RequireFromString("0.1")
	qty := budget.Div(price)
	fmt.Println("decimal qty =", qty.String())
	aligned := qty.Div(step).Floor().Mul(step)
	fmt.Println("decimal aligned =", aligned.String())

	fmt.Println("\n=== 10000 笔手续费 ===")
	feeS := "0.04"
	feeD := decimal.RequireFromString(feeS)
	feeF, _ := feeD.Float64()
	var sumF float64
	for i := 0; i < 10000; i++ {
		sumF += feeF
	}
	sumD := feeD.Mul(decimal.NewFromInt(10000))
	fmt.Printf("float64 累加 = %.15f\n", sumF)
	fmt.Printf("decimal      = %s\n", sumD.String())
	fmt.Printf("float64 差   = %s\n", decimal.NewFromFloat(sumF).Sub(sumD).String())
	fmt.Println("sumF == 400 ?", sumF == 400)

	fmt.Println("\n=== 分摊 400 到 3 个账户 ===")
	total := 400.0
	each := total / 3
	fmt.Printf("float64 each = %.17g\n", each)
	fmt.Printf("float64 3*each = %.17g  == 400 ? %v\n", each*3, each*3 == 400)
	dTotal := decimal.NewFromInt(400)
	dEach := dTotal.DivRound(decimal.NewFromInt(3), 8)
	fmt.Println("decimal each(DivRound 8) =", dEach.String())
	rem := dTotal.Sub(dEach.Mul(decimal.NewFromInt(3)))
	fmt.Println("decimal 3*each =", dEach.Mul(decimal.NewFromInt(3)).String(), " 余", rem.String())

	fmt.Println("\n=== Round 与 StringFixed 组合 ===")
	v := decimal.RequireFromString("0.123456789")
	fmt.Printf("Round(8)=%s StringFixed(8)=%s\n", v.Round(8).String(), v.Round(8).StringFixed(8))
	fmt.Printf("Truncate(4)=%s\n", v.Truncate(4).String())
	fmt.Printf("Floor=%s Ceil=%s\n", v.Floor().String(), v.Ceil().String())

	fmt.Println("\n=== 负数舍入 ===")
	n := decimal.RequireFromString("-0.125")
	fmt.Printf("-0.125 Round(2)=%s RoundBank(2)=%s Truncate(2)=%s Floor=%s Ceil=%s\n",
		n.Round(2).String(), n.RoundBank(2).String(), n.Truncate(2).String(), n.Floor().String(), n.Ceil().String())

	fmt.Println("\n=== 精度位 ===")
	x := decimal.RequireFromString("1.500")
	fmt.Println("x.String()=", x.String(), " Exponent=", x.Exponent(), " Coefficient=", x.Coefficient().String())
}
