package entities

// CardCraftAiGeneratedContent represents the response structure from the Python API
type CardCraftAiGeneratedContent struct {
	Title       string            `json:"title"`
	Attributes  map[string]string `json:"attributes"`
	Description string            `json:"description"`
	ParentID    *int32            `json:"parent_id"`
	ParentName  *string           `json:"parent_name"`
	SubjectID   *int32            `json:"subject_id"`
	SubjectName *string           `json:"subject_name"`
	TypeID      *int32            `json:"type_id"`
	TypeName    *string           `json:"type_name"`
	RootID      *int32            `json:"root_id"`
	RootName    *string           `json:"root_name"`
	SubID       *int32            `json:"sub_id"`
	SubName     *string           `json:"sub_name"`
	SessionID   string            `json:"session_id"`
}
