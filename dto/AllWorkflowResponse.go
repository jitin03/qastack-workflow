package dto

import "github.com/jmoiron/sqlx/types"

type AllWorkflowResponse struct {
	Workflow_Id       string         `json:"Id"`
	Name              string         `json:"workflow_name"`
	Workflow_Run_Name string         `json:"workflow_run_name"`
	Project_Id        string         `json:"project_Id"`
	WorkflowStatus    string         `json:"workflow_status"`
	Username          string         `json:"username"`
	Config            types.JSONText `json:",omitempty"`
	NodeStatus        types.JSONText `json:"node_status"`
	UpdatedDate       string         `json:"updated_date"`
	CreatedDate       string         `json:"created_date"`
	LastExecutedDate  string         `json:"last_execution_date"`
}
