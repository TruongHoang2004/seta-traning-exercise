package dto

// dto/team.go
type CreateTeamInput struct {
	TeamName string         `json:"teamName" binding:"required"`
	Managers []ManagerInput `json:"managers"`
	Members  []MemberInput  `json:"members"`
}

type ManagerInput struct {
	ManagerID   string `json:"managerId" binding:"required"`
	ManagerName string `json:"managerName" binding:"required"`
}

type MemberInput struct {
	MemberID   string `json:"memberId" binding:"required"`
	MemberName string `json:"memberName" binding:"required"`
}

type AddMemberToTeamInput struct {
	UserID string `json:"userId" binding:"required"`
	TeamID string `json:"teamId" binding:"required"`
}

type AddManagerToTeamInput struct {
	UserID string `json:"userId" binding:"required"`
	TeamID string `json:"teamId" binding:"required"`
}
