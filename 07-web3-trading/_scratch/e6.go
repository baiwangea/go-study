//go:build level

package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Println("=== 大数丢精度：常量 vs 变量 ===")
	fmt.Println("常量  1e18+1 == 1e18 :", 1e18+1 == 1e18, "(常量是任意精度，算得准)")
	var big float64 = 1e18
	fmt.Println("变量  big+1 == big   :", big+1 == big, "(运行时 float64 加法，1 被吃掉)")
	fmt.Printf("变量  big+1 = %.0f\n", big+1)
	fmt.Printf("变量  big+2 = %.0f\n", big+2)
	fmt.Printf("变量  big+64 = %.0f\n", big+64)
	fmt.Printf("变量  big+127 = %.0f\n", big+127)
	fmt.Printf("变量  big+128 = %.0f\n", big+128)
	fmt.Println("ULP(1e18) =", math.Nextafter(big, math.Inf(1))-big)

	fmt.Println("\n=== wei 场景 ===")
	var wei float64 = 1e18 // 1 ETH
	fmt.Printf("1 ETH  = %.0f wei\n", wei)
	fmt.Printf("+1 wei = %.0f wei（等于没加）\n", wei+1)
	var dust float64 = 1e15 // 0.001 ETH
	fmt.Printf("0.001 ETH = %.0f wei\n", dust)
	fmt.Printf("1 ETH - 0.001 ETH = %.0f wei\n", wei-dust)
	fmt.Printf("(1e18 - 1e15) 期望 = 999000000000000000\n")

	fmt.Println("\n=== 余额累加场景 ===")
	var bal float64
	steps := []float64{1e15, 1e15, 1e15, 1e15, 1e15}
	for _, s := range steps {
		bal += s
	}
	fmt.Printf("累加 5 次 0.001 ETH = %.0f wei = %.18f ETH\n", bal, bal/1e18)
	fmt.Println("bal == 5e15 ?", bal == 5e15)

	fmt.Println("\n=== stepSize 对齐（真实拒单场景）===")
	// 场景：价格 7 USDT，预算 0.7 USDT，stepSize 0.1
	price, budget, step := 7.0, 0.7, 0.1
	qty := budget / price
	fmt.Printf("qty = %.20f\n", qty)
	aligned := math.Floor(qty/step) * step
	fmt.Printf("aligned = %.20f\n", aligned)
	fmt.Println("aligned <= 0 ?", aligned <= 0)

	// 更真实：用 math.Round 修正的常见「土办法」
	aligned2 := math.Floor(qty/step+1e-9) * step
	fmt.Printf("加 1e-9 容差后 aligned2 = %.20f\n", aligned2)

	fmt.Println("\n=== 另一种拒单：累加成交后对齐 ===")
	var filled float64
	for i := 0; i < 3; i++ {
		filled += 0.1
	}
	fmt.Printf("3 笔 0.1 累加 = %.20f\n", filled)
	al := math.Floor(filled/0.1) * 0.1
	fmt.Printf("对齐到 stepSize 0.1 = %.20f\n", al)
	fmt.Println("al == 0.3 ?", al == 0.3)

	fmt.Println("\n=== 0.3/0.1 ===")
	fmt.Printf("0.3/0.1 = %.20f  floor=%.0f\n", 0.3/0.1, math.Floor(0.3/0.1))

	fmt.Println("\n=== 除法本身 ===")
	var a3, a1 float64 = 0.3, 0.1
	fmt.Printf("变量 0.3/0.1 = %.20f → floor=%v\n", a3/a1, math.Floor(a3/a1))
}
