package form

import "time"

type FormResult struct {
	Questions map[QuestionID]string
	Answers   []FormAnswer
}

type QuestionID = string
type FormAnswer struct {
	ResponseID        string
	RespondentEmail   string
	CreateTime        time.Time
	LastSubmittedTime time.Time
	AnswerTexts       map[QuestionID][]string
}
