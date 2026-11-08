package response

import (
	"Noviterra/server/app/article/articlerpc/types/articlerpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
)

type ArticleResponse struct {
}

func NewArticleResponse() *ArticleResponse {
	return &ArticleResponse{}
}

// v2-20261022

// SuccessResponse 封装response响应并返回
func (arp *ArticleResponse) SuccessResponse(resp *articlerpc.Response, msg string) (*articlerpc.Response, error) {
	if resp == nil {
		resp = new(articlerpc.Response)
	}
	resp.Message = msg
	resp.Code = errcode.Success
	resp.Status = errcode.RequestSuccess

	return resp, nil
}

// FailResponse 返回错误提示而非错误本身
func (arp *ArticleResponse) FailResponse(resp *articlerpc.Response, msg string, status int32, code int64) (*articlerpc.Response, error) {
	if resp == nil {
		resp = new(articlerpc.Response)
	}
	resp.Message = msg
	resp.Code = code
	resp.Status = status

	return resp, nil
}

// FailResponseWithErr 返回错误本身 rpc层返回error api层不会收到结构体中的内容
func (arp *ArticleResponse) FailResponseWithErr(resp *articlerpc.Response, err error) (*articlerpc.Response, error) {
	if resp == nil {
		resp = new(articlerpc.Response)
	}

	return resp, err
}

// SuccessGetArticleListResponse 封装GetArticleListResponse
func (arp *ArticleResponse) SuccessGetArticleListResponse(resp *articlerpc.GetArticleListResponse,
	msg string, articles []model.ArticleList, totalCount int64) (*articlerpc.GetArticleListResponse, error) {
	if resp == nil {
		resp = new(articlerpc.GetArticleListResponse)
	}

	resp.TotalCount = totalCount
	if totalCount > 0 {
		for _, list := range articles {
			ls := new(articlerpc.ArticleList)

			ls.ArticleID = list.ArticleID
			ls.Title = list.Title
			ls.Information = list.Information
			ls.Tags = list.Tags
			ls.Cover = list.Cover
			ls.CreatedAt = list.CreatedAt.Unix()

			resp.Articles = append(resp.Articles, ls)
		}
	} else {
		resp.Articles = []*articlerpc.ArticleList{}
	}

	resp.Resp = &articlerpc.Response{
		Message: msg,
		Code:    errcode.Success,
	}

	return resp, nil
}

// FailGetArticleListResponse 不携带err
func (arp *ArticleResponse) FailGetArticleListResponse(resp *articlerpc.GetArticleListResponse,
	msg string, status int32, code int64) (*articlerpc.GetArticleListResponse, error) {
	if resp == nil {
		resp = new(articlerpc.GetArticleListResponse)
	}

	resp.Articles = nil
	resp.TotalCount = 0
	resp.Resp = &articlerpc.Response{
		Message: msg,
		Code:    code,
		Status:  status,
	}

	return resp, nil
}

// FailGetArticleListResponseWithErr 携带err
func (arp *ArticleResponse) FailGetArticleListResponseWithErr(resp *articlerpc.GetArticleListResponse,
	err error) (*articlerpc.GetArticleListResponse, error) {
	if resp == nil {
		resp = new(articlerpc.GetArticleListResponse)
	}

	return resp, err
}

// SuccessGetArticleDetailResponse 封装响应GetArticleDetailResponse
func (arp *ArticleResponse) SuccessGetArticleDetailResponse(resp *articlerpc.GetArticleDetailResponse,
	msg string, article *model.Article) (*articlerpc.GetArticleDetailResponse, error) {
	if resp == nil {
		resp = new(articlerpc.GetArticleDetailResponse)
	}
	if resp.A == nil {
		resp.A = new(articlerpc.Article)
	}

	resp.A.ArticleID = article.ArticleID
	resp.A.Title = article.Title
	resp.A.Information = article.Information
	resp.A.Cover = article.Cover
	resp.A.Content = article.Content
	resp.A.PublisherID = article.PublisherID
	resp.A.Tags = article.Tags
	resp.A.UpdatedAt = article.UpdatedAt
	resp.A.CreatedAt = article.CreatedAt

	resp.Resp = &articlerpc.Response{
		Message: msg,
		Code:    errcode.Success,
	}

	return resp, nil
}

func (arp *ArticleResponse) FailGetArticleDetailResponse(resp *articlerpc.GetArticleDetailResponse,
	msg string, status int32, code int64) (*articlerpc.GetArticleDetailResponse, error) {
	if resp == nil {
		resp = new(articlerpc.GetArticleDetailResponse)
	}

	resp.A = nil
	resp.Resp = &articlerpc.Response{
		Message: msg,
		Code:    code,
		Status:  status,
	}

	return resp, nil
}

func (arp *ArticleResponse) FailGetArticleDetailResponseWithErr(resp *articlerpc.GetArticleDetailResponse,
	err error) (*articlerpc.GetArticleDetailResponse, error) {
	if resp == nil {
		resp = new(articlerpc.GetArticleDetailResponse)
	}

	return resp, err
}

// SuccessUpdateArticleCoverResponse 封装UpdateArticleCoverResponse响应体
func (arp *ArticleResponse) SuccessUpdateArticleCoverResponse(resp *articlerpc.UpdateArticleCoverResponse,
	msg, cover string) (*articlerpc.UpdateArticleCoverResponse, error) {
	if resp == nil {
		resp = new(articlerpc.UpdateArticleCoverResponse)
	}

	resp.Cover = cover
	resp.Resp = &articlerpc.Response{
		Message: msg,
		Code:    errcode.Success,
	}

	return resp, nil
}

func (arp *ArticleResponse) FailUpdateArticleCoverResponse(resp *articlerpc.UpdateArticleCoverResponse,
	msg string, status int32, code int64) (*articlerpc.UpdateArticleCoverResponse, error) {
	if resp == nil {
		resp = new(articlerpc.UpdateArticleCoverResponse)
	}

	resp.Cover = ""
	resp.Resp = &articlerpc.Response{
		Message: msg,
		Code:    code,
		Status:  status,
	}

	return resp, nil
}

func (arp *ArticleResponse) FailUpdateArticleCoverResponseWithErr(resp *articlerpc.UpdateArticleCoverResponse,
	err error) (*articlerpc.UpdateArticleCoverResponse, error) {
	if resp == nil {
		resp = new(articlerpc.UpdateArticleCoverResponse)
	}

	return resp, err
}

// SuccessPublishArticleResponse 封装PublishArticleResponse
func (arp *ArticleResponse) SuccessPublishArticleResponse(resp *articlerpc.PublishArticleResponse,
	msg string, articleID int64) (*articlerpc.PublishArticleResponse, error) {
	if resp == nil {
		resp = new(articlerpc.PublishArticleResponse)
	}

	resp.ArticleID = articleID
	resp.Resp = &articlerpc.Response{
		Message: msg,
		Code:    errcode.Success,
	}

	return resp, nil
}

func (arp *ArticleResponse) FailPublishArticleResponse(resp *articlerpc.PublishArticleResponse,
	msg string, status int32, code int64) (*articlerpc.PublishArticleResponse, error) {
	if resp == nil {
		resp = new(articlerpc.PublishArticleResponse)
	}

	resp.ArticleID = -1
	resp.Resp = &articlerpc.Response{
		Message: msg,
		Code:    code,
		Status:  status,
	}

	return resp, nil
}

func (arp *ArticleResponse) FailPublishArticleResponseWithErr(resp *articlerpc.PublishArticleResponse,
	err error) (*articlerpc.PublishArticleResponse, error) {
	if resp == nil {
		resp = new(articlerpc.PublishArticleResponse)
	}

	return resp, err
}

// SuccessSearchArticleResponse 封装SearchArticleResponse
func (arp *ArticleResponse) SuccessSearchArticleResponse(resp *articlerpc.SearchArticleResponse,
	msg string, articles []*articlerpc.ArticleList, totalCount int64) (*articlerpc.SearchArticleResponse, error) {
	if resp == nil {
		resp = new(articlerpc.SearchArticleResponse)
	}

	resp.Articles = articles
	resp.TotalCount = totalCount
	resp.Resp = &articlerpc.Response{
		Message: msg,
		Code:    errcode.Success,
	}

	return resp, nil
}

func (arp *ArticleResponse) FailSearchArticleResponse(resp *articlerpc.SearchArticleResponse,
	msg string, status int32, code int64) (*articlerpc.SearchArticleResponse, error) {
	if resp == nil {
		resp = new(articlerpc.SearchArticleResponse)
	}

	resp.Articles = nil
	resp.TotalCount = 0
	resp.Resp = &articlerpc.Response{
		Message: msg,
		Code:    code,
		Status:  status,
	}

	return resp, nil
}

func (arp *ArticleResponse) FailSearchArticleResponseWithErr(resp *articlerpc.SearchArticleResponse,
	err error) (*articlerpc.SearchArticleResponse, error) {
	if resp == nil {
		resp = new(articlerpc.SearchArticleResponse)
	}

	return resp, err
}

// SuccessGetArchiveResponse 封装GetArchiveResponse
func (arp *ArticleResponse) SuccessGetArchiveResponse(resp *articlerpc.GetArchiveResponse,
	msg string, archives []model.ArticleArchive) (*articlerpc.GetArchiveResponse, error) {
	// 空指针保护
	if resp == nil {
		resp = new(articlerpc.GetArchiveResponse)
	}

	if len(archives) > 0 {
		// 数据映射
		for _, archive := range archives {
			var a articlerpc.ArchiveArticle

			a.Title = archive.Title
			a.Time = archive.Time
			a.Action = archive.Action

			resp.Archives = append(resp.Archives, &a)
		}
	}

	resp.Resp = &articlerpc.Response{
		Message: msg,
		Code:    errcode.Success,
	}

	return resp, nil
}

func (arp *ArticleResponse) FailGetArchiveResponse(resp *articlerpc.GetArchiveResponse,
	msg string, status int32, code int64) (*articlerpc.GetArchiveResponse, error) {
	if resp == nil {
		resp = new(articlerpc.GetArchiveResponse)
	}

	resp.Archives = nil
	resp.Resp = &articlerpc.Response{
		Message: msg,
		Code:    code,
		Status:  status,
	}

	return resp, nil
}

func (arp *ArticleResponse) FailGetArchiveResponseWithErr(resp *articlerpc.GetArchiveResponse,
	err error) (*articlerpc.GetArchiveResponse, error) {
	if resp == nil {
		resp = new(articlerpc.GetArchiveResponse)
	}

	return resp, err
}
