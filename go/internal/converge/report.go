package converge

import (
	"errors"
	"fmt"
	"io"
)

type Status string

const (
	Completed  Status = "completed"
	Failed     Status = "failed"
	Unresolved Status = "unresolved"
)

type Operation struct {
	Scope   string
	Action  string
	Subject string
	Status  Status
	Err     error
}

type Report struct{ Operations []Operation }

func (r *Report) Add(operation Operation) { r.Operations = append(r.Operations, operation) }

func (r Report) HasFailures() bool {
	for _, operation := range r.Operations {
		if operation.Status == Failed {
			return true
		}
	}
	return false
}

func (r Report) Error() error {
	if !r.HasFailures() {
		return nil
	}
	return &ReportError{Report: r}
}

func (r Report) Print(writer io.Writer) {
	counts := map[Status]int{}
	for _, operation := range r.Operations {
		counts[operation.Status]++
		if operation.Status == Completed {
			continue
		}
		message := fmt.Sprintf("[%s] %s %s", operation.Status, operation.Action, operation.Subject)
		if operation.Scope != "" {
			message = fmt.Sprintf("[%s] %s", operation.Scope, message)
		}
		if operation.Err != nil {
			message += ": " + operation.Err.Error()
		}
		fmt.Fprintln(writer, message)
	}
	fmt.Fprintf(writer, "[operations] completed=%d unresolved=%d failed=%d\n", counts[Completed], counts[Unresolved], counts[Failed])
}

type ReportError struct{ Report Report }

func (e *ReportError) Error() string { return "runtime convergence incomplete" }
func (e *ReportError) Unwrap() error {
	var failures []error
	for _, operation := range e.Report.Operations {
		if operation.Status == Failed && operation.Err != nil {
			failures = append(failures, operation.Err)
		}
	}
	return errors.Join(failures...)
}
