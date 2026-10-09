package logic

import (
	"context"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateArticleCoverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateArticleCoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateArticleCoverLogic {
	return &UpdateArticleCoverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改文章封面
func (l *UpdateArticleCoverLogic) UpdateArticleCover(in *articlerpc.UpdateArticleCoverRequest) (*articlerpc.UpdateArticleCoverResponse, error) {
	// todo: add your logic here and delete this line

	return &articlerpc.UpdateArticleCoverResponse{}, nil
}
