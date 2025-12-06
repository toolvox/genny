package generator

import "fmt"

// ComponentNotFoundError indicates a component was referenced but doesn't exist
type ComponentNotFoundError struct {
	Name string
}

func (e *ComponentNotFoundError) Error() string {
	return fmt.Sprintf("component not found: %s", e.Name)
}

// DataPathError indicates an invalid or unreachable data path
type DataPathError struct {
	Path string
	Part string
}

func (e *DataPathError) Error() string {
	return fmt.Sprintf("invalid data path '%s' at part '%s'", e.Path, e.Part)
}

// TemplateParseError indicates a template could not be parsed
type TemplateParseError struct {
	Name   string
	Source string
	Err    error
}

func (e *TemplateParseError) Error() string {
	return fmt.Sprintf("failed to parse template '%s': %v", e.Name, e.Err)
}

func (e *TemplateParseError) Unwrap() error {
	return e.Err
}

// TemplateExecuteError indicates a template could not be executed
type TemplateExecuteError struct {
	Name string
	Err  error
}

func (e *TemplateExecuteError) Error() string {
	return fmt.Sprintf("failed to execute template '%s': %v", e.Name, e.Err)
}

func (e *TemplateExecuteError) Unwrap() error {
	return e.Err
}

// FileNotFoundError indicates a required file was not found
type FileNotFoundError struct {
	Path string
}

func (e *FileNotFoundError) Error() string {
	return fmt.Sprintf("file not found: %s", e.Path)
}

// ValidationError indicates invalid configuration or data
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
}
