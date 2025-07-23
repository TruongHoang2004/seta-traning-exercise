package dto

// NoteDTO is the Data Transfer Object for notes
type NoteDTO struct {
	Title    string `json:"title" binding:"required"`
	Body     string `json:"body" binding:"required"`
	FolderID string `json:"folder_id" binding:"required"`
}
