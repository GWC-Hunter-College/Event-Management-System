package models

type Student struct {
	ID    string `json:"id" db:"id"`       // Cognito sub or UUID
	Email string `json:"email" db:"email"` // Student email
}

// type StudentInfo struct {
// 	StudentID string  `json:"studentId" db:"fk_student_id"`
// 	Username  *string `json:"username,omitempty" db:"username"`
// 	FirstName *string `json:"firstName,omitempty" db:"first_name"`
// 	LastName  *string `json:"lastName,omitempty" db:"last_name"`
// }

// type Admin struct {
// 	StudentID string `json:"studentId" db:"fk_student_id"`
// }
