# Go Study 多模块学习仓库统一入口
#
# 每个示例目录都是独立的 Go module（见 go.work），因此根目录无法用 `go build ./...`
# 覆盖全部代码，这里统一按模块遍历执行。
#
#   make path          打印学习路线（阶段表，完整版见 LEARNING_PATH.md）
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
# 关卡式学习：
#   直观学习模式（02-go-pointers 已改造）：一个文件 = 一关 = 完整 main()，直接跑单个文件
#     make levels M=./02-go-pointers                           列出该模块的关卡文件
#     make level  M=./02-go-pointers F=11_value_vs_pointer.go  跑指定关卡
#   关卡制（01 / 03 / 04，待改造）：
#     cd 01-go-fundamentals && go run . list                   看关卡目录
#     cd 01-go-fundamentals && go run . 3                      只跑第 3 关

# 从 go.work 的 use ( ... ) 块里取出所有模块目录（保留换行，供 shell 按词拆分）。
# 注意：这里必须用 [[:space:]] 而不是 \s —— \s 是 GNU 扩展，
# macOS 自带的 grep 不认，会让 MODULES 变成空串，build/vet/test 全部静默空跑。
MODULES := $(shell sed -n '/^use (/,/^)/p' go.work | grep -E '^[[:space:]]+\./' | tr -d '\t ')
BIN := $(CURDIR)/bin

.PHONY: list build vet test tidy fmt clean path build-linux help levels level

path:
	@echo "📍 Go 学习路线（完整说明见 LEARNING_PATH.md）"
	@echo
	@sed -n '/^| \*\*L1/,/^| \*\*L4/p' LEARNING_PATH.md
	@echo
	@echo "开始今天这一关："
	@echo "  cd 01-go-fundamentals && go run . list"

help:
	@grep -E '^#   ' Makefile | sed 's/^#   //'

list:
	@for m in $(MODULES); do echo $$m; done

# 产物写入 bin/，避免 go build 把二进制落在各模块目录里污染 git status
build:
	@mkdir -p $(BIN)
	@for m in $(MODULES); do \
		printf '%-32s' "$$m"; \
		if go -C $$m build -o $(BIN)/ ./... > /dev/null 2>&1; then echo OK; \
		else echo FAIL; go -C $$m build -o $(BIN)/ ./...; exit 1; fi; \
	done
	@echo "✅ 全部模块编译通过，产物位于 bin/"

clean:
	@rm -rf $(BIN)
	@echo "🧹 已清理 bin/"

vet:
	@for m in $(MODULES); do \
		printf '%-32s' "$$m"; \
		if go -C $$m vet ./... > /dev/null 2>&1; then echo OK; \
		else echo FAIL; go -C $$m vet ./...; exit 1; fi; \
	done
	@echo "✅ 全部模块检查通过"

test:
	@for m in $(MODULES); do \
		echo "🧪 $$m"; \
		go -C $$m test ./... || exit 1; \
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

# 列出某个模块的关卡文件（直观学习模式：一个文件 = 一关）
levels:
	@test -n "$(M)" || (echo "用法: make levels M=./02-go-pointers" && exit 1)
	@echo "📚 $(M) 的关卡文件："
	@find $(M) -name '*.go' ! -name 'doc.go' | sort | sed 's|^|  |'
	@echo
	@echo "跑某一关：make level M=$(M) F=<文件名>"

# 运行单个关卡文件：make level M=./02-go-pointers F=11_value_vs_pointer.go
level:
	@test -n "$(M)" -a -n "$(F)" || (echo "用法: make level M=./02-go-pointers F=11_value_vs_pointer.go" && exit 1)
	@go -C $(M) run $(F)

# 实测交叉编译：关 CGO + 去调试信息，产出可直接上传 VPS 的单文件（L3-16 关卡同款）
build-linux:
	@mkdir -p 04-goframe/runtime/dist
	@cd 04-goframe; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w' -o runtime/dist/bot-linux-amd64 .
	@ls -lh 04-goframe/runtime/dist | tail -1
	@file 04-goframe/runtime/dist/bot-linux-amd64 | cut -d, -f1-3
