package entity

type AdminStats struct {
	TotalUsers int            `json:"total_users"`
	ByRole     map[string]int `json:"by_role"`
}
