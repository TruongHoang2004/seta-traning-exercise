package dto

type FolderDTO struct {
	Name string `json:"name" binding:"required"`
}
