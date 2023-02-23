package main

import "godvanced.forstes.github.com/internal/data"

func (app *application) EvaluateActivity(activity *data.Activity) {
	maxPoints := int16(len(activity.AnswerPoints) * 3)
	toolBound := maxPoints * 2 / 3

	for _, ans := range activity.AnswerPoints {
		activity.AnswersSum += ans
	}

	switch {
	case activity.AnswersSum == maxPoints:
		activity.Status = data.Ikigai
	case activity.AnswersSum > toolBound:
		activity.Status = data.Tool
	default:
		activity.Status = data.Trash
	}
}
