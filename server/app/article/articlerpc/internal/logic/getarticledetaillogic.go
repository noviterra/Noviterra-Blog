package logic

import (
	"context"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetArticleDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetArticleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleDetailLogic {
	return &GetArticleDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询文章详细信息
func (l *GetArticleDetailLogic) GetArticleDetail(in *articlerpc.GetArticleDetailRequest) (*articlerpc.GetArticleDetailResponse, error) {
	// todo: add your logic here and delete this line

	return &articlerpc.GetArticleDetailResponse{}, nil
}
