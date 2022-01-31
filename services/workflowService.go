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
	"time"

	log "github.com/sirupsen/logrus"
)

const dbTSLayout = "2006-01-02 15:04:05"

type WorkflowServices interface {
	AddWorkflow(request dto.AddWorkflowRequest) (*dto.AddWorkflowResponse, *errs.AppError)
	AllWorkflows(projectKey string, pageId int) ([]dto.AllWorkflowResponse, *errs.AppError)
	RunWorkflow(string, userId string) *errs.AppError
	RetryRunWorkflow(string) *errs.AppError
	DeleteWorkflow(id string) *errs.AppError
	GetWorkflowDetail(string) (*dto.AllWorkflowResponse, *errs.AppError)
}

type DefaultWorkflowService struct {
	repo domain.WorkflowRepository
}

func (s DefaultWorkflowService) GetWorkflowDetail(workflowName string) (*dto.AllWorkflowResponse, *errs.AppError) {

	workflow, err := s.repo.GetWorkflowDetail(workflowName)
	if err != nil {
		return nil, err
	}

	response := workflow.ToDto()
	return &response, err
}

func (s DefaultWorkflowService) AllWorkflows(projectKey string, pageId int) ([]dto.AllWorkflowResponse, *errs.AppError) {

	workflows, err := s.repo.AllWorkflows(projectKey, pageId)
	if err != nil {
		return nil, err
	}
	response := make([]dto.AllWorkflowResponse, 0)
	for _, workflow := range workflows {
		response = append(response, workflow.ToDto())
	}
	return response, err
}

func (s DefaultWorkflowService) DeleteWorkflow(id string) *errs.AppError {

	err := s.repo.DeleteWorkflow(id)
	if err != nil {
		return errs.NewUnexpectedError("Unexpected error in delete action")
	}

	return nil
}

func (s DefaultWorkflowService) RunWorkflow(id string, userId string) *errs.AppError {

	url := "https://" + os.Getenv("ARGO_SERVER_ENDPOINT") + ":2746/api/v1/workflows/argo"
	method := "POST"
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	template, err := s.repo.RunWorkflow(id, userId)
	if err != nil {
		logger.Info("err in run workflow")

		return errs.NewUnexpectedError("Unexpected from cluster")
	}

	r := strings.NewReader(template)
	client := &http.Client{}
	req, argoErr := http.NewRequest(method, url, r)

	res, argoErr := client.Do(req)
	if argoErr != nil {
		logger.Info("err in run workflow")
		return errs.NewUnexpectedError("Unexpected from cluster")
	}

	defer res.Body.Close()

	body, bodyerr := ioutil.ReadAll(res.Body)

	if bodyerr != nil {
		logger.Info("err in run workflow")
		return errs.NewUnexpectedError("Unexpected from cluster")
	}
	logger.Info("ass")
	fmt.Println(string(body))
	if res.StatusCode == 409 {
		return errs.NewUnexpectedError("Workflow name has already triggered ")
	} else {
		status := "Running"
		lastExecutedDate := time.Now().Format(dbTSLayout)
		triggeredBy := userId
		err := s.repo.UpdateWorkflowRun(id, status, lastExecutedDate, triggeredBy)
		if err != nil {
			logger.Info("err in run workflow")
			return errs.NewUnexpectedError("Unexpected from UpdateWorkflowRun")
		}
	}

	return nil

}

func (s DefaultWorkflowService) RetryRunWorkflow(name string) *errs.AppError {
	url := "https://" + os.Getenv("ARGO_SERVER_ENDPOINT") + ":2746/api/v1/workflows/argo/" + name + "/retry"
	method := "PUT"

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		log.Info(err)
		return errs.NewUnexpectedError("Unexpected from cluster")
	}
	res, err := client.Do(req)
	if err != nil {
		log.Info(err)
		return errs.NewUnexpectedError("Unexpected from cluster")
	}
	defer res.Body.Close()

	_, Readerr := ioutil.ReadAll(res.Body)
	if Readerr != nil {
		log.Info(err)
		return errs.NewUnexpectedError("Unexpected from cluster")
	}

	return nil
}

func (s DefaultWorkflowService) AddWorkflow(req dto.AddWorkflowRequest) (*dto.AddWorkflowResponse, *errs.AppError) {

	config := req.Config

	c := domain.Workflow{

		Name:           req.Name,
		Project_Id:     req.Project_id,
		Created_By:     req.Created_By,
		Config:         config,
		CreatedDate:    time.Now().Format(dbTSLayout),
		WorkflowStatus: "Unexecuted",
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
