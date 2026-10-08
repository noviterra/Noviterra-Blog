package logic

import (
	"context"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishArticleLogic {
	return &PublishArticleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发布文章
func (l *PublishArticleLogic) PublishArticle(in *articlerpc.PublishArticleRequest) (*articlerpc.Response, error) {
	// todo: add your logic here and delete this line

	return &articlerpc.Response{}, nil
}
