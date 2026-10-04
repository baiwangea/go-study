package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"go-study/goframe/level"
)

const jwtSecret = "change-me-in-config" // 真实项目从 g.Cfg()/环境变量读取

type jwtClaims struct {
	User string `json:"user"`
	Sym  string `json:"sym"`
	Exp  int64  `json:"exp"`
}

func b64(m map[string]any) string {
	b, _ := json.Marshal(m)
	return base64.RawURLEncoding.EncodeToString(b)
}

func signHS256(payload string) string {
	m := hmac.New(sha256.New, []byte(jwtSecret))
	m.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

// makeToken 手写 HS256 JWT：header.payload.signature
func makeToken(user string, ttl time.Duration) string {
	payload := b64(map[string]any{"user": user, "sym": "HS256", "exp": time.Now().Add(ttl).Unix()})
	head := b64(map[string]any{"alg": "HS256", "typ": "JWT"})
	return head + "." + payload + "." + signHS256(head+"."+payload)
}

func verifyToken(token string) (*jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("token 段数不对")
	}
	if parts[1] == "" || parts[0] == "" {
		return nil, errors.New("空段")
	}
	if signHS256(parts[0]+"."+parts[1]) != parts[2] {
		return nil, errors.New("签名不匹配")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var c jwtClaims
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}
	if time.Now().Unix() > c.Exp {
		return nil, errors.New("token 已过期")
	}
	return &c, nil
}

// authMiddleware 校验 Authorization: Bearer <token>，失败直接 401 中断请求
func authMiddleware(r *ghttp.Request) {
	got := r.Header.Get("Authorization")
	if !strings.HasPrefix(got, "Bearer ") {
		r.Response.WriteStatus(http.StatusUnauthorized, "缺少 Bearer token")
		return
	}
	claims, err := verifyToken(strings.TrimPrefix(got, "Bearer "))
	if err != nil {
		r.Response.WriteStatus(http.StatusUnauthorized, "token 无效："+err.Error())
		return
	}
	r.SetCtxVar("user", claims.User) // 往下游传身份
	r.Middleware.Next()
}

// L3-12：JWT 鉴权（手写 HS256，看清 token 到底是什么）。
func L12() level.Level {
	return level.Level{
		ID:      "L3-12",
		Title:   "手写 JWT 与鉴权中间件",
		Tags:    "HS256 · HMAC · 过期 · 中间件",
		Pre:     "L3-03",
		Goal:    "看清 JWT 的三段结构与校验流程，并把它做成 GoFrame 中间件保护接口",
		Observe: "无 token 401、篡改签名 401、过期 401、合法 token 200 并回显 payload 里的 user",
		Questions: []string{
			"把 secret 换成另一个值再请求，为什么全部 401？这说明服务端重启后 secret 变化会带来什么后果？",
			"exp 用时间戳意味着校验依赖服务器时钟 —— 多台机器时钟漂移会怎样？",
			"JWT 无法主动作废：用户点了「退出登录」该怎么处理？（提示：Redis 黑名单，正好接 L3-09）",
			"生产上该用哪个成熟库替代本关的手写实现？为什么先手写一遍是值得的？",
		},
		Check: "能徒手解释并验证 JWT 三段结构、签名校验与过期判断，并挂成路由中间件",
		Run: func() {
			type loginReq struct {
				User string `json:"user" v:"required#用户名必填"`
			}

			withServer(func(s *ghttp.Server) {
				s.BindHandler("POST:/login", func(r *ghttp.Request) {
					var req *loginReq
					if err := r.Parse(&req); err != nil {
						r.Response.WriteStatus(http.StatusBadRequest, err.Error())
						return
					}
					r.Response.WriteJson(g.Map{"token": makeToken(req.User, time.Hour)})
				})
				s.Group("/api", func(group *ghttp.RouterGroup) {
					group.Middleware(authMiddleware)
					group.GET("/me", func(r *ghttp.Request) {
						r.Response.Writef("hello %s", r.GetCtxVar("user").String())
					})
				})
			}, func(base string) {
				resp, err := http.Post(base+"/login", "application/json", strings.NewReader(`{"user":"trader01"}`))
				if err != nil {
					fmt.Println("  登录失败：", err)
					return
				}
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				var out struct {
					Token string `json:"token"`
				}
				_ = json.Unmarshal(body, &out)
				tok := out.Token
				fmt.Printf("  登录拿到 token：%s…（三段用 . 分隔）\n", firstN(tok, 42))

				call := func(name, value string) {
					req, _ := http.NewRequest(http.MethodGet, base+"/api/me", nil)
					if value != "" {
						req.Header.Set("Authorization", value)
					}
					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						fmt.Println("   请求失败：", err)
						return
					}
					b, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					fmt.Printf("  %-14s → %d %s\n", name, resp.StatusCode, strings.TrimSpace(string(b)))
				}
				call("不带 token", "")
				call("合法 token", "Bearer "+tok)
				call("签名被篡改", "Bearer "+strings.Replace(tok, tok[len(tok)-4:], "0000", 1))
				call("过期 token", "Bearer "+makeToken("trader01", -time.Minute))
			})
		},
	}
}

func firstN(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
