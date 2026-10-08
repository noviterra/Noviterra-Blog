package logic

import (
	"context"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchArticleLogic {
	return &SearchArticleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 搜索文章
func (l *SearchArticleLogic) SearchArticle(in *articlerpc.SearchArticleRequest) (*articlerpc.SearchArticleResponse, error) {
	// todo: add your logic here and delete this line

	return &articlerpc.SearchArticleResponse{}, nil
}
