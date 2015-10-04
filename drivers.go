package typeform

import (
	"log"
	"os/exec"
	"time"

	"sourcegraph.com/sourcegraph/go-selenium"
)

// Wait up to DriverTimeout seconds before the driver is ready to accept requests
var DriverTimeout = 5 * time.Second

func ChromeDriver(args ...string) selenium.WebDriver {
	start := time.Now()
	cmd := exec.Command("chromedriver")
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	capabilities := selenium.Capabilities(map[string]interface{}{
		"browserName": "chrome",
		"chromeOptions": map[string]interface{}{
			"args": args,
		},
	})

	driver, err := selenium.NewRemote(capabilities, "http://localhost:9515")
	if err != nil {
		log.Fatalf("Failed to create remote: %s\n", err)
		return nil
	}

	_, err = driver.Status()
	for err != nil && time.Since(start) < DriverTimeout {
		log.Println("Waiting for chromedriver ...")
		time.Sleep(500 * time.Millisecond)
		_, err = driver.Status()
	}

	if err != nil {
		log.Fatalf("Failed to open session: %s\n", err)
		return nil
	}

	return driver
}

func SilentDriver(silent bool) {
	if silent {
		selenium.Log = nil
	}
}
