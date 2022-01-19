package dto

import "github.com/jmoiron/sqlx/types"

type AddWorkflowRequest struct {
	Project_id string         `json:"project_Id"`
	Name       string         `json:"name"`
	Config     types.JSONText `json:"config"`
	Created_By int            `json:"user_Id"`
}

type Parameter struct {
	Name  string `json:"name",omitempty"`
	Value string `json:"value",omitempty"`
}
