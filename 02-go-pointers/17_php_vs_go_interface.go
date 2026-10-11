//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 02-go-pointers && go run 17_php_vs_go_interface.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-17：PHP 的 implements vs Go 的隐式实现
═══════════════════════════════════════════════════════════════════════

【核心概念】
    PHP/Java：必须显式写 `class Cat extends Animal implements Loggable`，
              类型和契约是绑死的。
    Go      ：没有任何 implements 关键字。只要一个类型的方法集里
              「凑齐了接口要求的方法」，它就自动满足该接口。
              接口描述的是「调用方需要的最小能力」，不是「实现方的身份」。

【你会看到】
    ✓ 同一个 Speaker 变量先后接住 Dog 和 Cat —— 两者毫无继承关系
    ✓ Cat 额外满足第二个接口 Loggable，用独立的小接口表达第二种能力
    ✓ NotASpeaker 不满足 Speaker（真实代码里是编译期报错：missing method Speak）

【思考题】（动手改代码验证，别空想）
    1. Go 里没有 implements，那「谁实现了谁」怎么固定住？
       提示：编译期断言 var _ Speaker = (*Dog)(nil)。
    2. PHP 用 extends Animal 复用 speak()；Go 没有继承，要复用怎么做？
       提示：嵌入结构体 / 嵌入接口，本质是组合。
    3. Cat 多了一个 Log() 方法，还能当 Speaker 用吗？
       这说明 Go 的接口为什么被称作「最小契约」？
    4. 把 NotASpeaker{} 直接传给接受 Speaker 的函数，编译期报什么错？
       本关末尾有提示，动手试一次。

【通关标准】
    能说清「显式声明」与「隐式满足」的差别，
    并会写编译期断言来防止实现关系悄悄跑偏。

【前置】L1-08, L1-13       【标签】interface · 隐式实现 · 继承 vs 组合
═══════════════════════════════════════════════════════════════════════
*/

// ─────────────────────────────────────────────────────────────────────────────
// 对照代码：取自仓库根目录的 demo/index.php 与 demo/main.go（原样引用，不修改那两个文件）。
// ─────────────────────────────────────────────────────────────────────────────

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

// ─────────────────────────────────────────────────────────────────────────────
// 本关用到的类型：故意让 Dog 与 Cat 之间没有任何继承关系。
// ─────────────────────────────────────────────────────────────────────────────

// Speaker 是本关的主接口：只要「会说话」就算满足。
type Speaker interface{ Speak() string }

// Loggable 是第二个独立契约，对应 PHP 里的 interface Loggable。
type Loggable interface{ Log() string }

// Dog 只有 Speak，所以只满足 Speaker。
type Dog struct{}

func (d Dog) Speak() string { return "Woof" }

// Cat 同时有 Speak 和 Log，所以同时满足 Speaker 和 Loggable。
type Cat struct{}

func (c Cat) Speak() string { return "Meow" }
func (c Cat) Log() string   { return "cat spoke" }

// NotASpeaker 没有任何方法，用于展示「不满足接口」的情况。
type NotASpeaker struct{}

func (NotASpeaker) String() string { return fmt.Sprint("我不是 Speaker") }

// Say 只依赖 Speaker 这个行为，不关心传进来的是谁的后代。
func Say(s Speaker) { fmt.Println("    Say() 收到 →", s.Speak()) }

// 编译期断言：如果哪天 Dog 的 Speak 被改名或删掉，这一行会直接编译失败。
// 这是 Go 里替代 implements 的「固定实现关系」的标准做法。
var (
	_ Speaker  = (*Dog)(nil)
	_ Speaker  = (*Cat)(nil)
	_ Loggable = (*Cat)(nil)
)

func main() {
	fmt.Println("\n=== L1-17 PHP implements vs Go 隐式实现 ===")

	// ── 示例 1：PHP 写法（关键行） ───────────────────────────────────────
	fmt.Println("\n【示例 1】PHP 的写法（demo/index.php）")
	for _, line := range strings.Split(phpCode, "\n") {
		if strings.Contains(line, "interface") || strings.Contains(line, "class Cat") || strings.Contains(line, "extends") {
			fmt.Println("    " + strings.TrimSpace(line))
		}
	}

	// ── 示例 2：Go 写法（关键行） ────────────────────────────────────────
	fmt.Println("\n【示例 2】Go 的写法（demo/main.go）")
	for _, line := range strings.Split(goDemoCode, "\n") {
		if strings.Contains(line, "interface") || strings.Contains(line, "func (") || strings.Contains(line, "func Say") {
			fmt.Println("    " + strings.TrimSpace(line))
		}
	}

	// ── 示例 3：同一个接口变量，先后接住两个毫无关系的类型 ───────────────
	fmt.Println("\n【示例 3】Speaker 谁都能接 —— 只要方法齐了")
	var s Speaker = Dog{}
	fmt.Printf("    Speaker 接住 Dog → %s\n", s.Speak())
	s = Cat{}
	fmt.Printf("    Speaker 接住 Cat → %s（没人声明过它们有任何继承关系）\n", s.Speak())
	Say(Dog{})
	Say(Cat{})

	// ── 示例 4：用「多个小接口」表达多种能力 ─────────────────────────────
	fmt.Println("\n【示例 4】能力用独立的小接口表达，不用继承链")
	var l Loggable = Cat{}
	fmt.Printf("    Cat 另满足 Loggable → %s\n", l.Log())
	fmt.Println("    PHP 的 Cat 同时是 Animal 和 Loggable；Go 用两个互不相干的小接口表达")

	// ── 示例 5：反例 —— 不满足接口的类型 ─────────────────────────────────
	fmt.Println("\n【示例 5】不满足接口会怎样")
	// 真实代码里 var x Speaker = NotASpeaker{} 会在编译期直接失败：
	//   cannot use NotASpeaker{} (value of type NotASpeaker) as Speaker value
	//   in variable declaration: NotASpeaker does not implement Speaker (missing method Speak)
	// 所以这里只能先用 any 装住，再用类型断言演示「它没有 Speak」。
	var maybe any = NotASpeaker{}
	if _, ok := maybe.(Speaker); !ok {
		fmt.Println("    类型断言：NotASpeaker 不满足 Speaker")
		fmt.Println("    若写成 var x Speaker = NotASpeaker{} → 编译失败：missing method Speak")
	}

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • PHP/Java：先声明 implements，类型和契约绑死")
	fmt.Println("  • Go：方法齐了就算实现，接口是「调用方需要的最小能力」")
	fmt.Println("  • 复用不要想继承，用嵌入结构体做组合")
	fmt.Println("  • 用 var _ Speaker = (*Dog)(nil) 做编译期断言，锁住实现关系")
}
