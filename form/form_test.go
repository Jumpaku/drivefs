package form_test

import (
	"testing"

	"github.com/Jumpaku/go-drivefs/form"
	formsapi "google.golang.org/api/forms/v1"
)

func TestFormItem_UpdateMask(t *testing.T) {
	cases := []struct {
		name string
		set  func(*form.Item)
		want string
	}{
		{"empty", func(i *form.Item) {}, ""},
		{"single-infoTitle", func(i *form.Item) { i.SetInfoTitle("t") }, "infoTitle"},
		{"multi-infoTitle-desc", func(i *form.Item) { i.SetInfoTitle("t"); i.SetInfoDescription("d") }, "infoTitle,infoDescription"},
		{"all", func(i *form.Item) {
			i.SetInfoTitle("t").SetInfoDescription("d").SetImageItem(&formsapi.ImageItem{}).SetPageBreakItem(&formsapi.PageBreakItem{}).SetQuestionGroupItem(&formsapi.QuestionGroupItem{}).SetQuestionItem(&formsapi.QuestionItem{}).SetTextItem(&formsapi.TextItem{}).SetVideoItem(&formsapi.VideoItem{})
		}, "infoTitle,infoDescription,imageItem,pageBreakItem,questionGroupItem,questionItem,textItem,videoItem"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			var it form.Item
			c.set(&it)
			if got := it.UpdateMask(); got != c.want {
				t.Fatalf("UpdateMask() = %q, want %q", got, c.want)
			}
		})
	}
}
