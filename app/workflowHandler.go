package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"qastack-workflows/dto"
	"qastack-workflows/services"
	"strconv"
)

type WorkflowHandler struct {
	service services.WorkflowServices
}

func (u WorkflowHandler) AddWorkflow(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
	var request dto.AddWorkflowRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		WriteResponse(w, http.StatusBadRequest, err.Error())
	} else {

		userId, appError := u.service.AddWorkflow(request)
		if appError != nil {
			WriteResponse(w, appError.Code, appError.AsMessage())
		} else {
			WriteResponse(w, http.StatusCreated, userId)
		}
	}
}

func (u WorkflowHandler) AllWorkflows(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	projectKey := r.URL.Query().Get("projectKey")

	pageId, _ := strconv.Atoi(page)
	// projectKeyId, _ := strconv.Atoi(projectKey)
	components, err := u.service.AllWorkflows(projectKey, pageId)

	if err != nil {
		fmt.Println("Inside error" + err.Message)

		WriteResponse(w, err.Code, err.AsMessage())
	} else {

		WriteResponse(w, http.StatusOK, components)
	}
}

func (u WorkflowHandler) RunWorkflow(w http.ResponseWriter, r *http.Request) {
	workflowId := r.URL.Query().Get("id")

	fmt.Println(workflowId)
	type responseBody struct {
		WorkflowResponse string `json:"workflow_response"`
	}
	_, err := u.service.RunWorkflow(workflowId)
	if err != nil {
		fmt.Println("Inside error" + err.Message)

		WriteResponse(w, err.Code, err.AsMessage())
	} else {

		respondWithJSON(w, 200, responseBody{
			WorkflowResponse: "workflow:" + workflowId + " is triggered!",
		})
	}
}
