package typeform

import "strings"

type DropDownQuestion struct {
	BaseQuestion

	Question    string
	Description string

	Choices []string

	Options *DropDownQuestionOptions
	Jumps   []Jump
}

func (mcq DropDownQuestion) Q() string     { return mcq.Question }
func (mcq DropDownQuestion) Desc() string  { return mcq.Description }
func (mcq DropDownQuestion) Logic() []Jump { return mcq.Jumps }

func (mcq DropDownQuestion) label() string    { return "Dropdown" }
func (mcq DropDownQuestion) prefix() string   { return "dropdown" }
func (mcq DropDownQuestion) selector() string { return "question" }

type DropDownQuestionOptions struct {
	Required     bool
	Alphabetical bool
}

func (f *form) AddDropDownQuestion(q *DropDownQuestion) FormElement {
	f.createQuestion(q)

	f.t.driverT.Q("#dropdown_options").SendKeys(strings.Join(q.Choices, "\n"))

	// Options
	if q.Options != nil {
		f.setOption(q, q.Options.Required, "required")
		f.setOption(q, q.Options.Alphabetical, "alphabetical")
	}

	return f.saveQuestion(q, q.Options.Required)
}
