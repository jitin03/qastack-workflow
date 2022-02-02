package dto

type AddWorkflowResponse struct {
	Workflow_Id string `db:"workflow_Id"`
}

type ReSubmitRunWorkflowResponse struct {
	Workflow_run_name string `json:"workflow_run_name"`
}
