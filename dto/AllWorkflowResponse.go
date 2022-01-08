package dto

type AllWorkflowResponse struct {
	Workflow_Id string `json:"Id"`
	Name        string `json:"workflow_name"`
	Project_Id  string `json:"project_Id"`
	Created_By  int    `json:"user_id"`
}
