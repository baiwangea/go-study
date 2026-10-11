# Go Study 多模块学习仓库统一入口
#
# 每个示例目录都是独立的 Go module（见 go.work），因此根目录无法用 `go build ./...`
# 覆盖全部代码，这里统一按模块遍历执行。
#
#   make path          打印学习路线（四个阶段，完整说明见 README.zh.md）
#   make list          列出所有模块
#   make build         编译所有模块（产物统一输出到 bin/）
#   make vet           静态检查所有模块
#   make test          运行所有模块的测试（当前多数模块暂无测试用例）
#   make tidy          对所有模块执行 go mod tidy
#   make fmt           gofmt 全仓库
#   make clean         清理 bin/ 产物
#   make build-linux     用 GoFrame 模块实测交叉编译出 linux/amd64 静态二进制
#   make run M=./04-goframe   运行指定模块（M 为 go.work 中列出的模块目录）
#   make run M=./03-go-concurrency
#
# 关卡式学习（直观学习模式）：一个文件 = 一关 = 完整 main()，直接跑单个文件
#     make levels M=./02-go-pointers    列出该模块的关卡文件（只列带 //go:build level 的）
#     make level  F=11_value_vs_pointer.go    跑指定关卡（M 可省略，会自动找）
#     make level  F=11                         F 支持前缀匹配、可省略 .go
#   也可以直接：cd 02-go-pointers && go run 11_value_vs_pointer.go

# 从 go.work 的 use ( ... ) 块里取出所有模块目录（保留换行，供 shell 按词拆分）。
# 注意：这里必须用 [[:space:]] 而不是 \s —— \s 是 GNU 扩展，
# macOS 自带的 grep 不认，会让 MODULES 变成空串，build/vet/test 全部静默空跑。
MODULES := $(shell sed -n '/^use (/,/^)/p' go.work | grep -E '^[[:space:]]+\./' | tr -d '\t ')
BIN := $(CURDIR)/bin

# 「直观学习模式」下，某些模块的 .go 文件全部带 `//go:build level` 标签
# （一个文件一关、各自带 func main()，靠标签避免 main 重复声明）。
# 这类模块在常规构建下匹配不到任何包，go 会报下面两种信息之一并以退出码 1 结束。
# 这不是错误，只是「本模块没有可编译的命令行程序」，所以三个目标都要放行。
NOPKG := matched no packages|no main packages to build

.PHONY: list build vet test tidy fmt clean path build-linux help levels level

path:
	@echo "📍 Go 学习路线（完整说明见 README.zh.md 的「四个阶段」）"
	@echo
	@sed -n '/^| 阶段 |/,/^$$/p' README.zh.md
	@echo
	@echo "开始今天这一关："
	@echo "  make levels M=./02-go-pointers"
	@echo "  make level  M=./02-go-pointers F=11_value_vs_pointer.go"

help:
	@grep -E '^#   ' Makefile | sed 's/^#   //'

list:
	@for m in $(MODULES); do echo $$m; done

# 产物写入 bin/，避免 go build 把二进制落在各模块目录里污染 git status
#
# 直观学习模式下，模块里可能一个 main 包都没有（全是 //go:build level 的单文件关卡），
# 这时 `go build -o bin/ ./...` 会报 no main packages to build。遇到这种模块，
# 退回 `go build ./...` 把剩下的真实包（helper / internal / demoserver 之类）编译一遍，
# 避免「什么都没编译却报 OK」。
build:
	@mkdir -p $(BIN)
	@for m in $(MODULES); do \
		printf '%-32s' "$$m"; \
		out=$$(go -C $$m build -o $(BIN)/ ./... 2>&1); rc=$$?; \
		if [ $$rc -eq 0 ]; then echo OK; \
		elif printf '%s' "$$out" | grep -q 'no main packages to build'; then \
			chk=$$(go -C $$m build ./... 2>&1); \
			if [ $$? -ne 0 ]; then echo FAIL; printf '%s\n' "$$chk"; exit 1; \
			elif printf '%s' "$$chk" | grep -q 'matched no packages'; then \
				echo 'OK (只有单文件关卡，无常规可编译包)'; \
			else echo 'OK (无命令行程序，已编译其余包)'; fi; \
		elif printf '%s' "$$out" | grep -q 'matched no packages'; then \
			echo 'OK (只有单文件关卡，无常规可编译包)'; \
		else echo FAIL; printf '%s\n' "$$out"; exit 1; fi; \
	done
	@echo "✅ 全部模块编译通过，产物位于 bin/"

clean:
	@rm -rf $(BIN)
	@echo "🧹 已清理 bin/"

vet:
	@for m in $(MODULES); do \
		printf '%-32s' "$$m"; \
		out=$$(go -C $$m vet ./... 2>&1); rc=$$?; \
		if [ $$rc -eq 0 ]; then echo OK; \
		elif printf '%s' "$$out" | grep -qE '$(NOPKG)'; then echo 'OK (本模块没有可检查的包)'; \
		else echo FAIL; printf '%s\n' "$$out"; exit 1; fi; \
	done
	@echo "✅ 全部模块检查通过"

test:
	@for m in $(MODULES); do \
		echo "🧪 $$m"; \
		out=$$(go -C $$m test ./... 2>&1); rc=$$?; \
		if [ $$rc -eq 0 ]; then printf '%s\n' "$$out"; \
		elif printf '%s' "$$out" | grep -qE '$(NOPKG)'; then echo "   (本模块没有可测试的包)"; \
		else printf '%s\n' "$$out"; exit 1; fi; \
	done

tidy:
	@for m in $(MODULES); do \
		echo "📦 $$m"; \
		go -C $$m mod tidy || exit 1; \
	done

fmt:
	@gofmt -l -w $$(find . -name '*.go' -not -path './.idea/*')

# 运行单个模块：make run M=./04-goframe；入口不在模块根目录时再加 P，如 P=./cmd
run:
	@test -n "$(M)" || (echo "用法: make run M=./04-goframe [P=./cmd]" && exit 1)
	@go -C $(M) run $(or $(P),.)

# 列出某个模块的关卡文件（只列带 //go:build level 标签的，即真正的关卡；
# 路径相对模块目录，可以直接喂给 make level）
levels:
	@test -n "$(M)" || (echo "用法: make levels M=./02-go-pointers" && exit 1)
	@test -z "$(F)" || (echo "⚠️  F=$(F) 对 levels 无效（它只负责列清单）。跑关卡请用：make level F=$(F)" && echo)
	@echo "📚 $(M) 的关卡文件："
	@find $(M) -name '*.go' -exec grep -l '^//go:build level' {} + \
		| sort | sed 's|^$(M)/||' | sed 's|^|  |'
	@echo
	@echo "跑某一关：make level F=<上面的文件名>"

# 运行单个关卡文件。三种写法都行：
#   make level F=11_value_vs_pointer.go                       # M 可省略，自动在所有模块里找
#   make level F=11_value_vs_pointer                          # .go 可省略
#   make level F=channels/02_buffered.go                      # 带子目录路径也行
#   make level M=./02-go-pointers F=11                        # 前缀匹配 + 限定模块
# 匹配到多个时会让你写全，不会猜。
level:
	@test -n "$(F)" || (echo "用法: make level F=<关卡文件名或前缀> [M=<模块目录>]" && exit 1)
	@case "$(F)" in \
	  */*) hit=$$(find $(MODULES) -name '*.go' -path "*$(M)*" -path "*$(F)*" \
	         -exec grep -l '^//go:build level' {} + 2>/dev/null);; \
	  *)   hit=$$(find $(MODULES) -name '*.go' -path "*$(M)*" -name '$(F)*' \
	         -exec grep -l '^//go:build level' {} + 2>/dev/null);; \
	esac; \
	n=$$(printf '%s\n' "$$hit" | grep -c .); \
	if [ "$$n" -eq 0 ]; then \
		echo "❌ 没找到关卡 '$(F)'$(if $(M),（限定在 $(M) 内）)"; \
		echo "   列清单：make levels M=<模块目录>，例如 make levels M=./02-go-pointers"; exit 1; \
	elif [ "$$n" -gt 1 ]; then \
		echo "❌ '$(F)' 匹配到 $$n 个关卡，请写全文件名或加 M= 限定模块："; \
		printf '%s\n' "$$hit" | sed 's|^|   |'; exit 1; \
	fi; \
	dir=$$(dirname "$$hit"); file=$$(basename "$$hit"); \
	echo "▶ cd $$dir && go run $$file"; \
	echo; \
	go -C "$$dir" run "$$file"

# 实测交叉编译：关 CGO + 去调试信息，产出可直接上传 VPS 的单文件（L3-16 关卡同款）
build-linux:
	@mkdir -p 04-goframe/runtime/dist
	@cd 04-goframe; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w' -o runtime/dist/bot-linux-amd64 .
	@ls -lh 04-goframe/runtime/dist | tail -1
	@file 04-goframe/runtime/dist/bot-linux-amd64 | cut -d, -f1-3
