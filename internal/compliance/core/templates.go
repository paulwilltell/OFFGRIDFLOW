package core

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

// Template represents a compliance report template.
type Template struct {
	Name string
	Body string
}

// TemplateRegistry manages compliance report templates.
type TemplateRegistry struct {
	templates map[string]*Template
}

// NewTemplateRegistry creates a new template registry.
func NewTemplateRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		templates: make(map[string]*Template),
	}
}

// Register adds a template to the registry.
func (r *TemplateRegistry) Register(name, body string) {
	r.templates[name] = &Template{
		Name: name,
		Body: body,
	}
}

// Get retrieves a template by name.
func (r *TemplateRegistry) Get(name string) (*Template, error) {
	tmpl, ok := r.templates[name]
	if !ok {
		return nil, fmt.Errorf("template %q not found", name)
	}
	return tmpl, nil
}

// Render executes a template with the provided data.
func (t *Template) Render(data interface{}) (string, error) {
	tmpl, err := template.New(t.Name).Parse(t.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %q: %w", t.Name, err)
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %q: %w", t.Name, err)
	}
	
	return buf.String(), nil
}

// RenderHTML executes a template and returns HTML-safe output.
func (t *Template) RenderHTML(data interface{}) (string, error) {
	tmpl, err := template.New(t.Name).Parse(t.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %q: %w", t.Name, err)
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %q: %w", t.Name, err)
	}
	
	return buf.String(), nil
}

// ValidateFields checks if required fields are present in the data.
func ValidateFields(data map[string]interface{}, required []string) error {
	var missing []string
	
	for _, field := range required {
		if _, ok := data[field]; !ok {
			missing = append(missing, field)
		}
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	
	return nil
}
