package rest

import "fmt"

type Diagnostic struct {
	Status          int
	Message         string
	HostPathMethod  string
	ProceedFallback bool
	IsFatal         bool
}

func (d *Diagnostic) Error() string {
	return fmt.Sprintf(`status: %d
message: %s
hostPathMethod: %s
isRetryAble: %v`, d.Status, d.Message, d.HostPathMethod, d.ProceedFallback)
}

func (d *Diagnostic) WithMessage(m string) *Diagnostic {
	d.Message = m
	return d
}

func (d *Diagnostic) WithStatus(s int) *Diagnostic {
	d.Status = s
	return d
}

func (d *Diagnostic) WithProceedFallback(v bool) *Diagnostic {
	d.ProceedFallback = v
	return d
}

func (d *Diagnostic) WithIsFatal(v bool) *Diagnostic {
	d.IsFatal = v
	return d
}
