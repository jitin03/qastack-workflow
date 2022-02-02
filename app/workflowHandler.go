package app

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"qastack-workflows/dto"
	"qastack-workflows/services"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/r3labs/sse"
	log "github.com/sirupsen/logrus"
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
	workflows, err := u.service.AllWorkflows(projectKey, pageId)

	if err != nil {
		fmt.Println("Inside error" + err.Message)

		WriteResponse(w, err.Code, err.AsMessage())
	} else {

		WriteResponse(w, http.StatusOK, workflows)
	}
}

func (u WorkflowHandler) DeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	// convert the id type from string to int
	id := params["id"]
	type responseBody struct {
		WorkflowResponse string `json:"message"`
	}
	error := u.service.DeleteWorkflow(id)
	if error != nil {
		fmt.Println("Inside error" + error.Message)

		WriteResponse(w, error.Code, error.AsMessage())
	} else {
		respondWithJSON(w, 200, responseBody{
			WorkflowResponse: "workflow:" + id + " is deleted successfully!",
		})
	}

}

func (u WorkflowHandler) GetWorkflowDetail(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	// convert the id type from string to int
	id := params["workflowName"]
	type responseBody struct {
		WorkflowResponse string `json:"message"`
	}
	workflow, error := u.service.GetWorkflowDetail(id)
	if error != nil {
		fmt.Println("Inside error" + error.Message)

		WriteResponse(w, error.Code, error.AsMessage())
	} else {
		WriteResponse(w, http.StatusOK, workflow)
	}
}

func (u WorkflowHandler) RunWorkflow(w http.ResponseWriter, r *http.Request) {
	workflowName := r.URL.Query().Get("id")
	userId := r.URL.Query().Get("userId")

	fmt.Println(workflowName)
	type responseBody struct {
		WorkflowResponse string `json:"workflow_response"`
	}
	err := u.service.RunWorkflow(workflowName, userId)
	if err != nil {
		fmt.Println("Inside error" + err.Message)

		WriteResponse(w, err.Code, err.AsMessage())
	} else {

		respondWithJSON(w, 200, responseBody{
			WorkflowResponse: "workflow:" + workflowName + " is triggered!",
		})
	}
}

func (u WorkflowHandler) RetryRunWorkflow(w http.ResponseWriter, r *http.Request) {
	workflowId := r.URL.Query().Get("workflowName")

	fmt.Println(workflowId)
	type responseBody struct {
		WorkflowResponse string `json:"workflow_response"`
	}
	err := u.service.RetryRunWorkflow(workflowId)
	if err != nil {
		fmt.Println("Inside error" + err.Message)

		WriteResponse(w, err.Code, err.AsMessage())
	} else {

		respondWithJSON(w, 200, responseBody{
			WorkflowResponse: "workflow:" + workflowId + " is triggered!",
		})
	}
}

func (u WorkflowHandler) ReSubmitRunWorkflow(w http.ResponseWriter, r *http.Request) {
	workflowId := r.URL.Query().Get("workflowName")
	userId := r.URL.Query().Get("userId")

	fmt.Println(workflowId)
	type responseBody struct {
		WorkflowResponse string `json:"workflow_response"`
	}
	res, err := u.service.ReSubmitRunWorkflow(workflowId, userId)
	if err != nil {
		fmt.Println("Inside error" + err.Message)

		WriteResponse(w, err.Code, err.AsMessage())
	} else {

		WriteResponse(w, http.StatusOK, res)
	}
}

func (u WorkflowHandler) UpdateWorkflowStatus(w http.ResponseWriter, r *http.Request) {

	var request dto.UpdateWorkflowStatus
	err := json.NewDecoder(r.Body).Decode(&request)
	type responseBody struct {
		WorkflowResponse string `json:"workflow_response"`
	}

	if err != nil {
		WriteResponse(w, http.StatusBadRequest, err.Error())
	} else {

		err := u.service.UpdateWorkflowStatus(request)
		if err != nil {
			WriteResponse(w, err.Code, err.AsMessage())
		} else {
			respondWithJSON(w, 200, responseBody{
				WorkflowResponse: "workflow is updated ",
			})
		}
	}

}

// https://a973a7c68601640278113fe98be8a89d-49052598.us-east-1.elb.amazonaws.com:2746/api/v1/workflows/argo/jjjj-9549b/log?logOptions.container=main&grep=&logOptions.follow=true&podName=jjjj-9549
// https://a973a7c68601640278113fe98be8a89d-49052598.us-east-1.elb.amazonaws.com:2746/api/v1/workflows/argo/jjjj-9549b/log?logOptions.container=main&grep=&logOptions.follow=true
func (u WorkflowHandler) WorkflowLogs(w http.ResponseWriter, r *http.Request) {
	workflowName := r.URL.Query().Get("workflowName")
	fmt.Println("hello Event")

	events := make(chan *sse.Event)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// instantiate the channel
	// messageChan = make(chan string)
	// close the channel after exit the function
	defer func() {
		close(events)
		events = nil
		log.Printf("client connection is closed")
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Connection doesnot support streaming", http.StatusBadRequest)
		return
	}

	d := make(chan interface{})
	defer close(d)
	defer fmt.Println("Closing channel.")
	eventURI := "https://" + os.Getenv("ARGO_SERVER_ENDPOINT") + ":2746/api/v1/workflows/argo/" + workflowName + "/log?logOptions.container=main&grep=&logOptions.follow=true"
	log.Info(eventURI)
	client := sse.NewClient(eventURI)
	client.Connection.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client.SubscribeChan("logs", events)

	// client.Subscribe("message", func(msg *sse.Event) {
	// 	// Got some data!
	// 	// fmt.Println(msg.Data)
	// 	argoMessage := string(msg.Data)
	// 	log.Info(argoMessage)
	// 	fmt.Printf("var1 = %T\n", msg.Data)
	// 	fmt.Fprintf(w, "data: %c \n", msg.Data)

	// })

	// client.SubscribeRaw(func(msg *sse.Event) {
	// 	// Got some data!
	// 	// fmt.Println(string(msg.Data))
	// 	// fmt.Printf("data: %v \n\n", string(msg.Data))
	// })
	for {
		select {
		case <-d:
			close(events)
			return
		case data := <-events:

			// fmt.Printf("data: %v ", *data)
			// fmt.Fprintf(w, "data: %v \n\n", data)
			fmt.Fprintf(w, "data:%s \n\n", data.Data)
			flusher.Flush()
		}
	}

}

func (u WorkflowHandler) SubscribeToEvent(w http.ResponseWriter, r *http.Request) {
	workflowName := r.URL.Query().Get("workflowName")
	fmt.Println("hello Event")

	events := make(chan *sse.Event)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// instantiate the channel
	// messageChan = make(chan string)
	// close the channel after exit the function
	defer func() {
		close(events)
		events = nil
		log.Printf("client connection is closed")
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Connection doesnot support streaming", http.StatusBadRequest)
		return
	}

	d := make(chan interface{})
	defer close(d)
	defer fmt.Println("Closing channel.")
	eventURI := "https://" + os.Getenv("ARGO_SERVER_ENDPOINT") + ":2746/api/v1/workflow-events/argo?listOptions.fieldSelector=metadata.namespace=argo,metadata.name=" + workflowName
	log.Info(eventURI)
	client := sse.NewClient(eventURI)
	client.Connection.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client.SubscribeChan("message", events)

	// client.Subscribe("message", func(msg *sse.Event) {
	// 	// Got some data!
	// 	// fmt.Println(msg.Data)
	// 	argoMessage := string(msg.Data)
	// 	log.Info(argoMessage)
	// 	fmt.Printf("var1 = %T\n", msg.Data)
	// 	fmt.Fprintf(w, "data: %c \n", msg.Data)

	// })

	// client.SubscribeRaw(func(msg *sse.Event) {
	// 	// Got some data!
	// 	// fmt.Println(string(msg.Data))
	// 	// fmt.Printf("data: %v \n\n", string(msg.Data))
	// })
	for {
		select {
		case <-d:
			close(events)
			return
		case data := <-events:

			// fmt.Printf("data: %v ", *data)
			// fmt.Fprintf(w, "data: %v \n\n", data)
			fmt.Fprintf(w, "data:%s \n\n", data.Data)
			flusher.Flush()
		}
	}
}
