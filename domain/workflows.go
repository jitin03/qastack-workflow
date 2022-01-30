package domain

import (
	"qastack-workflows/dto"
	"qastack-workflows/errs"

	"github.com/jmoiron/sqlx/types"
)

type Workflow struct {
	Workflow_Id string         `db:"id"`
	Name        string         `db:"workflowname"`
	Project_Id  string         `db:"project_id"`
	Created_By  int            `db:"created_by"`
	Config      types.JSONText `db:"config"`
	CreatedDate string         `db:"created_date"`
}

type Config struct {
	Name       string   `json:"name"`
	Git_Token  string   `json:"token"`
	Branch     string   `json:"branch"`
	EntryPath  []string `json:"entrypath"`
	Parameters []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	Repository   string   `json:"repository"`
	Dependencies []string `json:"dependencies"`
	DockerImage  string   `json:"docker_image"`
	Source       string   `json:"input_command"`
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
	RunWorkflow(workflowId string) (string, *errs.AppError)
	DeleteWorkflow(id string) *errs.AppError
	GetWorkflowDetail(string) (*Workflow, *errs.AppError)
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
		Config:      t.Config,
	}
}

type GenerateTemplate struct {
	WorkflowTemplate WorkflowTemplate `json:"workflow"`
}

type WorkflowTemplate struct {
	APIVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Metadata   Metadata `json:"metadata"`
	Spec       Spec     `json:"spec"`
}

type Metadata struct {
	GenerateName string `json:"name"`
}

type Spec struct {
	Entrypoint string      `json:"entrypoint"`
	Templates  []Templates `json:"templates"`
	Arguments  *Arguments  `json:"arguments",omitempty"`
}

type Arguments struct {
	Parameters []Parameters `json:"parameters"`
}

type Parameters struct {
	Name  string `json:"name",omitempty"`
	Value string `json:"value",omitempty"`
}
type Script struct {
	Image   string   `json:"image"`
	Command []string `json:"command"`
	Source  string   `json:"source"`
}

type Templates struct {
	Name   string  `json:"name"`
	Dag    *Dag    `json:"dag,omitempty"`
	Inputs *Inputs `json:"inputs"`
	Script *Script `json:"script,omitempty"`
}

type Inputs struct {
	Parameters []Parameters `json:"parameters"`
}
type Dag struct {
	Tasks []Tasks `json:"tasks"`
}

type Tasks struct {
	Name         string     `json:"name"`
	Arguments    *Arguments `json:"arguments",omitempty`
	Template     string     `json:"template"`
	Dependencies []string   `json:"dependencies,omitempty"`
}
