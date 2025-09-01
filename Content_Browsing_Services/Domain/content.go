package domain

type Content struct {
	ID          string `json:"id" bson:"_id,omitempty"`
	GroupName   string `json:"group_name" bson:"group_name"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	URL         string `json:"url" bson:"url"`
	Language    string `json:"language" bson:"language"`
}

type PaginatedContentResponse struct {
	Items       []Content `json:"items"`
	TotalItems  int       `json:"total_items"`
	TotalPages  int       `json:"total_pages"`
	CurrentPage int       `json:"current_page"`
	PageSize    int       `json:"page_size"`
}