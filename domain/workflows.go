package domain

import (
	"qastack-workflows/dto"
	"qastack-workflows/errs"
)

type Workflow struct {
	Workflow_Id string `db:"id"`
	Name        string `db:"name"`
	Project_Id  string `db:"project_id"`
	Created_By  int    `db:"created_by"`
	Config      []struct {
		Name       string `db:"name"`
		Repository string `db:"repository"`
		Branch     string `db:"branch"`
		Git_Token  string `db:"token"`
	} `db:"config"`
}

type Steps struct {
	Steps []Step_Config `json:"steps"`
}

type Step_Config struct {
	StepId     int    `json:"id"`
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Git_Token  string `json:"token"`
}

type WorkflowRepository interface {
	AddWorkflow(workflow Workflow) (*Workflow, *errs.AppError)
	AllWorkflows(projectKey string, pageId int) ([]Workflow, *errs.AppError)
}

func (w Workflow) ToAddWorkflowResponseDto() *dto.AddWorkflowResponse {
	return &dto.AddWorkflowResponse{w.Workflow_Id}
}
func (t Workflow) ToDto() dto.AllWorkflowResponse {
	return dto.AllWorkflowResponse{
		Workflow_Id: t.Workflow_Id,
		Name:        t.Name,
		Project_Id:  t.Project_Id,
		Created_By:  t.Created_By,
	}
}
