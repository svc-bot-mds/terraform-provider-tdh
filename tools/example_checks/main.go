package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type checkResponse struct {
	err     error
	purpose string
}

type checkValidator func(chan checkResponse)

// just add the checks here
var checks = []checkValidator{
	validateResourceExamples,
	validateDatasourceExamples,
}

func main() {
	totalChecks := len(checks)
	ch := make(chan checkResponse, totalChecks)
	// cleanup channel
	defer close(ch)

	// trigger checks
	for _, check := range checks {
		go check(ch)
	}

	// wait as they come, fail early
	for ; totalChecks != 0; totalChecks-- {
		select {
		case response := <-ch:
			if response.err != nil {
				fmt.Printf("❌  Check %q failed with: %v\n", response.purpose, response.err)
				os.Exit(1)
			}
			fmt.Printf("✅  Check %q successful\n", response.purpose)
		}
	}

	fmt.Println("All Checks completed!")
}

func validateResourceExamples(ch chan checkResponse) {
	response := checkResponse{
		purpose: "Validate resource examples",
	}

	// main logic
	matches, _ := filepath.Glob("tdh/resource_*.go")
	for _, p := range matches {
		re := regexp.MustCompile(`tdh/resource_(\w*)\.go`)
		match := re.FindStringSubmatch(p)
		if len(match) <= 1 {
			continue
		}
		resource := match[1]
		examplePath := fmt.Sprintf("examples/resources/tdh_%s", resource)
		_, err := os.ReadDir(examplePath)
		if err != nil {
			response.err = fmt.Errorf("missing example at path %q", examplePath)
			break
		}
		if err = checkExampleFileAtPath(examplePath, "resource.tf"); err != nil {
			response.err = err
			break
		}
		if err = checkExampleFileAtPath(examplePath, "import.sh"); err != nil {
			response.err = err
			break
		}
	}
	ch <- response
}

func validateDatasourceExamples(ch chan checkResponse) {
	response := checkResponse{
		purpose: "Validate datasource examples",
	}

	// main logic
	matches, _ := filepath.Glob("tdh/data_source_*.go")
	for _, p := range matches {
		re := regexp.MustCompile(`tdh/data_source_(\w*)\.go`)
		match := re.FindStringSubmatch(p)
		if len(match) <= 1 {
			continue
		}
		resource := match[1]
		examplePath := fmt.Sprintf("examples/data-sources/tdh_%s", resource)
		_, err := os.ReadDir(examplePath)
		if err != nil {
			response.err = fmt.Errorf("missing example at path %q", examplePath)
			break
		}
		if err = checkExampleFileAtPath(examplePath, "data-source.tf"); err != nil {
			response.err = err
			break
		}
	}
	ch <- response
}

func checkExampleFileAtPath(path string, file string) error {
	matches, err := filepath.Glob(fmt.Sprintf("%s/%s", path, file))
	if err != nil {
		return fmt.Errorf("unexpected error checking for %q: %q", file, err.Error())
	}
	filePresent := len(matches) > 0
	if !filePresent {
		return fmt.Errorf("missing example %q in %q", file, path)
	}
	return nil
}
