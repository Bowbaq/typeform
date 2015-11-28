package typeform

type YesNoQuestion struct {
	BaseQuestion

	Question    string
	Description string

	Options *YesNoQuestionOptions
	Jumps   []Jump
}

func (ynq YesNoQuestion) Q() string     { return ynq.Question }
func (ynq YesNoQuestion) Desc() string  { return ynq.Description }
func (ynq YesNoQuestion) Logic() []Jump { return ynq.Jumps }

func (ynq YesNoQuestion) label() string    { return "Yes / No" }
func (ynq YesNoQuestion) prefix() string   { return "yes_no" }
func (ynq YesNoQuestion) selector() string { return "question" }

type YesNoQuestionOptions struct {
	Required bool
}

func (f *form) AddYesNoQuestion(q *YesNoQuestion) FormElement {
	f.createQuestion(q)

	// Options
	if q.Options != nil {
		f.setOption(q, q.Options.Required, "required")
	}

	return f.saveQuestion(q, q.Options.Required)
}
