package dto

import "github.com/jmoiron/sqlx/types"

type AllWorkflowResponse struct {
	Workflow_Id string         `json:"Id"`
	Name        string         `json:"workflow_name"`
	Project_Id  string         `json:"project_Id"`
	Created_By  int            `json:"user_id"`
	Config      types.JSONText `db:"config"`
}
