package form

import (
	"strings"

	"google.golang.org/api/forms/v1"
)

type Item struct {
	itemId            string
	infoTitle         string
	infoDescription   string
	imageItem         *forms.ImageItem
	pageBreakItem     *forms.PageBreakItem
	questionGroupItem *forms.QuestionGroupItem
	questionItem      *forms.QuestionItem
	textItem          *forms.TextItem
	videoItem         *forms.VideoItem

	updateInfoTitle         bool
	updateInfoDescription   bool
	updateImageItem         bool
	updatePageBreakItem     bool
	updateQuestionGroupItem bool
	updateQuestionItem      bool
	updateTextItem          bool
	updateVideoItem         bool
}

func (i *Item) SetInfoTitle(title string) *Item {
	i.updateInfoTitle = true
	i.infoTitle = title
	return i
}

func (i *Item) SetInfoDescription(description string) *Item {
	i.updateInfoDescription = true
	i.infoDescription = description
	return i
}

func (i *Item) SetImageItem(img *forms.ImageItem) *Item {
	i.updateImageItem = true
	i.imageItem = img
	return i
}

func (i *Item) SetPageBreakItem(pb *forms.PageBreakItem) *Item {
	i.updatePageBreakItem = true
	i.pageBreakItem = pb
	return i
}

func (i *Item) SetQuestionGroupItem(qg *forms.QuestionGroupItem) *Item {
	i.updateQuestionGroupItem = true
	i.questionGroupItem = qg
	return i
}

func (i *Item) SetQuestionItem(q *forms.QuestionItem) *Item {
	i.updateQuestionItem = true
	i.questionItem = q
	return i
}

func (i *Item) SetTextItem(t *forms.TextItem) *Item {
	i.updateTextItem = true
	i.textItem = t
	return i
}

func (i *Item) SetVideoItem(v *forms.VideoItem) *Item {
	i.updateVideoItem = true
	i.videoItem = v
	return i
}

// Getters
func (i Item) ItemId() string                              { return i.itemId }
func (i Item) InfoTitle() string                           { return i.infoTitle }
func (i Item) InfoDescription() string                     { return i.infoDescription }
func (i Item) ImageItem() *forms.ImageItem                 { return i.imageItem }
func (i Item) PageBreakItem() *forms.PageBreakItem         { return i.pageBreakItem }
func (i Item) QuestionGroupItem() *forms.QuestionGroupItem { return i.questionGroupItem }
func (i Item) QuestionItem() *forms.QuestionItem           { return i.questionItem }
func (i Item) TextItem() *forms.TextItem                   { return i.textItem }
func (i Item) VideoItem() *forms.VideoItem                 { return i.videoItem }

// UpdateMask returns a comma-separated list of field names for which the
// corresponding update flag is true. Field names are lowerCamelCase (e.g.
// itemId, infoTitle, infoDescription, ...).
func (i Item) UpdateMask() string {
	var parts []string
	if i.updateInfoTitle {
		parts = append(parts, "infoTitle")
	}
	if i.updateInfoDescription {
		parts = append(parts, "infoDescription")
	}
	if i.updateImageItem {
		parts = append(parts, "imageItem")
	}
	if i.updatePageBreakItem {
		parts = append(parts, "pageBreakItem")
	}
	if i.updateQuestionGroupItem {
		parts = append(parts, "questionGroupItem")
	}
	if i.updateQuestionItem {
		parts = append(parts, "questionItem")
	}
	if i.updateTextItem {
		parts = append(parts, "textItem")
	}
	if i.updateVideoItem {
		parts = append(parts, "videoItem")
	}
	return strings.Join(parts, ",")
}
