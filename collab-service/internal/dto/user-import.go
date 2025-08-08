package dto

type ImportUserResult struct {
	Line    int    `json:"line"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ImportUserSummary struct {
	Total   int                `json:"total"`
	Success int                `json:"success"`
	Failed  int                `json:"failed"`
	Results []ImportUserResult `json:"results"`
}
