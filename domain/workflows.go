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
		Name         string   `db:"name"`
		Repository   string   `db:"repository"`
		Branch       string   `db:"branch"`
		Git_Token    string   `db:"token"`
		DockerImage  string   `db:"docker_image"`
		EntryPath    []string `db:"entrypath"`
		Dependencies []string `db:"dependencies"`
		Parameters   []struct {
			Name  string `db:"name"`
			Value string `db:"value"`
		}
		Source string `db:"input_command"`
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
	RunWorkflow(workflowId string) (string, *errs.AppError)
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
	GenerateName string `json:"generateName"`
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
