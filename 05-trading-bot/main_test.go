// package main 的测试：只覆盖 main.go 里可单独测试的纯逻辑 envDecimal 与 runBacktest。
//
// main() 本身把 flag、行情源、存储、协程全串在一起，属于「装配代码」，
// 适合端到端跑（go run -race .），不适合塞进单测。但 envDecimal 决定了
// BOT_MAX_NOTIONAL 这类**金额**配置怎么解析 —— 金额从 float64 换成 decimal 后
// 它最容易被顺手改坏，所以单独钉住三条分支：未设置 / 合法数字 / 非法数字。
//
// 断言用 String() 而不是 Equal()：String() 能同时验证「数值对」和「表示没被浮点洗过」，
// 如果实现偷偷走了一遍 float64（decimal.NewFromFloat），长小数会立刻露馅。
//
// 回测模式的装配（runBacktest）也在这里测：它必须在**零网络、零密钥**下跑通，
// 且入参有问题时返回 error 而不是 panic。

package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestEnvDecimal(t *testing.T) {
	tests := []struct {
		name string
		set  bool
		val  string
		def  string
		want string
	}{
		{name: "未设置用默认值", set: false, def: "500", want: "500"},
		{name: "空字符串用默认值", set: true, val: "", def: "500", want: "500"},
		{name: "合法数字被解析", set: true, val: "123.5", def: "500", want: "123.5"},
		{name: "非法数字退回默认值", set: true, val: "一万", def: "500", want: "500"},
		{name: "负数也能解析", set: true, val: "-3", def: "0", want: "-3"},
		{name: "高精度小数不丢位", set: true, val: "0.30000000000000004", def: "0", want: "0.30000000000000004"},
	}

	const key = "BOT_TEST_ENV_DECIMAL"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				t.Setenv(key, tt.val)
			} else {
				// t.Setenv 会自动还原；显式置空即可模拟「未设置」（os.Getenv 对两者都返回 ""）
				t.Setenv(key, "")
			}

			got := envDecimal(key, tt.def)
			if got.String() != tt.want {
				t.Errorf("envDecimal(%q, %q) = %s, want %s", key, tt.def, got, tt.want)
			}
		})
	}
}

// captureStdout 把 f 期间写到 os.Stdout 的内容收集起来，避免回测报告污染测试输出。
// 调用它的测试不能 t.Parallel —— 它临时改了进程级全局变量。
func captureStdout(t *testing.T, f func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("创建管道失败: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	f()

	if err := w.Close(); err != nil {
		t.Fatalf("关闭管道写端失败: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("读取管道失败: %v", err)
	}
	_ = r.Close()
	return string(out)
}

// backtestOpts 的默认值，测试里按需改。
func defaultBacktestOpts() backtestOpts {
	return backtestOpts{
		Pattern: "range", Bars: 120,
		FeeBP: "5", SlippageBP: "1",
		Symbol: "BTCUSDT", Qty: decimal.RequireFromString("0.002"),
	}
}

// 内置形态：零网络零密钥跑通并打印报告。
func TestRunBacktestBuiltin(t *testing.T) {
	var out string
	err := func() error {
		var e error
		out = captureStdout(t, func() { e = runBacktest(defaultBacktestOpts()) })
		return e
	}()
	if err != nil {
		t.Fatalf("回测应跑通，实际: %v", err)
	}
	for _, want := range []string{"开始回测", "回测报告", "形态：震荡", "净利", "最大回撤"} {
		if !strings.Contains(out, want) {
			t.Errorf("报告缺少 %q：\n%s", want, out)
		}
	}
}

// 外部 CSV：正常路径能读进来并跑完。
func TestRunBacktestCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quotes.csv")
	if err := os.WriteFile(path, []byte("timestamp,price\n1704067200,100\n1704067260,101\n"), 0o600); err != nil {
		t.Fatalf("写临时 CSV 失败: %v", err)
	}

	opts := defaultBacktestOpts()
	opts.CSVPath = path

	var out string
	err := func() error {
		var e error
		out = captureStdout(t, func() { e = runBacktest(opts) })
		return e
	}()
	if err != nil {
		t.Fatalf("CSV 回测应跑通，实际: %v", err)
	}
	if !strings.Contains(out, "形态：外部 CSV") || !strings.Contains(out, "行情 2 笔") {
		t.Errorf("CSV 回测报告不对：\n%s", out)
	}
}

// CSV 文件不存在：必须明确报错，而不是 panic。
func TestRunBacktestCSVMissing(t *testing.T) {
	opts := defaultBacktestOpts()
	opts.CSVPath = filepath.Join(t.TempDir(), "nonexistent.csv")

	err := runBacktest(opts)
	if err == nil {
		t.Fatal("CSV 不存在时应返回 error，实际 nil")
	}
	if !strings.Contains(err.Error(), "打开行情 CSV 失败") {
		t.Errorf("错误 = %q, 期望说明是 CSV 打不开", err)
	}
}

// 入参非法时返回 error（由 main 打印并以非零码退出），全部路径都不 panic。
func TestRunBacktestBadOptions(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*backtestOpts)
		wantSub string
	}{
		{"未知形态", func(o *backtestOpts) { o.Pattern = "sideways" }, "未知形态"},
		{"手续费非数字", func(o *backtestOpts) { o.FeeBP = "五个点" }, "手续费率"},
		{"滑点非数字", func(o *backtestOpts) { o.SlippageBP = "x" }, "滑点"},
		{"笔数为零", func(o *backtestOpts) { o.Bars = 0 }, "必须为正"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultBacktestOpts()
			tt.mutate(&opts)

			var err error
			_ = captureStdout(t, func() { err = runBacktest(opts) })
			if err == nil {
				t.Fatal("期望返回 error，实际 nil")
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("错误 = %q, 期望包含 %q", err, tt.wantSub)
			}
		})
	}
}
