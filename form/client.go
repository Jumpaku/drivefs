package form

import (
	"context"
	"time"

	"github.com/Jumpaku/go-drivefs"
	"github.com/Jumpaku/go-drivefs/errors"
	"google.golang.org/api/forms/v1"
)

type Client struct {
	service *forms.Service
}

func New(service *forms.Service) *Client {
	return &Client{service: service}
}

func (c *Client) Get(formID drivefs.FileID) (form *Form, err error) {
	f, err := c.service.Forms.Get(string(formID)).Fields().Do()
	if err != nil {
		return nil, errors.NewAPIError("failed to change publish state", err)
	}

	items := []Item{}
	for _, item := range f.Items {
		items = append(items, Item{
			itemId:            item.ItemId,
			infoTitle:         item.Title,
			infoDescription:   item.Description,
			imageItem:         item.ImageItem,
			pageBreakItem:     item.PageBreakItem,
			questionGroupItem: item.QuestionGroupItem,
			questionItem:      item.QuestionItem,
			textItem:          item.TextItem,
			videoItem:         item.VideoItem,
		})
	}

	publishState := PublishStateUnpublished
	if f.PublishSettings.PublishState.IsPublished {
		if f.PublishSettings.PublishState.IsAcceptingResponses {
			publishState = PublishStateAccepting
		} else {
			publishState = PublishStateNotAccepting
		}
	}
	return &Form{
		formID:              drivefs.FileID(f.FormId),
		infoTitle:           f.Info.Title,
		infoDescription:     f.Info.Description,
		emailCollectionType: EmailCollectionType(f.Settings.EmailCollectionType),
		publishState:        publishState,
		items:               items,
	}, nil
}

func (c *Client) Save(form *Form) (result *Form, err error) {
	formID := string(form.FormID())
	if formID == "" {
		items := []*forms.Item{}
		for _, item := range form.items {
			items = append(items, &forms.Item{
				Description:       item.InfoDescription(),
				ImageItem:         item.ImageItem(),
				PageBreakItem:     item.PageBreakItem(),
				QuestionGroupItem: item.QuestionGroupItem(),
				QuestionItem:      item.QuestionItem(),
				TextItem:          item.TextItem(),
				Title:             item.InfoTitle(),
				VideoItem:         item.VideoItem(),
			})
		}
		f, err := c.service.Forms.Create(&forms.Form{
			Info: &forms.Info{
				Description: form.infoDescription,
				Title:       form.infoTitle,
			},
			Settings: &forms.FormSettings{
				EmailCollectionType: string(form.emailCollectionType),
			},
			Items: items,
		}).Do()
		if err != nil {
			return nil, errors.NewAPIError("failed to create form", err)
		}
		formID = f.FormId
	} else {
		updates := []*forms.Request{}
		if form.updateInfoTitle {
			updates = append(updates, &forms.Request{
				UpdateFormInfo: &forms.UpdateFormInfoRequest{
					Info:       &forms.Info{Title: form.InfoTitle()},
					UpdateMask: "title",
				},
			})
		}
		if form.updateInfoDescription {
			updates = append(updates, &forms.Request{
				UpdateFormInfo: &forms.UpdateFormInfoRequest{
					Info:       &forms.Info{Description: form.InfoDescription()},
					UpdateMask: "description",
				},
			})
		}
		if form.updateEmailCollectionType {
			updates = append(updates, &forms.Request{
				UpdateSettings: &forms.UpdateSettingsRequest{
					Settings: &forms.FormSettings{
						EmailCollectionType: string(form.EmailCollectionType()),
					},
					UpdateMask: "email_collection_type",
				},
			})
		}
		updates = append(updates, form.updateItemsRequests...)
		_, err := c.service.Forms.BatchUpdate(formID, &forms.BatchUpdateFormRequest{
			Requests: nil,
		}).Do()
		if err != nil {
			return nil, errors.NewAPIError("failed to update form", err)
		}
	}
	if form.updatePublishState {
		state := form.PublishState()
		_, err = c.service.Forms.SetPublishSettings(formID, &forms.SetPublishSettingsRequest{
			PublishSettings: &forms.PublishSettings{
				PublishState: &forms.PublishState{
					IsAcceptingResponses: state == PublishStateAccepting,
					IsPublished:          state != PublishStateNotAccepting,
				},
			},
			UpdateMask: "publish_state",
		}).Do()
		if err != nil {
			return nil, errors.NewAPIError("failed to change publish state", err)
		}
	}

	return c.Get(drivefs.FileID(formID))
}

func (c *Client) FetchResult(formID drivefs.FileID) (result *FormResult, err error) {
	form, err := c.service.Forms.Get(string(formID)).Do()
	if err != nil {
		return nil, errors.NewAPIError("failed to get form", err)
	}

	var responses []*forms.FormResponse
	err = c.service.Forms.Responses.
		List(string(formID)).
		Pages(context.Background(), func(resp *forms.ListFormResponsesResponse) error {
			responses = append(responses, resp.Responses...)
			return nil
		})
	if err != nil {
		return nil, errors.NewAPIError("failed to list responses", err)
	}
	result = &FormResult{Questions: map[string]string{}}
	for _, item := range form.Items {
		if item.QuestionItem != nil && item.QuestionItem.Question != nil {
			result.Questions[item.QuestionItem.Question.QuestionId] = item.Title
		}
	}

	for _, response := range responses {
		createTime, _ := time.Parse(time.RFC3339Nano, response.CreateTime)
		lastSubmittedTime, _ := time.Parse(time.RFC3339Nano, response.CreateTime)
		answer := FormAnswer{
			ResponseID:        response.ResponseId,
			RespondentEmail:   response.RespondentEmail,
			CreateTime:        createTime,
			LastSubmittedTime: lastSubmittedTime,
			AnswerTexts:       map[QuestionID][]string{},
		}
		for questionId := range result.Questions {
			textAnswers := response.Answers[questionId].TextAnswers
			if textAnswers == nil {
				continue
			}

			answerTexts := []string{}
			for _, textAnswer := range textAnswers.Answers {
				answerTexts = append(answerTexts, textAnswer.Value)
			}

			answer.AnswerTexts[questionId] = answerTexts
		}
		result.Answers = append(result.Answers, answer)
	}

	return result, nil
}
