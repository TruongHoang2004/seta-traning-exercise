package dto

// CreateNoteDTO is the Data Transfer Object for notes
type CreateNoteDTO struct {
	Title    string `json:"title" binding:"required"`
	Body     string `json:"body"`
	FolderID string `json:"folder_id" binding:"required"`
}

type UpdateNoteDTO struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	FolderID string `json:"folder_id"`
}
