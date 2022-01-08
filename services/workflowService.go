package services

import (
	"qastack-workflows/domain"
	"qastack-workflows/dto"
	"qastack-workflows/errs"
)

type WorkflowServices interface {
	AddWorkflow(request dto.AddWorkflowRequest) (*dto.AddWorkflowResponse, *errs.AppError)
	AllWorkflows(projectKey string, pageId int) ([]dto.AllWorkflowResponse, *errs.AppError)
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

func (s DefaultWorkflowService) AddWorkflow(req dto.AddWorkflowRequest) (*dto.AddWorkflowResponse, *errs.AppError) {

	c := domain.Workflow{

		Name:       req.Name,
		Project_Id: req.Project_id,
		Created_By: req.Created_By,
		Config: []struct {
			Name       string `db:"name"`
			Repository string `db:"repository"`
			Branch     string `db:"branch"`
			Git_Token  string `db:"token"`
		}(req.Config),
	}

	if newComponent, err := s.repo.AddWorkflow(c); err != nil {
		return nil, err
	} else {
		return newComponent.ToAddWorkflowResponseDto(), nil
	}
	return nil, nil
}

func NewWorkflowService(repository domain.WorkflowRepository) DefaultWorkflowService {
	return DefaultWorkflowService{repository}
}
