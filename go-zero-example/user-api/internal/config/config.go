// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	// UserRpc 是下游 user.rpc 服务的客户端配置。
	// 学习用示例采用 Endpoints 直连（本地不需要 etcd）；
	// 生产环境建议改为服务发现：Etcd: {Hosts: [...], Key: user.rpc}
	UserRpc zrpc.RpcClientConf
}
