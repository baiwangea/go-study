//go:build level

package main

import (
	"fmt"
	"math"
	"strconv"
)

func main() {
	fmt.Println("--- 0.1+0.2 ---")
	a, b := 0.1, 0.2
	fmt.Printf("%.20f\n", a+b)
	fmt.Printf("%v\n", a+b)
	fmt.Println(a+b == 0.3)
	fmt.Printf("%x\n", math.Float64bits(a))
	fmt.Printf("%x\n", math.Float64bits(b))
	fmt.Printf("%x\n", math.Float64bits(a+b))
	fmt.Printf("%x\n", math.Float64bits(0.3))
	fmt.Printf("%b\n", math.Float64bits(0.1))

	fmt.Println("--- 累加 ---")
	var sum float64
	for i := 0; i < 1000; i++ {
		sum += 0.1
	}
	fmt.Printf("累加1000次: %.20f  ==100? %v\n", sum, sum == 100)
	fmt.Printf("1000*0.1 : %.20f\n", 1000*0.1)
	fmt.Printf("差: %.20f\n", sum-1000*0.1)

	fmt.Println("--- 大数 ---")
	fmt.Println(1e18+1 == 1e18)
	fmt.Printf("1e18   = %.0f\n", 1e18)
	fmt.Printf("1e18+1 = %.0f\n", 1e18+1)
	fmt.Printf("float64bits(1e18)=%x\n", math.Float64bits(1e18))
	fmt.Printf("float64bits(1e18+1)=%x\n", math.Float64bits(1e18+1))
	fmt.Printf("float64bits(1e18+2)=%x\n", math.Float64bits(1e18+2))

	fmt.Println("--- 最小可分辨 ---")
	// ULP at 1e18
	fmt.Printf("Nextafter(1e18, inf)-1e18 = %.0f\n", math.Nextafter(1e18, math.Inf(1))-1e18)
	fmt.Printf("Nextafter(1, inf)-1 = %v\n", math.Nextafter(1, math.Inf(1))-1)

	fmt.Println("--- wei 精度 ---")
	fmt.Println("2^53 =", 1<<53)
	fmt.Printf("float64 能精确表示的整数上限 = %d\n", 1<<53)
	// 1 wei of ETH as float
	fmt.Printf("1 wei as float64 = %v\n", 1.0/1e18)
	fmt.Printf("0.1 ETH in wei via float: %.0f\n", 0.1*1e18)
	fmt.Printf("0.3 ETH in wei via float: %.0f\n", 0.3*1e18)
	fmt.Printf("0.1+0.2 ETH in wei via float: %.0f\n", (0.1+0.2)*1e18)
	fmt.Println("0.3*1e18 == 3e17 ?", 0.3*1e18 == 3e17)

	fmt.Println("--- strconv 打印 ---")
	fmt.Println(strconv.FormatFloat(0.1+0.2, 'f', -1, 64))
	fmt.Println(strconv.FormatFloat(0.1+0.2, 'f', 17, 64))
	fmt.Println(strconv.FormatFloat(0.1+0.2, 'f', 20, 64))
}
