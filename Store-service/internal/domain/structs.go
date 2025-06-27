package domain

type System struct {
	Num_CPU    int                `json:"num_cpu"`
	CPU_usage  map[string]float64 `json:"cpu_usage"`
	RAM        int64              `json:"ram"`
	RAM_used   int64              `json:"ram_used"`
	DISC       float64            `json:"disc"`
	DISC_used  float64            `json:"disc_used"`
	GOMAXPROCS int                `json:"gomaxprocs"`
}

type UserGroup string

const (
	AdminGroup    UserGroup = "admin"
	CustomerGroup UserGroup = "customer"
)

type User struct {
	ID         int       `json:"id"`
	LastName   string    `json:"last_name"`
	FirstName  string    `json:"first_name"`
	MiddleName string    `json:"middle_name,omitempty"`
	Group      UserGroup `json:"group"`
}
