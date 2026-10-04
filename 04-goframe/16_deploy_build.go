package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"go-study/goframe/level"
)

// L3-16：交付形态 —— 交叉编译、体积与运行前提。
//
// 注意文件名：本文件原名 16_deploy_linux.go，结果在 macOS 上整个文件被编译器忽略——
// 因为 Go 把 *_GOOS.go / *_GOARCH.go 后缀当作隐式构建约束。这是个真实踩坑点。
func L16() level.Level {
	return level.Level{
		ID:      "L3-16",
		Title:   "交叉编译与部署产物",
		Tags:    "GOOS/GOARCH · CGO_ENABLED=0 · 静态二进制 · systemd",
		Pre:     "L3-15",
		Goal:    "把服务变成一个可上传 VPS 的单文件二进制，并理解 CGO 关闭与目标平台的关系",
		Observe: "默认只展示命令与产物规划；设 DEPLOY_DEMO=1 时真实编译出 linux/amd64 二进制并校验 ELF 头",
		Questions: []string{
			"为什么 CGO_ENABLED=0 之后产物能在没有 glibc 版本的容器里跑？代价是哪些库用不了？",
			"配置文件、时区、日志目录这三样，交叉编译时哪些必须随产物一起分发？",
			"机器人要 7x24 跑：systemd 的 Restart=always 与自研守护进程，哪个更省心？为什么？",
		},
		Check: "能独立产出 linux/amd64 静态二进制，并说清部署机上还需要准备什么",
		Run: func() {
			out := filepath.Join("runtime", "dist", "bot-linux-amd64")
			fmt.Println("  常用命令：")
			fmt.Println("    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o " + out + " .")
			fmt.Println("    或直接用 make build-linux（见根目录 Makefile）")

			if os.Getenv("DEPLOY_DEMO") == "" {
				fmt.Println("  默认跳过真实编译（耗时较长）。想看真结果：DEPLOY_DEMO=1 go run . 16")
				return
			}

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
		},
	}
}
