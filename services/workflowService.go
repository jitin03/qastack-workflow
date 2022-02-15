package services

import (
	"crypto/tls"
	"encoding/json"
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
	UpdateWorkflowConfig(request dto.UpdateWorkflowRequest, id string) *errs.AppError
	AllWorkflows(projectKey string, pageId int) ([]dto.AllWorkflowResponse, *errs.AppError)
	RunWorkflow(string, userId string) *errs.AppError
	RetryRunWorkflow(string, userId string, workflowId string) *errs.AppError
	ReSubmitRunWorkflow(name string, userId string) (*dto.ReSubmitRunWorkflowResponse, *errs.AppError)
	UpdateWorkflowStatus(request dto.UpdateWorkflowStatus) *errs.AppError
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

func (s DefaultWorkflowService) RetryRunWorkflow(name string, userId string, workflowId string) *errs.AppError {

	deleteUrl := "https://" + os.Getenv("ARGO_SERVER_ENDPOINT") + ":2746/api/v1/workflows/argo/" + name
	log.Info(deleteUrl)
	method := "DELETE"
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := &http.Client{}
	req, err := http.NewRequest(method, deleteUrl, nil)

	if err != nil {
		fmt.Println(err)
		return errs.NewUnexpectedError("Unexpected from cluster")
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return errs.NewUnexpectedError("Unexpected from cluster")
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return errs.NewUnexpectedError("Unexpected from cluster")
	}
	if res.StatusCode != 200 {
		fmt.Println(err)
		return errs.NewUnexpectedError("Unexpected from cluster")
	}

	url := "https://" + os.Getenv("ARGO_SERVER_ENDPOINT") + ":2746/api/v1/workflows/argo"
	postMethod := "POST"
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	template, runNorErr := s.repo.RunWorkflow(workflowId, userId)
	if runNorErr != nil {
		logger.Info("err in run workflow")

		return errs.NewUnexpectedError("Unexpected from cluster")
	}

	r := strings.NewReader(template)
	client = &http.Client{}
	req, argoErr := http.NewRequest(postMethod, url, r)

	res, argoErr = client.Do(req)
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
		err := s.repo.UpdateWorkflowRun(workflowId, status, lastExecutedDate, triggeredBy)
		if err != nil {
			logger.Info("err in run workflow")
			return errs.NewUnexpectedError("Unexpected from UpdateWorkflowRun")
		}
	}
	// url := "https://" + os.Getenv("ARGO_SERVER_ENDPOINT") + ":2746/api/v1/workflows/argo/" + name + "/retry"
	// method := "PUT"

	// http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// client := &http.Client{}
	// req, err := http.NewRequest(method, url, nil)

	// if err != nil {
	// 	log.Info(err)
	// 	return errs.NewUnexpectedError("Unexpected from cluster")
	// }
	// res, err := client.Do(req)
	// if err != nil {
	// 	log.Info(err)
	// 	return errs.NewUnexpectedError("Unexpected from cluster")
	// }
	// defer res.Body.Close()

	// _, Readerr := ioutil.ReadAll(res.Body)
	// if Readerr != nil {
	// 	log.Info(err)
	// 	return errs.NewUnexpectedError("Unexpected from cluster")
	// }

	return nil
}

func (s DefaultWorkflowService) ReSubmitRunWorkflow(name string, userId string) (*dto.ReSubmitRunWorkflowResponse, *errs.AppError) {

	url := "https://" + os.Getenv("ARGO_SERVER_ENDPOINT") + ":2746/api/v1/workflows/argo/" + name + "/resubmit"
	method := "PUT"

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	res, _ := client.Do(req)

	defer res.Body.Close()

	response, Readerr := ioutil.ReadAll(res.Body)
	if Readerr != nil {
		log.Info(err)
		return nil, errs.NewUnexpectedError("Unexpected from cluster")
	}
	var resubmitWorkflow dto.ReSubmitResponse
	json.Unmarshal([]byte(response), &resubmitWorkflow)
	fmt.Printf("New WorkflowName: %s", resubmitWorkflow.Metadata.Name)
	newWorkflowname := resubmitWorkflow.Metadata.Name
	status := "Resubmitted"
	lastExecutedDate := time.Now().Format(dbTSLayout)
	triggeredBy := userId
	if workflow, err := s.repo.UpdateReSubmitedWorkflowRun(name, newWorkflowname, status, lastExecutedDate, triggeredBy); err != nil {
		logger.Info("err in run workflow")
		return nil, errs.NewUnexpectedError("Unexpected from UpdateReSubmitedWorkflowRun")
	} else {
		logger.Info("err in run workflow")
		logger.Info(workflow.Workflow_Run_Name)
		workflow.Workflow_Run_Name = newWorkflowname
		return workflow.ToReSubmitRunWorkflowResponseDto(), nil
	}

}

func (s DefaultWorkflowService) UpdateWorkflowStatus(req dto.UpdateWorkflowStatus) *errs.AppError {

	w := domain.WorkflowRuns{
		WorkflowId:       req.WorkflowId,
		WorkflowName:     req.WorkflowName,
		LastExecutedDate: time.Now().Format(dbTSLayout),
		Status:           req.Status,
		UserId:           req.UserId,
	}

	if err := s.repo.UpdateWorkflowStatus(w); err != nil {
		logger.Info("err in run workflow")
		return errs.NewUnexpectedError("Unexpected from UpdateWorkflowStatus")
	}

	return nil
}
func (s DefaultWorkflowService) AddWorkflow(req dto.AddWorkflowRequest) (*dto.AddWorkflowResponse, *errs.AppError) {

	config := req.Config

	c := domain.Workflow{

		Name:              req.Name,
		Project_Id:        req.Project_id,
		Created_By:        req.Created_By,
		Updated_By:        req.Created_By,
		Config:            config,
		CreatedDate:       time.Now().Format(dbTSLayout),
		UpdatedDate:       time.Now().Format(dbTSLayout),
		LastExecutedDate:  "-",
		WorkflowStatus:    "Build Now",
		Workflow_Run_Name: req.Name,
	}

	if newComponent, err := s.repo.AddWorkflow(c); err != nil {
		return nil, err
	} else {
		return newComponent.ToAddWorkflowResponseDto(), nil
	}

}
func (s DefaultWorkflowService) UpdateWorkflowConfig(req dto.UpdateWorkflowRequest, id string) *errs.AppError {
	config := req.Config

	c := domain.Workflow{

		Name:              req.Name,
		Project_Id:        req.Project_id,
		Updated_By:        req.Updated_By,
		Config:            config,
		UpdatedDate:       time.Now().Format(dbTSLayout),
		WorkflowStatus:    "Build Again",
		Workflow_Run_Name: req.Name,
	}

	if err := s.repo.UpdateWorkflowConfig(c, id); err != nil {
		return nil
	}
	return nil
}

func NewWorkflowService(repository domain.WorkflowRepository) DefaultWorkflowService {
	return DefaultWorkflowService{repository}
}
