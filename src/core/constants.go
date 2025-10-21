package core

type Domain string

const (
	DomainSystem  Domain = "system"
	DomainUser    Domain = "user"
	DomainGui     Domain = "gui"
	DomainUnknown Domain = "unknown"
)
