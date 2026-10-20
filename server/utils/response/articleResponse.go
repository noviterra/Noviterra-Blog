package response

import (
	"Noviterra/server/app/article/articlerpc/types/articlerpc"
	"Noviterra/server/utils/errcode"
)

type ArticleResponse struct {
}

func NewArticleResponse() *ArticleResponse {
	return &ArticleResponse{}
}

// RPCSuccessResponse 封装response响应并返回
func (arp *ArticleResponse) RPCSuccessResponse(msg string) (*articlerpc.Response, error) {
	return &articlerpc.Response{
		Message: msg,
		Code:    errcode.Success,
	}, nil
}

// RPCFailResponse 返回错误提示而非错误本身
func (arp *ArticleResponse) RPCFailResponse(msg string, code int64) (*articlerpc.Response, error) {
	return &articlerpc.Response{
		Message: msg,
		Code:    code,
	}, nil
}

// RPCFailResponseWithErr 返回错误本身 rpc层返回error api层不会收到结构体中的内容
func (arp *ArticleResponse) RPCFailResponseWithErr(err error) (*articlerpc.Response, error) {
	return &articlerpc.Response{
		Message: "",
		Code:    errcode.Failed,
	}, err
}

// RPCSuccessGetArticleListResponse 封装GetArticleListResponse
func (arp *ArticleResponse) RPCSuccessGetArticleListResponse(msg string, articles []*articlerpc.ArticleList, totalCount int32) (*articlerpc.GetArticleListResponse, error) {
	return &articlerpc.GetArticleListResponse{
		Articles:   articles,
		TotalCount: totalCount,
		Resp: &articlerpc.Response{
			Message: msg,
			Code:    errcode.Success,
		},
	}, nil
}

// RPCFailGetArticleListResponse 不携带err
func (arp *ArticleResponse) RPCFailGetArticleListResponse(msg string, code int64) (*articlerpc.GetArticleListResponse, error) {
	return &articlerpc.GetArticleListResponse{
		Articles:   nil,
		TotalCount: 0,
		Resp: &articlerpc.Response{
			Message: msg,
			Code:    code,
		},
	}, nil
}

// RPCFailGetArticleListResponseWithErr 携带err
func (arp *ArticleResponse) RPCFailGetArticleListResponseWithErr(err error) (*articlerpc.GetArticleListResponse, error) {
	return &articlerpc.GetArticleListResponse{
		Articles:   nil,
		TotalCount: 0,
		Resp: &articlerpc.Response{
			Message: "",
			Code:    errcode.Failed,
		},
	}, err
}

// RPCSuccessGetArticleDetailResponse 封装响应GetArticleDetailResponse
func (arp *ArticleResponse) RPCSuccessGetArticleDetailResponse(msg string, article *articlerpc.Article) (*articlerpc.GetArticleDetailResponse, error) {
	return &articlerpc.GetArticleDetailResponse{
		A: article,
		Resp: &articlerpc.Response{
			Message: msg,
			Code:    errcode.Success,
		},
	}, nil
}

func (arp *ArticleResponse) RPCFailGetArticleDetailResponse(msg string, code int64) (*articlerpc.GetArticleDetailResponse, error) {
	return &articlerpc.GetArticleDetailResponse{
		A: nil,
		Resp: &articlerpc.Response{
			Message: msg,
			Code:    code,
		},
	}, nil
}

func (arp *ArticleResponse) RPCFailGetArticleDetailResponseWithErr(err error) (*articlerpc.GetArticleDetailResponse, error) {
	return &articlerpc.GetArticleDetailResponse{
		A: nil,
		Resp: &articlerpc.Response{
			Message: "",
			Code:    errcode.Failed,
		},
	}, err
}

// RPCSuccessUpdateArticleCoverResponse 封装UpdateArticleCoverResponse响应体
func (arp *ArticleResponse) RPCSuccessUpdateArticleCoverResponse(msg, cover string) (*articlerpc.UpdateArticleCoverResponse, error) {
	return &articlerpc.UpdateArticleCoverResponse{
		Cover: cover,
		Resp: &articlerpc.Response{
			Message: msg,
			Code:    errcode.Success,
		},
	}, nil
}

func (arp *ArticleResponse) RPCFailUpdateArticleCoverResponse(msg string, code int64) (*articlerpc.UpdateArticleCoverResponse, error) {
	return &articlerpc.UpdateArticleCoverResponse{
		Cover: "",
		Resp: &articlerpc.Response{
			Message: msg,
			Code:    code,
		},
	}, nil
}

func (arp *ArticleResponse) RPCFailUpdateArticleCoverResponseWithErr(err error) (*articlerpc.UpdateArticleCoverResponse, error) {
	return &articlerpc.UpdateArticleCoverResponse{
		Cover: "",
		Resp: &articlerpc.Response{
			Message: "",
			Code:    0,
		},
	}, err
}

// RPCSuccessPublishArticleResponse 封装PublishArticleResponse
func (arp *ArticleResponse) RPCSuccessPublishArticleResponse(msg string, articleID string) (*articlerpc.PublishArticleResponse, error) {
	return &articlerpc.PublishArticleResponse{
		ArticleID: articleID,
		Resp: &articlerpc.Response{
			Message: msg,
			Code:    errcode.Success,
		},
	}, nil
}

func (arp *ArticleResponse) RPCFailPublishArticleResponse(msg string, code int64) (*articlerpc.PublishArticleResponse, error) {
	return &articlerpc.PublishArticleResponse{
		ArticleID: "",
		Resp: &articlerpc.Response{
			Message: msg,
			Code:    code,
		},
	}, nil
}

func (arp *ArticleResponse) RPCFailPublishArticleResponseWithErr(err error) (*articlerpc.PublishArticleResponse, error) {
	return &articlerpc.PublishArticleResponse{
		ArticleID: "",
		Resp: &articlerpc.Response{
			Message: "",
			Code:    errcode.Failed,
		},
	}, err
}

func (arp *ArticleResponse) RPCSuccessSearchArticleResponse(msg string, articles []*articlerpc.ArticleList, totalCount int64) (*articlerpc.SearchArticleResponse, error) {
	return &articlerpc.SearchArticleResponse{
		Articles:   articles,
		TotalCount: int32(totalCount),
		Resp: &articlerpc.Response{
			Message: msg,
			Code:    errcode.Success,
		},
	}, nil
}

func (arp *ArticleResponse) RPCFailSearchArticleResponse(msg string, code int64) (*articlerpc.SearchArticleResponse, error) {
	return &articlerpc.SearchArticleResponse{
		Articles:   nil,
		TotalCount: 0,
		Resp: &articlerpc.Response{
			Message: msg,
			Code:    code,
		},
	}, nil
}

func (arp *ArticleResponse) RPCFailSearchArticleResponseWithErr(err error) (*articlerpc.SearchArticleResponse, error) {
	return &articlerpc.SearchArticleResponse{
		Articles:   nil,
		TotalCount: 0,
		Resp: &articlerpc.Response{
			Message: "",
			Code:    errcode.Failed,
		},
	}, err
}
