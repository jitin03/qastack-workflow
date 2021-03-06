package dto

import "github.com/jmoiron/sqlx/types"

type AddWorkflowRequest struct {
	Project_id string         `json:"project_Id"`
	Name       string         `json:"name"`
	Config     types.JSONText `json:"config"`
	Created_By int            `json:"user_Id"`
}

type UpdateWorkflowRequest struct {
	Project_id string         `json:"project_Id"`
	Name       string         `json:"name"`
	Config     types.JSONText `json:"config"`
	Updated_By int            `json:"user_Id"`
}

type Parameter struct {
	Name  string `json:"name",omitempty"`
	Value string `json:"value",omitempty"`
}

type UpdateWorkflowStatus struct {
	WorkflowId   string         `json:"id"`
	Status       string         `json:"status"`
	UserId       string         `json:"user_Id"`
	WorkflowName string         `json:"workflow_name"`
	NodeStatus   types.JSONText `json:"node_status"`
}
