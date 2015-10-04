package typeform

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"sourcegraph.com/sourcegraph/go-selenium"
)

var Timeout = 5 * time.Second

type Typeform struct {
	logger *log.Logger

	driver  selenium.WebDriver
	driverT selenium.WebDriverT
}

func NewClient(driver selenium.WebDriver) *Typeform {
	tf := &Typeform{
		logger: log.New(os.Stderr, "", log.LstdFlags),

		driver: driver,
	}
	tf.driverT = driver.T(tf.logger)

	return tf
}

func (t *Typeform) Login(username, password string) {
	t.driver.Get("https://admin.typeform.com/login/")

	t.driverT.Q("#_username").SendKeys(username)
	t.driverT.Q("#_password").SendKeys(password)
	t.driverT.Q("#btnlogin").Click()

	t.ensureURL("https://admin.typeform.com/workspaces/")
}

func (t *Typeform) Open(name string) (*form, error) {
	for _, form := range t.ListForms() {
		if form.Name == name {
			form.Open()
			return &form, nil
		}
	}

	return nil, errors.New("Form does not exist: " + name)
}

func (t *Typeform) Create(name string) (*form, error) {
	t.assertURL("https://admin.typeform.com/workspaces")

	t.driverT.Q(".admin-button.add").Click()

	time.Sleep(200 * time.Millisecond)
	t.driverT.Q(".admin-button.upper").Click()

	t.driverT.Q("#add-form #quickyform_name").SendKeys(name)
	t.driverT.Q("#add-form input[type=submit]").Click()

	t.waitFor("li[data-label='Multiple choice']")

	return &form{
		t: t,

		Id:   strings.TrimSuffix(strings.TrimPrefix(t.driverT.CurrentURL(), "https://admin.typeform.com/form/"), "/fields/"),
		Name: name,
	}, nil
}

func (t *Typeform) OpenOrCreate(name string) (*form, error) {
	if form, err := t.Open(name); err == nil {
		return form, nil
	}

	return t.Create(name)
}

func (t *Typeform) ListForms() []form {
	var forms []form
	t.assertURL("https://admin.typeform.com/workspaces/")

	cards := t.driverT.QAll(".form-name")
	for _, card := range cards {
		forms = append(forms, form{
			t: t,

			Id:   strings.TrimPrefix(card.FindElement(selenium.ByXPATH, ".//ancestor::li").GetAttribute("id"), "form-"),
			Name: card.Text(),
		})
	}

	return forms
}

func (t *Typeform) Quit() {
	t.driverT.Quit()
}

func (t *Typeform) assertURL(prefix string) {
	currentURL := t.driverT.CurrentURL()

	if !strings.HasPrefix(currentURL, prefix) {
		t.logger.Fatalf("Expected page %s, got %s\n", prefix, currentURL)
	}
}

func (t *Typeform) ensureURL(prefix string) {
	start := time.Now()
	currentURL := t.driverT.CurrentURL()

	for !strings.HasPrefix(currentURL, prefix) && time.Since(start) < Timeout {
		time.Sleep(500 * time.Millisecond)
	}

	if !strings.HasPrefix(currentURL, prefix) {
		t.logger.Fatalf("Expected page %s, got %s (after %v)\n", prefix, currentURL, Timeout)
	}
}

func (t *Typeform) waitFor(selector string) {
	start := time.Now()

	_, err := t.driver.Q(selector)
	for err != nil && time.Since(start) < Timeout {
		time.Sleep(500 * time.Millisecond)
		_, err = t.driver.Q(selector)
	}

	if err != nil {
		log.Fatalf("Couldn't find element %s (after %v)\n", selector, Timeout)
	}
}
