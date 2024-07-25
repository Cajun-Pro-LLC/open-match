package main

import "strings"

type NestedError struct {
	errors []error
}

func (e *NestedError) Error() string {
	errors := make([]string, len(e.errors))
	for i, err := range e.errors {
		errors[i] = err.Error()
	}
	return strings.Join(errors, "\n")
}

func (e *NestedError) Add(err error) {
	e.errors = append(e.errors, err)
}

func (e *NestedError) Return() error {
	if len(e.errors) == 0 {
		return nil
	}
	return e
}
