package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"qastack-workflows/domain"
	logger "qastack-workflows/loggers"
	"qastack-workflows/services"
	"time"

	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/rs/cors"
)

func getDbClient() *sqlx.DB {

	dbUser := os.Getenv("DB_USER")
	dbPasswd := os.Getenv("DB_PASSWD")
	dbAddr := os.Getenv("DB_ADDR")
	//dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbAddr, 5432, dbUser, dbPasswd, dbName)
	logger.Info(psqlInfo)
	client, err := sqlx.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	// See "Important settings" section.
	client.SetConnMaxLifetime(time.Minute * 3)
	client.SetMaxOpenConns(10)
	client.SetMaxIdleConns(10)

	//client, err := sqlx.ConnectContext(context.Background(), "postgres",os.Getenv("DATABASE_URL") )
	//if err != nil {
	//	panic(err)
	//}
	return client
}

func Start() {

	//sanityCheck()

	router := mux.NewRouter()
	dbClient := getDbClient()

	router.Use()
	workflowRepositoryDb := domain.NewWorkflowRepositoryDb(dbClient)
	////wiring
	////u := ComponentHandler{service.NewUserService(userRepositoryDb,domain.GetRolePermissions())}
	//
	w := WorkflowHandler{services.NewWorkflowService(workflowRepositoryDb)}
	//
	// define routes

	router.HandleFunc("/api/workflow/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode("Running...")
	})
	//
	router.
		HandleFunc("/api/workflow/add", w.AddWorkflow).
		Methods(http.MethodPost).Name("AddWorkflow")

	router.
		HandleFunc("/api/workflows", w.AllWorkflows).
		Methods(http.MethodGet).Name("AllWorkflows")

	router.HandleFunc("/api/workflow/run", w.RunWorkflow).Methods(http.MethodPost).Name("RunWorkflow")

	router.
		HandleFunc("/api/workflow/delete/{id}", w.DeleteWorkflow).
		Methods(http.MethodDelete).Name("DeleteWorkflow")

	router.
		HandleFunc("/api/workflow", w.GetWorkflowDetail).
		Methods(http.MethodGet).Name("GetWorkflowDetail")

	router.HandleFunc("/api/workflow/retry", w.RetryRunWorkflow).Methods(http.MethodPut).Name("RetryRunWorkflow")
	router.HandleFunc("/api/workflow/resubmit", w.ReSubmitRunWorkflow).Methods(http.MethodPut).Name("ReSubmitRunWorkflow")

	router.HandleFunc("/api/workflow/status", w.UpdateWorkflowStatus).Methods(http.MethodPost).Name("UpdateWorkflowStatus")
	router.
		HandleFunc("/api/workflow/update/{id}", w.UpdateWorkflowConfig).
		Methods(http.MethodPut).Name("UpdateWorkflowConfig")
	router.HandleFunc("/api/event/workflow", w.SubscribeToEvent).Methods(http.MethodGet).Name("SubscribeToEvent")
	router.HandleFunc("/api/event/logs", w.WorkflowLogs).Methods(http.MethodGet).Name("WorkflowLogs")

	cor := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Referer"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "PUT", "DELETE", "POST"},
	})

	handler := cor.Handler(router)
	am := AuthMiddleware{domain.NewAuthRepository()}
	router.Use(am.authorizationHandler())

	//logger.Info(fmt.Sprintf("Starting server on %s:%s ...", address, port))
	if err := http.ListenAndServe(":8094", handler); err != nil {
		fmt.Println("Failed to set up server")

	}

}
