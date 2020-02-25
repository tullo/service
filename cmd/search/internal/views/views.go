package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
)

// These constants represent the different layouts in use.
const (
	RESULTS = "Results"
	SEARCH  = "Search"
	LAYOUT  = "Layout"
)

// views contains a map of static templates for rendering views.
var views = make(map[string]*template.Template)

// Render generates the HTML response for this route.
func Render(results interface{}, host string) ([]byte, error) {
	fv := make(map[string]interface{})

	// Generate the markup for the results template.
	var markup bytes.Buffer
	if results != nil {
		vars := map[string]interface{}{"Items": results, "HOST": host}
		if err := views[RESULTS].Execute(&markup, vars); err != nil {
			return nil, fmt.Errorf("error processing results template : %v", err)
		}
		fv[RESULTS] = template.HTML(markup.String())
	}

	// Generate the markup for the search template.
	markup.Reset()
	if err := views[SEARCH].Execute(&markup, fv); err != nil {
		return nil, fmt.Errorf("error processing search template : %v", err)
	}

	// Generate the final markup with the layout template.
	vars := map[string]interface{}{"LayoutContent": template.HTML(markup.String())}
	markup.Reset()
	if err := views[LAYOUT].Execute(&markup, vars); err != nil {
		return nil, fmt.Errorf("error processing layout template : %v", err)
	}

	return markup.Bytes(), nil
}

// Init loads the existing templates for use to generate views.
func Init() error {

	// In order for the endpoint tests to run this needs to be
	// physically located. Trying to avoid configuration for now.
	pwd, _ := os.Getwd()
	if "/app" != pwd {
		pwd += "/cmd/search"
	}
	if err := loadTemplate(LAYOUT, pwd+"/internal/views/basic_layout.html"); err != nil {
		return err
	}

	if err := loadTemplate(SEARCH, pwd+"/internal/views/search.html"); err != nil {
		return err
	}

	if err := loadTemplate(RESULTS, pwd+"/internal/views/results.html"); err != nil {
		return err
	}
	return nil
}

// loadTemplate reads the specified template file for use.
func loadTemplate(name string, path string) error {

	// Read the html template file.
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// Create a template value for this code.
	tmpl, err := template.New(name).Parse(string(data))
	if name == RESULTS {
		tmpl, err = template.New(name).Funcs(template.FuncMap{
			"incr": func(idx int) int {
				return idx + 1
			},
		}).Parse(string(data))
	}
	if err != nil {
		return err
	}

	// Have we processed this file already?
	if _, exists := views[name]; exists {
		return fmt.Errorf("template %s already in use", name)
	}

	// Store the template for use.
	views[name] = tmpl
	return nil
}
