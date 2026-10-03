package logic

import (
	"context"

	"go-study/go-zero-example/user/internal/svc"
	"go-study/go-zero-example/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PingLogic) Ping(in *user.Request) (*user.Response, error) {
	l.Infof("收到 Ping 请求: %s", in.GetPing())

	return &user.Response{Pong: "pong: " + in.GetPing()}, nil
}
