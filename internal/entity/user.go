package entity

import "time"

type UserRole string

const (
	RoleAnon      UserRole = "anon"
	RoleUser      UserRole = "user"
	RoleDiffuseur UserRole = "diffuseur"
	RoleAdmin     UserRole = "admin"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsDiffuseur() bool {
	return u.Role == RoleDiffuseur || u.Role == RoleAdmin
}

func (u *User) HasRole(roles ...UserRole) bool {
	for _, r := range roles {
		if u.Role == r {
			return true
		}
	}
	return false
}
