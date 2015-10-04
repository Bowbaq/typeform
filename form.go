package typeform

import (
	"log"
	"net/http"
	"time"

	"github.com/Bowbaq/go-selenium"
)

type QuestionType int

const (
	MultipleChoice QuestionType = iota
	YesNo
	Statement
)

type form struct {
	t *Typeform

	Id   string
	Name string
}

func (f *form) Open() {
	f.t.driver.Get("https://admin.typeform.com/form/" + f.Id + "/fields")
}

func (f *form) Delete() {
	cookies, err := f.t.driver.GetCookies()
	if err != nil {
		log.Fatal(err)
	}

	req, _ := http.NewRequest("POST", "https://admin.typeform.com/form/delete/"+f.Id, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Host", "admin.typeform.com")
	req.Header.Add("Origin", "https://admin.typeform.com")
	for _, c := range cookies {
		req.AddCookie(&http.Cookie{
			Name:  c.Name,
			Value: c.Value,
		})
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
}

type MultipleChoiceQuestionOptions struct {
	Required               bool
	AllowMutipleSelections bool
	Randomize              bool
	ForceVerticalAlignment bool
	AddOtherOption         bool
}

func (f *form) AddMultipleChoiceQuestion(question, desc string, choices []string, opt *MultipleChoiceQuestionOptions) int {
	f.addQuestion(MultipleChoice, question, desc)

	// Choices
	for _, c := range choices {
		choice := f.t.driverT.Q(".choices .choice:last-child")
		choice.Q("input[type='text']").SendKeys(c)
		choice.Q(".add").Click()

		time.Sleep(300 * time.Millisecond)
	}
	f.t.driverT.Q(".choices .choice:last-child .remove").Click()

	// Options
	f.setOption(MultipleChoice, opt.Required, "required")
	f.setOption(MultipleChoice, opt.AllowMutipleSelections, "multiple")
	f.setOption(MultipleChoice, opt.Randomize, "randomize")
	f.setOption(MultipleChoice, opt.ForceVerticalAlignment, "vertical")
	f.setOption(MultipleChoice, opt.AddOtherOption, "other")

	return f.saveQuestion(question, opt.Required)
}

type YesNoQuestionOptions struct {
	Required bool
}

func (f *form) AddYesNoQuestion(question, desc string, opt *YesNoQuestionOptions) int {
	f.addQuestion(YesNo, question, desc)

	// Options
	f.setOption(YesNo, opt.Required, "required")

	return f.saveQuestion(question, opt.Required)
}

type StatementOptions struct {
	RemoveIcon bool
}

func (f *form) AddStatement(statement, desc, buttonText string, opt *StatementOptions) int {
	f.addQuestion(Statement, statement, desc)

	if buttonText != "" {
		f.t.driverT.Q("#statement_button").SendKeys(buttonText)
	}

	// Options
	f.setOption(Statement, opt.RemoveIcon, "quoteMarksEnabled")

	return f.saveQuestion(statement, false)
}

type question struct {
	label  string
	prefix string
}

var questions = map[QuestionType]question{
	MultipleChoice: {"Multiple choice", "list"},
	YesNo:          {"Yes / No", "yes_no"},
	Statement:      {"Statement", "statement"},
}

func (f *form) addQuestion(typ QuestionType, question, desc string) {
	log.Println("Adding question:", question)
	f.t.ensureURL("https://admin.typeform.com/form/" + f.Id + "/fields")

	q := questions[typ]
	qSelector := "question"
	if typ == Statement {
		qSelector = "content"
	}

	f.t.driverT.Q("li[data-label='" + q.label + "']").Click()

	// Question
	f.t.waitFor("#" + q.prefix + "_" + qSelector)
	f.t.driver.SwitchFrame(q.prefix + "_" + qSelector + "_ifr")
	f.t.driverT.Q("#tinymce").SendKeys(question)
	f.t.driver.SwitchFrameParent()

	if desc != "" {
		// Description
		f.t.driverT.Q("#description div.coolCheckbox").Click()
		f.t.waitFor("#" + q.prefix + "_description")
		f.t.driver.SwitchFrame(q.prefix + "_description_ifr")
		f.t.driverT.Q("#tinymce").SendKeys(desc)
		f.t.driver.SwitchFrameParent()
	}
}

func (f *form) setOption(typ QuestionType, enabled bool, suffix string) {
	if enabled {
		f.t.driverT.Q("#"+questions[typ].prefix+"_"+suffix).FindElement(selenium.ByXPATH, ".//following-sibling::div").Click()
	}
}

func (f *form) saveQuestion(question string, required bool) int {
	f.t.driverT.Q("#submit").Click()

	if required {
		question = "* " + question
	}

	// Wait for popup to disappear
	start := time.Now()

	popup := f.t.driverT.QAll("div.build.popup")
	for len(popup) > 0 && time.Since(start) < Timeout {
		time.Sleep(500 * time.Millisecond)
		popup = f.t.driverT.QAll("div.build.popup")
	}
	if len(popup) > 0 {
		log.Fatalf("Failed to save question %s (after %v)\n", question, Timeout)
	}
	time.Sleep(500 * time.Millisecond)

	// Find question
	for _, label := range f.t.driverT.QAll("li.field span.label") {
		if label.Text() == question {
			// TODO: grab actual question #
			return 0
		}
	}

	return 0
}
