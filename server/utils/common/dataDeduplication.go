package common

import "Noviterra/server/app/model"

type DataDeduplication struct {
}

func NewDataDeduplication() *DataDeduplication {
	return &DataDeduplication{}
}

func (dd *DataDeduplication) GetUserIDListsFromTopComment(commentList []model.TopCommentNode) []string {
	if len(commentList) == 0 {
		return []string{}
	}

	hasSpeakerIDs := make(map[string]struct{}, len(commentList))
	speakerIDs := make([]string, 0, len(commentList))
	for _, comment := range commentList {
		id := comment.SpeakerID
		if _, ok := hasSpeakerIDs[id]; ok {
			continue
		}
		hasSpeakerIDs[id] = struct{}{}
		speakerIDs = append(speakerIDs, id)
	}

	return speakerIDs
}
