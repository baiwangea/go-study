// 本文件是 T-08 的「故意不完整」的测试。
//
// 它只覆盖了 money.go 的一部分分支 —— 这是刻意的。
// 跑 go tool cover -html=cover.out 会看到哪些行是红的，
// 那就是你这一关要亲手补上的部分。
//
// ⚠ 请先自己看红色标注、想清楚缺哪些边界，再动手补。
//   不要为了让数字变绿而写「调用一次、不断言」的空测试。

package main

import (
	"testing"
)

func TestFormatQty(t *testing.T) {
	tests := []struct {
		name     string
		qty      float64
		stepSize float64
		want     string
		wantErr  bool
	}{
		{"三位小数", 1.5, 0.001, "1.500", false},
		{"整数步长", 3, 1, "3", false},
		{"两位小数", 2.25, 0.01, "2.25", false},
		{"数量小于步长", 0.005, 0.01, "", true},
		{"步长为 0", 1, 0, "", true},

		// ── 下面是没覆盖到的分支，留给你补（思考题 1）─────────────────
		// 提示：stepSize 是负数会怎样？digits 上限 8 那条路径呢？
		//       qty 恰好等于 stepSize 是边界还是错误？
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatQty(tt.qty, tt.stepSize)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("期望报错，实际返回 %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("意外报错: %v", err)
			}
			if got != tt.want {
				t.Errorf("FormatQty(%v, %v) = %q, want %q", tt.qty, tt.stepSize, got, tt.want)
			}
		})
	}
}

func TestFeeTier(t *testing.T) {
	tests := []struct {
		name   string
		volume float64
		want   int
	}{
		{"小额普通用户", 5_000, 10},
		{"十万以下", 99_999, 10},
		{"百万以下", 500_000, 8},

		// ── 留给你补：百万到千万之间、千万以上、以及负数异常值 ─────────
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FeeTier(tt.volume); got != tt.want {
				t.Errorf("FeeTier(%v) = %d, want %d", tt.volume, got, tt.want)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 注意：IsMakerOnly 和 DescribeTier **完全没有测试**。
//
// 跑 go tool cover -func=cover.out 会看到它们分别是 0.0%。
// 这两个函数就是本关的「盲区演示」——
//
// 关键问题不是「怎么把它们补到 100%」，
// 而是「**为什么它们会被漏掉**」：
//   · IsMakerOnly 只有一个表达式，看起来「简单到不用测」
//   · DescribeTier 的多个分支（VIP/高级/标准/基础）只测一个就以为够了
//
// 真实项目里的盲区几乎全是这么产生的：**越简单的函数越没人测，
// 而恰恰是 tier <= 5 这种边界写反了才会出事故。**
// ─────────────────────────────────────────────────────────────────────────────
