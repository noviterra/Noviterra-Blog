package logic

import (
	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"
	"Noviterra/server/utils/common"
	"context"
	"fmt"

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
	// 从上下文中获取用户信息
	publisherID, ok := l.svcCtx.CtxData.GetUserID(l.ctx)
	if !ok {
		return &articlerpc.Response{
			Message: "用户ID获取失败！",
			Code:    1,
		}, nil
	}
	if publisherID == "" {
		return &articlerpc.Response{
			Message: "用户ID为空！",
			Code:    1,
		}, nil
	}

	// 校验用户身份 查库
	role, err := l.svcCtx.Repo.GetUserRole(publisherID)
	if err != nil {
		return &articlerpc.Response{
			Message: "",
			Code:    1,
		}, err
	}
	if role != 1 {
		return &articlerpc.Response{
			Message: "用户权限不足！",
			Code:    1,
		}, nil
	}

	// 校验与清洁文章内容
	cleanArticle, err := l.svcCtx.CommonUtils.ValidateArticle(&common.ValidateArticleInput{
		Title:       in.Title,
		Information: in.Information,
		Content:     in.Content,
		Cover:       in.Cover,
		Tags:        in.Tags,
	})
	if err != nil {
		return &articlerpc.Response{
			Message: "",
			Code:    1,
		}, err
	}

	// 生成文章对应的唯一ID
	articleID := l.svcCtx.SnowFlake.GenerateSnowFlakeID()
	fmt.Println(articleID)

	// 插入到数据库中

	// 加入到缓存中的zset中 根据访问量或点赞量或收藏量进行排行top 对应三个top榜

	return &articlerpc.Response{
		Message: "发布成功！",
		Code:    0,
	}, nil
}
