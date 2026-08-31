package adzuna

type Item struct {
	ID           string `json:"id"`
	CreatedAt    string `json:"created_at"`
	Tittle       string `json:"tittle"`
	Location     string `json:"location"`
	MinSalary    int    `json:"min_salary"`
	MaxSalary    int    `json:"max_salary"`
	Company      string `json:"company"`
	ContractTime string `json:"contract_time"` //full-time|
	Description  string `json:"description"`
	RedirectURL  string `json:"redirect_url"`
}

type AdzunaItem struct {
	ID           string `json:"id"`
	Created      string `json:"created"`
	Title        string `json:"title"`
	SalaryMin    int    `json:"salary_min"`
	SalaryMax    int    `json:"salary_max"`
	Description  string `json:"description"`
	RedirectURL  string `json:"redirect_url"`
	ContractTime string `json:"contract_time"`

	Company struct {
		DisplayName string `json:"display_name"`
	} `json:"company"`

	Location struct {
		DisplayName string `json:"display_name"`
	} `json:"location"`
}
