package main

import "fmt"

// L1-17 用到的三个类型：故意让 Dog 与 Cat 之间没有任何继承关系。
type Speaker interface{ Speak() string }

// Loggable 是第二个独立契约，对应 PHP 里的 interface Loggable。
type Loggable interface{ Log() string }

type Dog struct{}

func (d Dog) Speak() string { return "Woof" }

type Cat struct{}

func (c Cat) Speak() string { return "Meow" }
func (c Cat) Log() string   { return "cat spoke" }

// NotASpeaker 没有任何方法，用于展示「不满足接口」的情况。
type NotASpeaker struct{}

func (NotASpeaker) String() string { return fmt.Sprint("我不是 Speaker") }
