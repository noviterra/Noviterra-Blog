// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"

	"Noviterra/server/app/comment/commentapi/internal/svc"
	"Noviterra/server/app/comment/commentapi/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CommentArticleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 对一篇文章进行评论
func NewCommentArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentArticleLogic {
	return &CommentArticleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CommentArticleLogic) CommentArticle(req *types.CommentArticleRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
