package typeform

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Bowbaq/go-selenium"
)

type FormElement interface {
	Q() string
	Desc() string
	Logic() []Jump

	e() selenium.WebElementT
	setE(selenium.WebElementT)

	label() string
	prefix() string
	selector() string
}

type BaseQuestion struct {
	name    string
	element selenium.WebElementT
}

func (bq BaseQuestion) e() selenium.WebElementT {
	return bq.element
}

func (bq *BaseQuestion) setE(e selenium.WebElementT) {
	bq.element = e
}

type Jump struct {
	When string
	GoTo string
}

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

func (f *form) Apply(elements []FormElement) {
	byName := make(map[string]FormElement)

	// Validate
	for _, e := range elements {
		if _, exists := byName[e.Q()]; exists {
			log.Fatalln("Elements must have a unique question name. Duplicate: " + e.Q())
		}
		byName[e.Q()] = e
	}

	for _, e := range elements {
		for _, jump := range e.Logic() {
			//TODO: validate when value
			_, exists := byName[jump.GoTo]
			if !exists {
				log.Fatalf("Trying to jump to an unknown question %s from %s\n", jump.GoTo, e.Q())
			}
		}
	}

	// Create elements
	for _, e := range elements {
		switch e.(type) {
		case *MultipleChoiceQuestion:
			f.AddMultipleChoiceQuestion(e.(*MultipleChoiceQuestion))
		case *DropDownQuestion:
			f.AddDropDownQuestion(e.(*DropDownQuestion))
		case *YesNoQuestion:
			f.AddYesNoQuestion(e.(*YesNoQuestion))
		case *StatementContent:
			f.AddStatementContent(e.(*StatementContent))
		}

	}

	// Add logic jumps
	for _, e := range elements {
		f.HookJumps(e, byName)
	}
}

func (f *form) createQuestion(q FormElement) {
	log.Println("Adding question:", q.Q())
	f.t.ensureURL("https://admin.typeform.com/form/" + f.Id + "/fields")

	f.t.driverT.Q("li[data-label='" + q.label() + "']").Click()

	// Question
	f.t.waitFor("#" + q.prefix() + "_" + q.selector())
	f.t.driver.SwitchFrame(q.prefix() + "_" + q.selector() + "_ifr")
	f.t.driverT.Q("#tinymce").SendKeys(q.Q())
	f.t.driver.SwitchFrameParent()

	if q.Desc() != "" {
		// Description
		f.t.driverT.Q("#description div.coolCheckbox").Click()
		f.t.waitFor("#" + q.prefix() + "_description")
		f.t.driver.SwitchFrame(q.prefix() + "_description_ifr")
		f.t.driverT.Q("#tinymce").SendKeys(q.Desc())
		f.t.driver.SwitchFrameParent()
	}
}

func (f *form) setOption(q FormElement, enabled bool, suffix string) {
	if enabled {
		f.t.driverT.Q("#"+q.prefix()+"_"+suffix).FindElement(selenium.ByXPATH, ".//following-sibling::div").Click()
	}
}

func (f *form) saveQuestion(q FormElement, required bool) FormElement {
	f.t.driverT.Q("#submit").Click()

	question := q.Q()
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
			q.setE(label.FindElement(selenium.ByXPATH, ".//ancestor::li"))
		}
	}

	return q
}

func (f *form) HookJumps(from FormElement, byName map[string]FormElement) {
	from.e().MoveTo(0, 0)
	time.Sleep(100 * time.Millisecond)
	from.e().Q(".action-jump").Click()

	f.t.waitFor(".help-wrapper .proceed")
	proceed := f.t.driverT.Q(".help-wrapper .proceed")
	if proceed != nil && proceed.IsDisplayed() {
		proceed.Click()
	}

	time.Sleep(500 * time.Millisecond)

	for _, jump := range from.Logic() {
		to := byName[jump.GoTo]

		if jump.When != "DEFAULT" {
			add_jump, err := f.t.driver.Q(".admin-button.add-jump")
			if err != nil || !add_jump.T(f.t.logger).IsDisplayed() {
				add_jump, _ = f.t.driver.Q(".row.jump.sortable:nth-last-child(2) .jump-actions .add")
			}
			add_jump.Click()

			time.Sleep(200 * time.Millisecond)
			f.t.driverT.Q(".row.jump.sortable:nth-last-child(2) .cell.value").Click()

			time.Sleep(200 * time.Millisecond)
			for _, option := range f.t.driverT.QAll(".select2-result-label") {
				if option.Text() == jump.When {
					option.Click()
					break
				}
			}

			time.Sleep(200 * time.Millisecond)
			f.t.driverT.Q(".row.jump.sortable:nth-last-child(2) .cell.jumps-header").Click()

			time.Sleep(500 * time.Millisecond)
			for _, option := range f.t.driverT.QAll(".select2-result-label") {
				if strings.HasSuffix(option.Text(), to.Q()) {
					option.Click()
					break
				}
			}
		} else {
			f.t.driverT.Q(".row.jump.otherwise .cell.destination").Click()

			time.Sleep(200 * time.Millisecond)
			for _, option := range f.t.driverT.QAll(".select2-result-selectable") {
				if strings.HasSuffix(option.Text(), to.Q()) {
					option.Click()
					break
				}
			}
		}

		time.Sleep(200 * time.Millisecond)
	}

	f.t.driverT.Q(".admin-button.save").Click()

	time.Sleep(500 * time.Millisecond)
}
