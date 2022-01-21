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

func (u WorkflowHandler) SubscribeToEvent(w http.ResponseWriter, r *http.Request) {
	workflowName := r.URL.Query().Get("id")
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
	eventURI := "https://" + os.Getenv("ARGO_SERVER_ENDPOINT") + ":2746/api/v1/workflow-events/argo?listOptions.fieldSelector=metadata.name=" + workflowName
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
			fmt.Fprintf(w, "data:%s \n", data.Data)
			flusher.Flush()
		}
	}
}
