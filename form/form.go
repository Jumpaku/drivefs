package form

import (
	"github.com/Jumpaku/go-drivefs"
	"google.golang.org/api/forms/v1"
)

type Form struct {
	formID                drivefs.FileID
	infoTitle             string
	updateInfoTitle       bool
	infoDescription       string
	updateInfoDescription bool
	publishState          PublishState
	updatePublishState    bool
	items                 []Item
	updateRequests        []*forms.Request
}

func (f *Form) FormID() (formID drivefs.FileID) {
	return f.formID
}

func (f *Form) SetInfoTitle(title string) *Form {
	f.updateInfoTitle = true
	f.infoTitle = title
	return f
}
func (f *Form) InfoTitle() (title string) {
	return f.infoTitle
}

func (f *Form) SetInfoDescription(description string) *Form {
	f.updateInfoDescription = true
	f.infoDescription = description
	return f
}
func (f *Form) InfoDescription() (description string) {
	return f.infoDescription
}

func (f *Form) SetPublishState(publishState PublishState) *Form {
	f.updatePublishState = true
	f.publishState = publishState
	return f
}
func (f *Form) PublishState() (publishState PublishState) {
	return f.publishState
}

func (f *Form) Items() (items []Item) {
	return append([]Item{}, f.items...)
}

func (f *Form) CreateItem(index int, item Item) *Form {
	old := f.items
	f.items = append(append(append([]Item{}, old[:index]...), item), old[index:]...)

	f.updateRequests = append(f.updateRequests, &forms.Request{
		CreateItem: &forms.CreateItemRequest{
			Item: &forms.Item{
				ItemId:            item.ItemId(),
				Title:             item.InfoTitle(),
				Description:       item.InfoDescription(),
				ImageItem:         item.ImageItem(),
				PageBreakItem:     item.PageBreakItem(),
				QuestionGroupItem: item.QuestionGroupItem(),
				QuestionItem:      item.QuestionItem(),
				TextItem:          item.TextItem(),
				VideoItem:         item.VideoItem(),
			},
			Location: &forms.Location{Index: int64(index)},
		},
	})

	return f
}

func (f *Form) DeleteItem(index int) *Form {
	old := f.items
	f.items = append(old[:index], old[index+1:]...)

	f.updateRequests = append(f.updateRequests, &forms.Request{
		DeleteItem: &forms.DeleteItemRequest{
			Location: &forms.Location{Index: int64(index)},
		},
	})

	return f
}

func (f *Form) MoveItem(index, newIndex int) *Form {
	old := f.items
	v := old[index]
	old = append(old[:index], old[index+1:]...)
	f.items = append(append(old[:newIndex], v), old[newIndex:]...)

	f.updateRequests = append(f.updateRequests, &forms.Request{
		MoveItem: &forms.MoveItemRequest{
			NewLocation:      &forms.Location{Index: int64(newIndex)},
			OriginalLocation: &forms.Location{Index: int64(index)},
		},
	})

	return f
}

func (f *Form) UpdateItem(index int, item Item) *Form {
	f.items[index] = item

	f.updateRequests = append(f.updateRequests, &forms.Request{
		UpdateItem: &forms.UpdateItemRequest{
			Item: &forms.Item{
				ItemId:            item.ItemId(),
				Title:             item.InfoTitle(),
				Description:       item.InfoDescription(),
				ImageItem:         item.ImageItem(),
				PageBreakItem:     item.PageBreakItem(),
				QuestionGroupItem: item.QuestionGroupItem(),
				QuestionItem:      item.QuestionItem(),
				TextItem:          item.TextItem(),
				VideoItem:         item.VideoItem(),
			},
			Location:   &forms.Location{Index: int64(index)},
			UpdateMask: item.UpdateMask(),
		},
	})

	return f
}
