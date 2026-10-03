// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"go-study/go-zero-example/user/userclient"

	"user-api/internal/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	// UserRpc 是对下游 user.rpc 服务的 gRPC 客户端封装，
	// 由 goctl 依据 user.proto 生成在 userclient 包中。
	UserRpc userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		UserRpc: userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}
