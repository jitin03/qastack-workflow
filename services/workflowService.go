package services

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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
	url := "https://" + os.Getenv("ARGO_SERVER_ENDPOINT") + ":2746/api/v1/workflows/argo"
	method := "POST"
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	template, err := s.repo.RunWorkflow(id)
	if err != nil {
		logger.Info("err in run workflow")

		return "", errs.NewUnexpectedError("Unexpected from cluster")
	}

	r := strings.NewReader(template)
	client := &http.Client{}
	req, argoErr := http.NewRequest(method, url, r)

	res, argoErr := client.Do(req)
	if argoErr != nil {
		logger.Info("err in run workflow")
		return "", errs.NewUnexpectedError("Unexpected from cluster")
	}

	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(string(body))
	return string(body), nil

}

func (s DefaultWorkflowService) AddWorkflow(req dto.AddWorkflowRequest) (*dto.AddWorkflowResponse, *errs.AppError) {

	config := req.Config

	c := domain.Workflow{

		Name:       req.Name,
		Project_Id: req.Project_id,
		Created_By: req.Created_By,
		Config:     config,
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
