// Package notify 的测试：Console / Telegram 都往标准输出打印，所以用管道捕获输出再断言文案；
// Telegram 的发送路径则用 httptest 起**假 TG 服务**，把 BaseURL 指过去 ——
// 这样既能测成功/失败/超时三条路径，又绝不真的连 api.telegram.org。
//
// 断言的是「可观察行为」：发出去的文案、HTTP 请求长什么样、Send 返回什么错误。
// Console.Sent 是唯一有状态的字段，单独断言它。

package notify

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// captureStdout 把 f 期间写到 os.Stdout 的内容收集起来。
// 注意：调用它的测试**不能** t.Parallel —— 它临时改了进程级全局变量。
func captureStdout(t *testing.T, f func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("创建管道失败: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }() // 无论如何都要还原，否则后续测试的输出会丢

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

// Console：每次 Notify 让 Sent +1，并把消息打出来。
func TestConsoleNotify(t *testing.T) {
	c := &Console{}

	out := captureStdout(t, func() {
		c.Notify("BTCUSDT BUY ORD-1 已受理")
		c.Notify("BTCUSDT SELL ORD-2 已受理")
	})

	if c.Sent != 2 {
		t.Errorf("Console.Sent = %d, want 2", c.Sent)
	}
	for _, want := range []string{"通知", "ORD-1", "ORD-2"} {
		if !strings.Contains(out, want) {
			t.Errorf("输出 = %q, 期望包含 %q", out, want)
		}
	}
}

// Telegram：没有 token 时降级为本地打印（不真的发），且文案要能看出是降级、并给出配置方法。
func TestTelegramNotifyWithoutToken(t *testing.T) {
	out := captureStdout(t, func() {
		Telegram{}.Notify("hello")
	})

	if !strings.Contains(out, "降级为本地打印") {
		t.Errorf("无 token 时应降级打印，实际输出 = %q", out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("降级打印也应带上原消息，实际输出 = %q", out)
	}
	// 光说「降级」没用，得告诉用户怎么配起来
	for _, want := range []string{EnvToken, EnvChatID, "export"} {
		if !strings.Contains(out, want) {
			t.Errorf("降级输出应包含配置方法 %q，实际输出 = %q", want, out)
		}
	}
}

// Telegram：有 token 但没 chat id → 同样降级，而不是发一个必然失败的空 chat_id 请求。
func TestTelegramNotifyWithoutChatID(t *testing.T) {
	out := captureStdout(t, func() {
		Telegram{Token: "fake-token"}.Notify("hello")
	})

	if !strings.Contains(out, "降级为本地打印") {
		t.Errorf("缺 chat id 时应降级打印，实际输出 = %q", out)
	}
	if !strings.Contains(out, EnvChatID) {
		t.Errorf("应指出缺的是 %s，实际输出 = %q", EnvChatID, out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("降级打印也应带上原消息，实际输出 = %q", out)
	}
}

// Telegram：token + chat id 齐全时走真实发送分支，BaseURL 指向假服务。
func TestTelegramNotifyWithToken(t *testing.T) {
	var gotPath, gotForm string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotForm = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
	}))
	defer ts.Close()

	tg := Telegram{Token: "fake-token", ChatID: "42", BaseURL: ts.URL}

	out := captureStdout(t, func() {
		tg.Notify("hello")
	})

	if !strings.Contains(out, "Telegram") {
		t.Errorf("有 token 时应走发送分支，实际输出 = %q", out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("发送内容应带上原消息，实际输出 = %q", out)
	}
	// 假服务必须真的收到了对 sendMessage 的 POST，且带上 chat_id / text
	if want := "/botfake-token/sendMessage"; gotPath != want {
		t.Errorf("请求路径 = %q, want %q", gotPath, want)
	}
	for _, want := range []string{"chat_id=42", "text=hello"} {
		if !strings.Contains(gotForm, want) {
			t.Errorf("请求表单 = %q, 期望包含 %q", gotForm, want)
		}
	}
}

// 成功路径：Send 返回 nil，且默认 BaseURL 不会被真的访问（这里显式注入）。
func TestTelegramSendSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":7}}`))
	}))
	defer ts.Close()

	tg := Telegram{Token: "t", ChatID: "1", BaseURL: ts.URL}
	if err := tg.Send("ping"); err != nil {
		t.Fatalf("假服务返回 ok:true，Send 不该报错，实际: %v", err)
	}
}

// 失败路径：HTTP 401 + ok:false → Send 返回错误（而不是 panic），错误里能看出原因。
func TestTelegramSendUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"ok":false,"error_code":401,"description":"Unauthorized"}`))
	}))
	defer ts.Close()

	tg := Telegram{Token: "bad-token", ChatID: "1", BaseURL: ts.URL}

	var err error
	out := captureStdout(t, func() {
		// Notify 必须**不 panic**、只记一行日志
		tg.Notify("hello")
		err = tg.Send("hello")
	})

	if err == nil {
		t.Fatal("401 + ok:false 应返回错误，实际 nil")
	}
	for _, want := range []string{"401", "Unauthorized"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("错误 = %q, 期望包含 %q", err, want)
		}
	}
	// 通知失败不能把管线带走：Notify 只打日志，且日志里能看出「不影响下单」
	if !strings.Contains(out, "发送失败") {
		t.Errorf("Notify 失败时应记一行日志，实际输出 = %q", out)
	}
	if !strings.Contains(out, "hello") || !strings.Contains(out, "Unauthorized") {
		t.Errorf("失败日志应带上原消息与原因，实际输出 = %q", out)
	}
}

// 失败路径：HTTP 200 但 body 里 ok:false —— 「两层成功」只认 body 的 ok 字段。
func TestTelegramSendOKFalse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":false,"description":"chat not found"}`))
	}))
	defer ts.Close()

	tg := Telegram{Token: "t", ChatID: "nope", BaseURL: ts.URL}
	err := tg.Send("hello")
	if err == nil {
		t.Fatal("HTTP 200 但 ok:false 应返回错误，实际 nil")
	}
	if !strings.Contains(err.Error(), "chat not found") {
		t.Errorf("错误 = %q, 期望包含 description", err)
	}
}

// 超时路径：假服务故意 sleep 超过客户端超时 → Send 报错，且必须**在超时附近返回**，不能被拖死。
func TestTelegramSendTimeout(t *testing.T) {
	release := make(chan struct{})
	var once sync.Once
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-release // 卡住，直到测试放行
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer func() {
		once.Do(func() { close(release) }) // 先放行，再关服务，避免 Close 等超时
		ts.Close()
	}()

	tg := Telegram{Token: "t", ChatID: "1", BaseURL: ts.URL, Timeout: 80 * time.Millisecond}

	start := time.Now()
	err := tg.Send("hello")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("服务端不返回时应超时报错，实际 nil")
	}
	if elapsed > 2*time.Second {
		t.Errorf("耗时 %v，说明没有按 Timeout 生效（应约 80ms 返回）", elapsed)
	}
}

// 缺配置时 Send 直接返回明确错误，而不是发一个必然失败的请求。
func TestTelegramSendMissingConfig(t *testing.T) {
	if err := (Telegram{}).Send("x"); !errors.Is(err, ErrNoToken) {
		t.Errorf("缺 token 时 Send 错误 = %v, want ErrNoToken", err)
	}
	if err := (Telegram{Token: "t"}).Send("x"); !errors.Is(err, ErrNoChatID) {
		t.Errorf("缺 chat id 时 Send 错误 = %v, want ErrNoChatID", err)
	}
}

// 环境变量：新名 TG_BOT_TOKEN 优先，未设置时回落到旧的 TG_TOKEN。
func TestTokenFromEnvCompatibility(t *testing.T) {
	tests := []struct {
		name string
		new  string
		old  string
		want string
	}{
		{name: "都没设", new: "", old: "", want: ""},
		{name: "只设新名", new: "new-token", old: "", want: "new-token"},
		{name: "只设旧名（兼容）", new: "", old: "old-token", want: "old-token"},
		{name: "两个都设时新名优先", new: "new-token", old: "old-token", want: "new-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(EnvToken, tt.new)
			t.Setenv(EnvTokenOld, tt.old)

			if got := TokenFromEnv(); got != tt.want {
				t.Errorf("TokenFromEnv() = %q, want %q", got, tt.want)
			}
		})
	}
}

// NewTelegramFromEnv 把两个环境变量装进结构体，ChatID 也要读到。
func TestNewTelegramFromEnv(t *testing.T) {
	t.Setenv(EnvToken, "tok")
	t.Setenv(EnvTokenOld, "")
	t.Setenv(EnvChatID, "12345")

	tg := NewTelegramFromEnv()
	if tg.Token != "tok" || tg.ChatID != "12345" {
		t.Errorf("NewTelegramFromEnv() = {Token:%q ChatID:%q}, want {tok 12345}", tg.Token, tg.ChatID)
	}
	if tg.BaseURL != "" || tg.Timeout != 0 {
		t.Errorf("BaseURL/Timeout 应留空走默认值，实际 = {%q %v}", tg.BaseURL, tg.Timeout)
	}
	// 默认值通过内部方法兑现，避免 main 里再写一遍常量
	if tg.baseURL() != DefaultBaseURL || tg.timeout() != DefaultTimeout {
		t.Errorf("默认 BaseURL/Timeout = {%q %v}, want {%q %v}",
			tg.baseURL(), tg.timeout(), DefaultBaseURL, DefaultTimeout)
	}
}

// 编译期契约：Telegram 必须是 Notifier（main 只依赖接口）。
var _ Notifier = Telegram{}
var _ Notifier = (*Console)(nil)

// 顺带钉住：SendContext 也能用调用方的 ctx 取消。
func TestTelegramSendContextCanceled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tg := Telegram{Token: "t", ChatID: "1", BaseURL: ts.URL}
	if err := tg.SendContext(ctx, "hello"); err == nil {
		t.Error("ctx 已取消时应返回错误，实际 nil")
	}
}

// 注入自定义 Client（覆盖 client() 的「外部传了客户端」分支）+ 响应不是 JSON：
// 错误要说明「响应无法解析」，且超长响应必须被截断，不能把整页 HTML 打进终端。
func TestTelegramSendUnparseableBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<html>" + strings.Repeat("x", 300) + "</html>"))
	}))
	defer ts.Close()

	tg := Telegram{
		Token: "t", ChatID: "1", BaseURL: ts.URL,
		Client: &http.Client{Timeout: time.Second}, // 显式注入
	}

	err := tg.Send("hello")
	if err == nil {
		t.Fatal("响应不是 JSON 时应返回错误，实际 nil")
	}
	if !strings.Contains(err.Error(), "响应无法解析") {
		t.Errorf("错误 = %q, 期望说明响应无法解析", err)
	}
	if !strings.Contains(err.Error(), "…") {
		t.Errorf("超长响应应被截断（带省略号），实际 = %q", err)
	}
	if len(err.Error()) > 250 {
		t.Errorf("错误信息过长（%d 字符），说明没截断", len(err.Error()))
	}
}
