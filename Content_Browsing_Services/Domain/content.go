package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Content struct {
	ID          string             `json:"id" bson:"_id,omitempty"`
	GroupID     primitive.ObjectID `bson:"group_id,omitempty"`
	GroupName   string             `json:"group_name" bson:"group_name"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	URL         string             `json:"url" bson:"url"`
	Language    string             `json:"language" bson:"language"`
}

type Group struct {
	ID        string `json:"group_id" bson:"group_id,omitempty"`
	GroupName string `json:"group_name" bson:"group_name"`
}

type PaginatedGroupResponse struct {
	Group       []Group `json:"group"`
	TotalItems  int     `json:"total_items"`
	TotalPages  int     `json:"total_pages"`
	CurrentPage int     `json:"current_page"`
	PageSize    int     `json:"page_size"`
}

type PaginatedContentResponse struct {
	Contents    []Content `json:"contents"`
	TotalItems  int       `json:"total_items"`
	TotalPages  int       `json:"total_pages"`
	CurrentPage int       `json:"current_page"`
	PageSize    int       `json:"page_size"`
}