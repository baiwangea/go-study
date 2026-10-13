//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 04-goframe && go run 16_deploy_build.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
L3-16：交叉编译与部署产物
═══════════════════════════════════════════════════════════════════════

【核心概念】
    把服务变成一个可上传 VPS 的单文件二进制，并理解 CGO 关闭与目标平台的关系。
    注意文件名：本文件原名 16_deploy_linux.go，结果在 macOS 上整个文件被编译器忽略 ——
    因为 Go 把 *_GOOS.go / *_GOARCH.go 后缀当作隐式构建约束。这是个真实踩坑点。

【你会看到】
    ✓ 默认只展示命令与产物规划；设 DEPLOY_DEMO=1 时真实编译出 linux/amd64 二进制并校验 ELF 头

【思考题】（动手改代码验证，别空想）
    1. 为什么 CGO_ENABLED=0 之后产物能在没有 glibc 版本的容器里跑？代价是哪些库用不了？
    2. 配置文件、时区、日志目录这三样，交叉编译时哪些必须随产物一起分发？
    3. 机器人要 7x24 跑：systemd 的 Restart=always 与自研守护进程，哪个更省心？为什么？

【通关标准】
    能独立产出 linux/amd64 静态二进制，并说清部署机上还需要准备什么

【前置】L3-15      【标签】GOOS/GOARCH · CGO_ENABLED=0 · 静态二进制 · systemd
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L3-16 交叉编译与部署产物")
	fmt.Println(strings.Repeat("=", 60) + "\n")
	out := filepath.Join("runtime", "dist", "bot-linux-amd64")

	// ── 示例 1：展示常用交叉编译命令 ─────────────────────────────────────
	fmt.Println("📝 示例 1：常用命令")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o " + out + " .")
	fmt.Println("    或直接用 make build-linux（见根目录 Makefile）")

	// ── 示例 2：默认跳过真实编译（耗时较长） ─────────────────────────────
	if os.Getenv("DEPLOY_DEMO") == "" {
		fmt.Println("  默认跳过真实编译（耗时较长）。想看真结果：DEPLOY_DEMO=1 go run 16_deploy_build.go")
		return
	}

	// ── 示例 3：真实编译并校验 ELF 头 ────────────────────────────────────
	fmt.Println("📌 示例 2：DEPLOY_DEMO 已设置，真实编译 linux/amd64 产物")
	fmt.Println(strings.Repeat("-", 60))
	cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", out, ".")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64")
	if b, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("  编译失败：%v\n%s\n", err, b)
		return
	}
	info, err := os.Stat(out)
	if err != nil {
		fmt.Println("  读不到产物：", err)
		return
	}
	f, _ := os.Open(out)
	magic := make([]byte, 4)
	_, _ = f.Read(magic)
	_ = f.Close()
	fmt.Printf("  产物：%s，%.1f MB，ELF 头校验：%v（说明是 Linux 可执行文件）\n",
		out, float64(info.Size())/1024/1024, binary.BigEndian.Uint32(magic) == 0x7f454c46)

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 文件名不能带 *_linux.go 这类后缀，否则会被当作隐式构建约束而整文件被忽略")
	fmt.Println("  ✅ CGO_ENABLED=0 产出静态二进制，不依赖目标机 glibc")
	fmt.Println("  ✅ 部署机还需要随产物分发配置文件、时区数据、日志目录")
}
