package services

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"qastack-workflows/domain"
	"qastack-workflows/dto"
	"qastack-workflows/errs"
	logger "qastack-workflows/loggers"
	"strings"
)

type WorkflowServices interface {
	AddWorkflow(request dto.AddWorkflowRequest) (*dto.AddWorkflowResponse, *errs.AppError)
	AllWorkflows(projectKey string, pageId int) ([]dto.AllWorkflowResponse, *errs.AppError)
	RunWorkflow(string) (string, *errs.AppError)
}

type DefaultWorkflowService struct {
	repo domain.WorkflowRepository
}

func (s DefaultWorkflowService) AllWorkflows(componentId string, pageId int) ([]dto.AllWorkflowResponse, *errs.AppError) {

	workflows, err := s.repo.AllWorkflows(componentId, pageId)
	if err != nil {
		return nil, err
	}
	response := make([]dto.AllWorkflowResponse, 0)
	for _, workflow := range workflows {
		response = append(response, workflow.ToDto())
	}
	return response, err
}

func (s DefaultWorkflowService) RunWorkflow(id string) (string, *errs.AppError) {
	// api/v1/workflows/argo
	url := "https://af8df61718f35408fbaf3f28bdec0b1d-1414162024.us-east-1.elb.amazonaws.com:2746/api/v1/workflows/argo"
	method := "POST"
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	template, err := s.repo.RunWorkflow(id)

	r := strings.NewReader(template)
	client := &http.Client{}
	req, _ := http.NewRequest(method, url, r)

	if err != nil {
		logger.Info("err in run workflow")

	}
	res, _ := client.Do(req)

	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(string(body))
	return string(body), nil

}

func (s DefaultWorkflowService) AddWorkflow(req dto.AddWorkflowRequest) (*dto.AddWorkflowResponse, *errs.AppError) {

	c := domain.Workflow{

		Name:       req.Name,
		Project_Id: req.Project_id,
		Created_By: req.Created_By,
		Config: []struct {
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
		}(req.Config),
	}

	if newComponent, err := s.repo.AddWorkflow(c); err != nil {
		return nil, err
	} else {
		return newComponent.ToAddWorkflowResponseDto(), nil
	}

}

func NewWorkflowService(repository domain.WorkflowRepository) DefaultWorkflowService {
	return DefaultWorkflowService{repository}
}
