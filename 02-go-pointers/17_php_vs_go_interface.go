package main

import (
	"fmt"
	"strings"

	"go-study/go-pointers/level"
)

// 下面两段对照代码取自仓库根目录的 demo/index.php 与 demo/main.go（原样引用，不修改那两个文件）。
const phpCode = `<?php
class Animal {
    public function speak(): string { return "..."; }
}
class Dog extends Animal {
    public function speak(): string { return "Woof"; }
}
interface Loggable {
    public function log(): string;
}
class Cat extends Animal implements Loggable {   // ← 必须显式声明 implements
    public function speak(): string { return "Meow"; }
    public function log(): string   { return "cat spoke"; }
}
$model = new Cat();
var_dump($model->speak(), $model->log());`

const goDemoCode = `package demo

type Speaker interface{ Speak() string }

type Dog struct{}
func (d Dog) Speak() string { return "Woof" }     // 没有任何 "implements" 声明

type cat struct{}                                 // 非导出类型也能实现导出接口
func (c cat) Speak() string { return "Meow" }

func Say(s Speaker) { fmt.Println(s.Speak()) }    // 只依赖行为，不依赖继承链`

// L17 关卡：PHP 的 implements/extends vs Go 的隐式实现。
func L17() level.Level {
	return level.Level{
		ID:      "L1-17",
		Title:   "PHP implements vs Go 隐式实现",
		Tags:    "interface · 继承 vs 组合 · 最小契约",
		Pre:     "L1-08, L1-13",
		Goal:    "把「先声明 implements 再写方法」的习惯，换成「方法齐了就算实现」的 Go 思维",
		Observe: "两个毫无关系的类型被同一个 Speaker 变量先后接走；Cat 的第二个能力用第二个接口表达",
		Questions: []string{
			"Go 里没有 implements，那「谁实现了谁」怎么固定住？（提示：var _ Speaker = (*Dog)(nil) 编译期断言，见 L1-09）",
			"PHP 用 extends Animal 复用 speak()；Go 没有继承，要复用怎么做？（提示：嵌入结构体/嵌入接口，是组合）",
			"Cat 多了一个 Log() 方法，还能当 Speaker 用吗？这说明 Go 的接口为什么是「最小契约」？",
			"如果把 NotASpeaker{} 传给 Say()，编译期报什么错？本关末尾有提示，动手试一次",
		},
		Check: "能说清显式声明与隐式满足的差别，并会写编译期断言防止实现关系跑偏",
		Run: func() {
			fmt.Println("  【PHP 写法】demo/index.php：")
			for _, l := range strings.Split(phpCode, "\n") {
				if strings.Contains(l, "interface") || strings.Contains(l, "class Cat") || strings.Contains(l, "extends") {
					fmt.Println("    " + strings.TrimSpace(l))
				}
			}
			fmt.Println("  【Go 写法】demo/main.go：")
			for _, l := range strings.Split(goDemoCode, "\n") {
				if strings.Contains(l, "interface") || strings.Contains(l, "func (") || strings.Contains(l, "func Say") {
					fmt.Println("    " + strings.TrimSpace(l))
				}
			}

			// 关键演示：Dog 与 Cat 之间没有任何关系，只是都恰好有 Speak() 方法。
			var s Speaker = Dog{}
			fmt.Printf("\n    Speaker 接住 Dog → %s\n", s.Speak())
			s = Cat{}
			fmt.Printf("    Speaker 接住 Cat → %s（没人声明过它们有任何继承关系）\n", s.Speak())

			// PHP 的 Cat 同时是 Animal 和 Loggable；Go 用两个独立的小接口来表达。
			var l Loggable = Cat{}
			fmt.Printf("    Cat 另满足 Loggable → %s\n", l.Log())

			// 反例：不满足接口的类型。真实代码里 var x Speaker = NotASpeaker{} 会在编译期直接失败，
			// 所以这里只能先用 any 装住，再用类型断言演示「它没有 Speak」。
			var maybe any = NotASpeaker{}
			if _, ok := maybe.(Speaker); !ok {
				fmt.Println("    类型断言：NotASpeaker 不满足 Speaker（若写成 var x Speaker = NotASpeaker{} 会编译失败：missing method Speak）")
			}
			fmt.Println("\n    结论：接口描述的是「调用方需要的最小能力」，不是「实现方的身份」。")
		},
	}
}
