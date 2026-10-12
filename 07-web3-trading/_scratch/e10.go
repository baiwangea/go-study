//go:build level

package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Println("=== 运行时 0.1 相关 ===")
	step := 0.1
	fmt.Printf("step*1000 = %.20f  ==100 ? %v\n", step*1000, step*1000 == 100)
	var sum float64
	for i := 0; i < 1000; i++ {
		sum += step
	}
	fmt.Printf("sum(1000) = %.20f\n", sum)
	fmt.Printf("diff      = %.20f\n", sum-100)
	fmt.Printf("diff sci  = %.6e\n", sum-100)

	fmt.Println("\n=== 0.04 x 10000 ===")
	fee := 0.04
	var sf float64
	for i := 0; i < 10000; i++ {
		sf += fee
	}
	fmt.Printf("float64 sum = %.15f\n", sf)
	fmt.Printf("exact       = %.15f\n", fee*10000)
	fmt.Printf("diff        = %.15f  (%e)\n", sf-fee*10000, sf-fee*10000)
	fmt.Println("sf == 400 ?", sf == 400)

	fmt.Println("\n=== Float64bits 分解 ===")
	for _, v := range []float64{0.1, 0.2, 0.3} {
		bits := math.Float64bits(v)
		sign := bits >> 63
		exp := (bits >> 52) & 0x7ff
		frac := bits & ((1 << 52) - 1)
		fmt.Printf("  %.1f bits=%016x sign=%d exp=%d(%d) frac=%013x\n", v, bits, sign, exp, int(exp)-1023, frac)
		fmt.Printf("      %%064b = %064b\n", bits)
	}
	fmt.Println("  0.1+0.2 运行时:")
	b := math.Float64bits(step + 0.2)
	fmt.Printf("      bits=%016x %%064b = %064b\n", b, b)
	c := math.Float64bits(0.3)
	fmt.Printf("  0.3 常量 bits=%016x %%064b = %064b\n", c, c)
	fmt.Println("  差 1 个 ULP ?", b-c)

	fmt.Println("\n=== 0.1 的二进制展开（前 60 位）===")
	fmt.Println("  0.1(十进制) = 0.00011001100110011001100110011001100110011001100110011001100110011... (二进制，1100 无限循环)")
	fmt.Println("  0.2(十进制) = 0.0011001100110011001100110011001100110011001100110011001100110011... (同样循环)")

	fmt.Println("\n=== 大数 ===")
	var big float64 = 1e18
	fmt.Println("big+1 == big :", big+1 == big)
	fmt.Println("ULP(1e18)    :", math.Nextafter(big, math.Inf(1))-big)
	var i53 float64 = 1 << 53
	fmt.Println("2^53 =", int64(i53), " 2^53+1 == 2^53 :", i53+1 == i53)
	var f float64 = 9007199254740993
	fmt.Printf("float64(9007199254740993) = %.0f\n", f)
	fmt.Printf("int64(9007199254740993)   = %d\n", int64(9007199254740993))
}
