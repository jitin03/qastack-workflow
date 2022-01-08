package dto

type AddWorkflowRequest struct {
	Project_id string `json:"project_Id"`
	Name       string `json:"name"`
	Config     []struct {
		Name       string `json:"name"`
		Repository string `json:"repository"`
		Branch     string `json:"branch"`
		Git_Token  string `json:"token"`
	} `json:"config"`
	Created_By int `json:"user_Id"`
}
