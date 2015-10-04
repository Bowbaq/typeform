# typeform
A library to automate form creation on typeform.com by way of selenium

```
package main

import  (
	"log"

	"github.com/Bowbaq/typeform"
)

func main() {
	typeform.SilentDriver(true)
	typeform.Timeout = 15 * time.Second

	// Requires chromedriver to work
	t := typeform.NewClient(typeform.ChromeDriver())
	defer t.Quit()

	t.Login("<username>", "<password>")

	form, err := t.Create("Automatic Form " + time.Now().Format("01/02 - 15:04:05"))
	if err != nil {
		log.Fatal(err)
	}

	form.AddMultipleChoiceQuestion(
		"Color",
		"Physical color of the product contained in the bottle.",
		[]string{"Red", "White", "Green", "Blue"},
		&typeform.MultipleChoiceQuestionOptions{
			Required: true,
			AddOtherOption: true,
		},
	)

	form.AddYesNoQuestion(
		"Is this awesome",
		"Answer yes if you think this is awesome",
		&typeform.YesNoQuestionOptions{
			Required: true,
		},
	)
}
```