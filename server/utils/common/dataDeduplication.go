package common

import "Noviterra/server/app/model"

type DataDeduplication struct {
}

func NewDataDeduplication() *DataDeduplication {
	return &DataDeduplication{}
}

// go泛型 泛型函数不能绑定在结构体之下
func dedupSpeakerIDs[T any](list []T, getID func(T) string) []string {
	if len(list) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(list))
	out := make([]string, 0, len(list))
	for _, item := range list {
		id := getID(item)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// GetUserIDListsFromTopComment 从顶层评论去重提取 SpeakerIDs
func (dd *DataDeduplication) GetUserIDListsFromTopComment(commentList []model.TopCommentNode) []string {
	return dedupSpeakerIDs(commentList, func(c model.TopCommentNode) string { return c.SpeakerID })
}

// GetUserIDListsFromTreeComment 从子评论数据中去重提取 SpeakerIDs
func (dd *DataDeduplication) GetUserIDListsFromTreeComment(commentList []model.CommentNode) []string {
	return dedupSpeakerIDs(commentList, func(c model.CommentNode) string { return c.SpeakerID })
}
