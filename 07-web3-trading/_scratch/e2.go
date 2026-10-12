//go:build level

package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Println("=== 0.3/0.1 ===")
	fmt.Printf("%.20f\n", 0.3/0.1)
	fmt.Println(math.Floor(0.3/0.1) * 0.1)
	fmt.Printf("%.20f\n", math.Floor(0.3/0.1)*0.1)

	fmt.Println("=== 候选：算出来 0.09999999999999999 ===")
	cands := [][2]float64{{0.11, 1.1}, {3.33, 33.3}, {0.1, 1.0}, {1.0, 10.0}, {0.33, 3.3}, {0.5, 5.0}, {0.7, 7.0}, {0.3, 3.0}, {0.6, 6.0}, {0.9, 9.0}, {2.1, 21.0}, {4.2, 42.0}}
	for _, c := range cands {
		q := c[0] / c[1]
		al := math.Floor(q/0.1) * 0.1
		fmt.Printf("  %v/%v = %.20f  floor→%.20f\n", c[0], c[1], q, al)
	}

	fmt.Println("=== 1.1*0.1 等 ===")
	fmt.Printf("1.1*0.1 = %.20f\n", 1.1*0.1)
	fmt.Printf("0.11/1.1 = %.20f\n", 0.11/1.1)

	fmt.Println("=== 累加 0.0001 x 10000 ===")
	var s float64
	for i := 0; i < 10000; i++ {
		s += 0.0001
	}
	fmt.Printf("sum = %.20f\n", s)
	fmt.Printf("10000*0.0001 = %.20f\n", 10000*0.0001)
	fmt.Printf("diff = %.20e\n", s-10000*0.0001)
	fmt.Printf("diff(USDT) = %.10f\n", s-10000*0.0001)

	fmt.Println("=== 累加 0.01 x 10000 ===")
	var s2 float64
	for i := 0; i < 10000; i++ {
		s2 += 0.01
	}
	fmt.Printf("sum = %.20f\n", s2)
	fmt.Printf("diff = %.20f\n", s2-100.0)

	fmt.Println("=== 累加 0.1 x 10000 ===")
	var s3 float64
	for i := 0; i < 10000; i++ {
		s3 += 0.1
	}
	fmt.Printf("sum = %.20f\n", s3)
	fmt.Printf("diff = %.20e\n", s3-1000.0)

	fmt.Println("=== 累加 0.0001 x 1000 / 100 ===")
	for _, n := range []int{10, 100, 1000, 10000} {
		var s float64
		for i := 0; i < n; i++ {
			s += 0.0001
		}
		fmt.Printf("  n=%d sum=%.20f diff=%.3e\n", n, s, s-float64(n)*0.0001)
	}
}
