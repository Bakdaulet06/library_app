package models

type Role string

const (
	RoleClient   Role = "client"
	RoleAdmin    Role = "admin"
	RoleEmployee Role = "employee"
)

const (
	EmpSalaryMin float64 = 5000
	EmpSalaryMax float64 = 20000
)

const (
	LibrarianSalary float64 = 10000
	SanitarSalary   float64 = 8000
	AssistantSalary float64 = 6000
)

type Position string

const (
	Librarian Position = "Librarian"
	Sanitar   Position = "Sanitar"
	Assistant Position = "Assistant"
)

const (
	MaxLibrarians = 1
	MaxAssistants = 2
	MaxSanitars   = 2
)

func (p Position) IsValid() bool {
	switch p {
	case Librarian, Sanitar, Assistant:
		return true
	default:
		return false
	}
}

func (p Position) DefaultSalary() float64 {
	switch p {
	case Librarian:
		return LibrarianSalary
	case Sanitar:
		return SanitarSalary
	case Assistant:
		return AssistantSalary
	default:
		return EmpSalaryMin
	}
}
