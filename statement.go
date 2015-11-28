package typeform

type StatementContent struct {
	BaseQuestion

	Statement   string
	Description string

	ButtonText string

	Options *StatementContentOptions
	Jumps   []Jump
}

func (s StatementContent) Q() string     { return s.Statement }
func (s StatementContent) Desc() string  { return s.Description }
func (s StatementContent) Logic() []Jump { return s.Jumps }

func (s StatementContent) label() string    { return "Statement" }
func (s StatementContent) prefix() string   { return "statement" }
func (s StatementContent) selector() string { return "content" }

type StatementContentOptions struct {
	RemoveIcon bool
}

func (f *form) AddStatementContent(q *StatementContent) FormElement {
	f.createQuestion(q)

	if q.ButtonText != "" {
		f.t.driverT.Q("#statement_button").SendKeys(q.ButtonText)
	}

	// Options
	if q.Options != nil {
		f.setOption(q, q.Options.RemoveIcon, "quoteMarksEnabled")
	}

	return f.saveQuestion(q, false)
}
