package common

const (
	StatusResourceUninitialized int = iota
	StatusResourceSetFinalizer
	StatusResourceCreated
	StatusResourceOrphaned
)
