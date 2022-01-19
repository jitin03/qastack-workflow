package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"qastack-workflows/errs"
	logger "qastack-workflows/loggers"

	"github.com/jmoiron/sqlx"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type WorkflowRepositoryDb struct {
	client *sqlx.DB
}

func (w WorkflowRepositoryDb) AddWorkflow(workflow Workflow) (*Workflow, *errs.AppError) {

	log.Info(workflow.Config)
	// starting the database transaction block
	tx, err := w.client.Begin()

	if err != nil {
		logger.Error("Error while starting a new transaction for workflow table transaction: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	sqlInsert := "INSERT INTO public.workflows (name, project_id,created_by,config) values ($1, $2, $3,$4) RETURNING id"

	_, err = tx.Exec(sqlInsert, workflow.Name, workflow.Project_Id, workflow.Created_By, workflow.Config)

	// in case of error Rollback, and changes from both the tables will be reverted
	if err != nil {
		tx.Rollback()
		logger.Error("Error while saving transaction into workflow: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	// Run a query to get new workflow id
	row := tx.QueryRow("SELECT id FROM public.workflows WHERE name=$1", workflow.Name)
	var id string
	// Store the count in the `catCount` variable
	err = row.Scan(&id)

	if err != nil {
		tx.Rollback()
		logger.Error("Error while getting workflow id : " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}
	workflow.Workflow_Id = id
	logrus.Info(id)

	// if err != nil {
	// 	logger.Error("Error while creating new workflow: " + err.Error())
	// 	return nil, errs.NewUnexpectedError("Unexpected error from database")
	// }
	// workflow.Workflow_Id = id

	// // for index, d := range workflow.Config {
	// // 	fmt.Println(index, d)
	// // 	x := d.Name
	// // 	fmt.Println(x)

	// // 	entrypoint := make([]string, 0)
	// // 	dependencies := make([]string, 0)

	// // 	entrypoint = append(entrypoint, d.EntryPath...)

	// // 	dependencies = append(dependencies, d.Dependencies...)

	// // 	dependencies_as_string, _ := json.Marshal(dependencies)
	// // 	entrypoint_as_string, _ := json.Marshal(entrypoint)
	// // 	fmt.Println(string(dependencies_as_string))
	// // 	fmt.Println(string(entrypoint_as_string))

	// // 	sqlWorkflowStepInsert := "INSERT INTO workflow_steps (workflow_id,name, repository,branch,token,docker_image,entrypath,dependencies,input_command) values ($1, $2,$3,$4,$5,$6,$7,$8,$9) RETURNING id"

	// // 	_, err := tx.Exec(sqlWorkflowStepInsert, id, d.Name, d.Repository, d.Branch, d.Git_Token, d.DockerImage, entrypoint_as_string, dependencies_as_string, d.Source)
	// // 	if err != nil {
	// // 		tx.Rollback()
	// // 		logger.Error("Error while saving transaction into workflow step: " + err.Error())
	// // 		return nil, errs.NewUnexpectedError("Unexpected database error")
	// // 	}

	// // 	for _, p := range d.Parameters {

	// // 		var param_id string

	// // 		// Run a query to get new workflow id
	// // 		stepParamRow := tx.QueryRow("SELECT id FROM public.workflow_steps WHERE name=$1 and workflow_id = $2", d.Name, id)
	// // 		err = stepParamRow.Scan(&param_id)

	// // 		if err != nil {
	// // 			tx.Rollback()
	// // 			logger.Error("Error while getting workflow id : " + err.Error())
	// // 			return nil, errs.NewUnexpectedError("Unexpected database error")
	// // 		}

	// // 		sqlWorkflowStepParamInsert := "INSERT INTO workflow_steps_param (workflow_step_id,name, value) values ($1, $2,$3) RETURNING id"

	// // 		_, err := tx.Exec(sqlWorkflowStepParamInsert, param_id, p.Name, p.Value)
	// // 		if err != nil {
	// // 			tx.Rollback()
	// // 			logger.Error("Error while saving transaction into workflow: " + err.Error())
	// // 			return nil, errs.NewUnexpectedError("Unexpected database error")
	// // 		}
	// 	}

	// }

	// commit the transaction when all is good
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		logger.Error("Error while commiting transaction for workflow: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}
	return &workflow, nil
}

func (d WorkflowRepositoryDb) AllWorkflows(projectKey string, pageId int) ([]Workflow, *errs.AppError) {
	var err error
	workflows := make([]Workflow, 0)
	logrus.Info(projectKey)
	findAllSql := "select id,name, project_id,created_by from public.workflows where project_id=$1 LIMIT $2"
	err = d.client.Select(&workflows, findAllSql, projectKey, pageId)

	if err != nil {
		fmt.Println("Error while querying component table " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	return workflows, nil

}

func (d WorkflowRepositoryDb) RunWorkflow(workflowId string) (string, *errs.AppError) {
	var err error
	var testWorkflow string
	log.Info("Workflow Run for " + workflowId)
	workflow := make([]Workflow, 0)
	parameters := []Parameters{}
	templates := []Templates{}

	//"select id,title,description,type,priority from testcase where component_id=$1 LIMIT $2"
	findAllSql := "select id,name,project_id,created_by,config from public.workflows w where id=$1"
	err = d.client.Select(&workflow, findAllSql, workflowId)

	log.Info(" from table", string(workflow[0].Config))

	if err != nil {
		fmt.Println("Error while querying workflow table " + err.Error())
		return "Error", errs.NewUnexpectedError("Unexpected database error")
	}

	config := []Config{}

	json.Unmarshal([]byte(workflow[0].Config), &config)
	task := []Tasks{}
	var argument Arguments
	var script Script
	log.Info(config[0])
	for _, c := range config {

		log.Info(c.Name)
		commands := make([]string, 0)
		dependencies := make([]string, 0)

		// entrypoint = append(entrypoint, d.EntryPath...)

		for _, dependency := range c.Dependencies {
			dependencies = append(dependencies, dependency)

		}

		for _, entrypath := range c.EntryPath {
			commands = append(commands, entrypath)

		}
		// commands := []string{"python"}
		// dependencies := []string{"Task1"}

		log.Info("c.DockerImage", c.DockerImage)
		script = Script{
			Image:   c.DockerImage,
			Command: commands,
			Source:  c.Source,
		}

		parameters := []Parameters{}

		for _, p := range c.Parameters {
			parameter := Parameters{
				Name:  p.Name,
				Value: p.Value,
			}
			parameters = append(parameters, parameter)
		}
		// parameter1 := Parameters{
		// 	Name:  "Param1",
		// 	Value: "ParamValue1",
		// }

		// parameters = append(parameters, parameter1)

		argument = Arguments{
			Parameters: parameters,
		}

		t1 := Tasks{Name: c.Name, Template: "task-template", Arguments: &argument}
		// t2 := Tasks{Name: "Task2", Template: "task-template", Dependencies: dependencies, Arguments: &argument}
		task = append(task, t1)
		// task = append(task, t2)

	}

	tasks := task

	dag := Dag{
		Tasks: tasks,
	}

	inputs := Inputs{
		Parameters: parameters,
	}

	template1 := Templates{
		Name:   "dag-template",
		Dag:    &dag,
		Inputs: &inputs,
	}

	template2 := Templates{
		Name:   "task-template",
		Script: &script,
		Inputs: &inputs,
	}

	templates = append(templates, template1)
	templates = append(templates, template2)

	spec := Spec{
		Entrypoint: "dag-template",
		Templates:  templates,
		Arguments:  &argument,
	}

	metadata := Metadata{
		GenerateName: "wf-dag-template-",
	}

	workflowTemplate := &WorkflowTemplate{
		APIVersion: "argoproj.io/v1alpha1",
		Kind:       "Workflow",
		Metadata:   metadata,
		Spec:       spec,
	}

	generateTemplate := &GenerateTemplate{
		WorkflowTemplate: *workflowTemplate,
	}
	data, _ := json.MarshalIndent(generateTemplate, "", "  ")
	// data is the JSON string represented as bytes
	// the second parameter here is the error, which we
	// are ignoring for now, but which you should ideally handle
	// in production grade code

	// to print the data, we can typecast it to a string
	// fmt.Println(string(data))
	testWorkflow = string(data)
	log.Info(testWorkflow)
	return string(testWorkflow), nil
}

func NewWorkflowRepositoryDb(dbClient *sqlx.DB) WorkflowRepositoryDb {
	return WorkflowRepositoryDb{dbClient}
}

// Make the Steps struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (a Steps) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Make the Steps struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (a *Steps) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}
