package models

type Employee struct {
	ID        int        `json:"id"`
	LibraryID int        `json:"library_id"`
	MemberID  int        `json:"member_id"`
	Position  string     `json:"position"`
	Salary    float64    `json:"salary"`
	Member    BookMember `json:"member,omitempty"` // Attached on JOINs
}
