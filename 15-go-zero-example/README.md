# go-zero 微服务示例

go-zero 是带服务治理的 Go 微服务框架，核心工作流是**先写 DSL/Proto，再用 `goctl` 生成代码骨架**，
业务逻辑只写在 `logic` 层。本模块包含三个独立 Go module：

| 目录 | 类型 | 默认地址 | 说明 |
| :-- | :-- | :-- | :-- |
| `greet/` | API 服务 | `0.0.0.0:8888` | 最小入门：`greet.api` → HTTP 服务 |
| `user/` | RPC 服务 | `0.0.0.0:8080` | `user.proto` → gRPC 服务，注册到 etcd（`user.rpc`） |
| `user-api/` | API 服务 | `0.0.0.0:8889` | 对外 HTTP 接口，并演示调用 `user.rpc` |

> `user-api` 的端口特意与 `greet-api` 错开（8889 / 8888），两个 API 可以同时启动。

## 目录结构（以 user-api 为例）

```
user-api/
├── user.api                       # API DSL：类型定义 + 路由声明
├── user.go                        # 程序入口：加载配置、注册路由、启动 HTTP 服务
├── etc/user-api.yaml              # 运行配置（端口 + 下游 RPC 客户端配置）
└── internal/
    ├── config/config.go           # Config 结构体（rest.RestConf + UserRpc）
    ├── handler/                   # HTTP 处理层：解析请求 → 调 logic → 返回（goctl 生成）
    │   └── routes.go              # 路由注册表
    ├── logic/                     # 业务逻辑层（唯一推荐手工编码的地方）
    ├── svc/servicecontext.go      # 依赖注入容器：数据库、Redis、RPC 客户端都挂在这里
    └── types/types.go             # 请求/响应结构体（由 .api 生成，勿手改）
```

RPC 服务 `user/` 结构类似：`user.proto` → `user/`（pb 代码）+ `userclient/`（客户端封装），
`internal/{server,logic,svc,config}` + `etc/user.yaml`。

## 快速开始

三个模块相互独立，在各自目录运行（根目录 `make run M=./15-go-zero-example/user-api` 亦可）：

```sh
# 1. 最小 API
cd greet && go run . -f etc/greet-api.yaml
curl 'http://127.0.0.1:8888/from/you'

# 2. RPC 服务（etc/user.yaml 配置了 etcd，需要先启动本地 etcd）
cd user && go run . -f etc/user.yaml

# 3. API 网关：HTTP 接口 + 调用上面的 RPC
cd user-api && go run . -f etc/user-api.yaml
```

验证接口：

```sh
curl -X POST http://127.0.0.1:8889/user/login \
     -H 'Content-Type: application/json' \
     -d '{"username":"admin","password":"secret"}'
# {"token":"mock-jwt-token-for-admin"}

curl http://127.0.0.1:8889/user/7
# {"id":7,"username":"alice","email":"alice@example.com"}

# API -> gRPC -> user.rpc 的完整链路
curl 'http://127.0.0.1:8889/user/rpc/ping?message=hello-go-zero'
# {"pong":"pong: hello-go-zero"}
```

## API 调用 RPC 是怎么串起来的

1. `user-api/etc/user-api.yaml` 里配置下游：

   ```yaml
   UserRpc:
     Endpoints: [127.0.0.1:8080]   # 直连模式，本地不需要 etcd
     NonBlock: true                # 下游未就绪时启动也不退出
     Timeout: 3000
   ```

2. `internal/config/config.go` 增加 `UserRpc zrpc.RpcClientConf` 字段；
3. `internal/svc/servicecontext.go` 构造客户端并注入：
   `UserRpc: userclient.NewUser(zrpc.MustNewClient(c.UserRpc))`；
4. `internal/logic/pingrpclogic.go` 中直接调用：
   `l.svcCtx.UserRpc.Ping(l.ctx, &user.Request{Ping: req.Message})`。

生产环境把 `Endpoints` 换成服务发现：`Etcd: { Hosts: [127.0.0.1:2379], Key: user.rpc }`。

`user-api` 依赖同仓库的 `user` 模块，靠 `go.work` + `go.mod` 中的
`replace go-study/go-zero-example/user => ../user` 解析，无需发布到远端。

## 用 goctl 改接口（推荐工作流）

```sh
# 修改 user.api 后重新生成：只新增/覆盖生成文件，已存在的自定义文件会被跳过
cd user-api && goctl api go -api user.api -dir .

# 修改 user.proto 后重新生成 RPC 代码
cd user && goctl rpc protoc user.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

goctl 的行为：`routes.go`、`types.go`、`handler` 壳会被重新生成；`logic`、`svc`、`config`
中已存在的文件打印 `exists, ignored generation` 并被保留，所以业务代码可以安全手写。

## 学习建议：下一步可以补的点

- `GetUserInfo` 目前返回硬编码数据 → 接入 MySQL/GORM，参考 `11-gin-framework-example`
- `Login` 返回 mock token → 结合 `13-go-jwt-example` 签发真实 JWT，并加 go-zero `JwtAuth` 中间件
- 给 `user-api` 加 `signature`/`rateLimit` 等 go-zero 内置中间件，观察限流效果
- 把 `user.rpc` 的 `Ping` 换成真正的 `GetUser`，体会 proto 契约变更引发的连锁改动
