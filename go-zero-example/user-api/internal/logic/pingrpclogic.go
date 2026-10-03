// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"user-api/internal/svc"
	"user-api/internal/types"

	"go-study/go-zero-example/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingRpcLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPingRpcLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingRpcLogic {
	return &PingRpcLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// PingRpc 演示 API 层调用下游 RPC 服务的完整链路：
// HTTP handler -> logic -> userclient(gRPC) -> user.rpc -> pong
func (l *PingRpcLogic) PingRpc(req *types.PingRpcReq) (resp *types.PingRpcResp, err error) {
	reply, err := l.svcCtx.UserRpc.Ping(l.ctx, &user.Request{Ping: req.Message})
	if err != nil {
		l.Errorf("调用 user.rpc Ping 失败: %v", err)
		return nil, err
	}

	return &types.PingRpcResp{Pong: reply.GetPong()}, nil
}
