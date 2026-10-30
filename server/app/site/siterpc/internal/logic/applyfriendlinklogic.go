package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/app/site/siterpc/internal/svc"
	"Noviterra/server/app/site/siterpc/types/siterpc"
	"Noviterra/server/utils/errcode"
	"Noviterra/server/utils/validate"
	"context"
	"errors"

	"go.uber.org/zap"
)

type (
	applyFriendLinkSiteRepository interface {
		ApplyFriendLink(ctx context.Context, linkInfo model.FriendLink) error
	}

	applyFriendLinkSiteResponse interface {
		SuccessResponse(resp *siterpc.Response, msg string) (*siterpc.Response, error)
		FailResponse(resp *siterpc.Response, msg string, status int32, code int64) (*siterpc.Response, error)
		FailResponseWithErr(resp *siterpc.Response, err error) (*siterpc.Response, error)
	}

	applyFriendLinkLogger interface {
		Error(msg string, fields ...zap.Field)
		Info(msg string, fields ...zap.Field)
	}

	applyFriendLinkCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}

	applyFriendLinkSiteValidate interface {
		ValidateFriendLinkInfo(link validate.InputValidateFriendLink) (validate.OutputValidateFriendLink, error)
	}

	applyFriendLinkIdentityRespository interface {
		GetUserEmailByID(id string) (string, error)
	}
)

type ApplyFriendLinkLogic struct {
	ctx                context.Context
	logger             applyFriendLinkLogger
	siteRepository     applyFriendLinkSiteRepository
	identityRepository applyFriendLinkIdentityRespository
	siteResponse       applyFriendLinkSiteResponse
	ctxdata            applyFriendLinkCtxData
	siteValidate       applyFriendLinkSiteValidate
}

func NewApplyFriendLinkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyFriendLinkLogic {
	return &ApplyFriendLinkLogic{
		ctx:                ctx,
		logger:             svcCtx.Logger,
		siteRepository:     svcCtx.SiteRepository,
		identityRepository: svcCtx.IdentityRepository,
		siteResponse:       svcCtx.SiteResponse,
		ctxdata:            svcCtx.CtxData,
		siteValidate:       svcCtx.SiteValidate,
	}
}

// 申请友链rpc方法
func (l *ApplyFriendLinkLogic) ApplyFriendLink(in *siterpc.FriendLinkRequest) (*siterpc.Response, error) {
	// todo: add your logic here and delete this line
	resp := new(siterpc.Response)
	// 用户身份校验
	id := l.ctxdata.GetUserIDFromContext(l.ctx)
	if id == "" {
		l.logger.Error("site/applyfriendlink：上下文获取用户ID为空！")
		return l.siteResponse.FailResponse(resp, "用户信息校验失败！", errcode.InvalidParam, errcode.Failed)
	}
	// 参数校验
	linkInfo, err := l.siteValidate.ValidateFriendLinkInfo(validate.InputValidateFriendLink{
		Link:        in.RequestInfo.Link,
		LinkName:    in.RequestInfo.LinkName,
		Information: in.RequestInfo.SelfInformation,
	})
	email := in.Email
	if err != nil {
		l.logger.Error("site/applyfriendlink：请求参数校验未通过！",
			zap.String("ERROR", err.Error()), zap.Any("link_info", linkInfo))
		return l.siteResponse.FailResponseWithErr(resp, err)
	}

	if email == "" {
		email, err = l.identityRepository.GetUserEmailByID(id)
		if err != nil {
			l.logger.Error("site/applyfriendlink：数据库异常，查询用户邮箱失败！",
				zap.String("ERROR", err.Error()))
			return l.siteResponse.FailResponse(resp, "数据库操作异常！", errcode.InternalError, errcode.Failed)
		}
		if email == "" {
			l.logger.Error("site/applyfriendlink：用户信息校验失败，查询到的用户邮箱为空！")
			return l.siteResponse.FailResponse(resp, "邮箱查询失败，请手动输入邮箱！", errcode.InvalidParam, errcode.Failed)
		}
	}

	// 插入库中
	err = l.siteRepository.ApplyFriendLink(l.ctx, model.FriendLink{
		Link:        in.RequestInfo.Link,
		LinkName:    in.RequestInfo.LinkName,
		Information: in.RequestInfo.SelfInformation,
		Applicant:   id,
		Email:       email,
		Status:      0,
	})
	var selfRepoErr *errcode.RepoError
	if err != nil {
		if errors.As(err, &selfRepoErr) { // 自定义err 服务被拒绝 展示给用户
			l.logger.Error("site/applyfriendlink：数据库插入友链操作未执行完成，用户服务被拒绝",
				zap.String("ERROR", err.Error()), zap.Any("link_info", linkInfo))
			return l.siteResponse.FailResponseWithErr(resp, err)
		}
		l.logger.Error("site/applyfriendlink：数据库异常，插入友链数据操作失败！",
			zap.String("ERROR", err.Error()))
		return l.siteResponse.FailResponse(resp, "数据库操作异常！", errcode.InternalError, errcode.Failed)
	}

	return l.siteResponse.SuccessResponse(resp, "友链申请成功，后续进度将通过邮箱进行通知！")
}
