package typeform

import "time"

type MultipleChoiceQuestion struct {
	BaseQuestion

	Question    string
	Description string

	Choices []string

	Options *MultipleChoiceQuestionOptions
	Jumps   []Jump
}

func (mcq MultipleChoiceQuestion) Q() string     { return mcq.Question }
func (mcq MultipleChoiceQuestion) Desc() string  { return mcq.Description }
func (mcq MultipleChoiceQuestion) Logic() []Jump { return mcq.Jumps }

func (mcq MultipleChoiceQuestion) label() string    { return "Multiple choice" }
func (mcq MultipleChoiceQuestion) prefix() string   { return "list" }
func (mcq MultipleChoiceQuestion) selector() string { return "question" }

type MultipleChoiceQuestionOptions struct {
	Required               bool
	AllowMutipleSelections bool
	Randomize              bool
	ForceVerticalAlignment bool
	AddOtherOption         bool
}

func (f *form) AddMultipleChoiceQuestion(q *MultipleChoiceQuestion) FormElement {
	f.createQuestion(q)

	// Choices
	for _, label := range q.Choices {
		choice := f.t.driverT.Q(".choices .choice:last-child")
		choice.Q("input[type='text']").SendKeys(label)
		choice.Q(".add").Click()

		time.Sleep(300 * time.Millisecond)
	}
	f.t.driverT.Q(".choices .choice:last-child .remove").Click()

	// Options
	if q.Options != nil {
		f.setOption(q, q.Options.Required, "required")
		f.setOption(q, q.Options.AllowMutipleSelections, "multiple")
		f.setOption(q, q.Options.Randomize, "randomize")
		f.setOption(q, q.Options.ForceVerticalAlignment, "vertical")
		f.setOption(q, q.Options.AddOtherOption, "other")
	}

	return f.saveQuestion(q, q.Options.Required)
}
