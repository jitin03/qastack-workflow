package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"qastack-workflows/errs"
	logger "qastack-workflows/loggers"

	"github.com/jmoiron/sqlx"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type WorkflowRepositoryDb struct {
	client *sqlx.DB
}

func (w WorkflowRepositoryDb) UpdateWorkflowConfig(workflow Workflow, id string) *errs.AppError {

	log.Info("WorkflowStatus", workflow.WorkflowStatus)
	updateWorkflowSql := "UPDATE workflows SET workflowname = $1 ,workflow_run_name = $2 ,project_id=$3, updated_by=$4,config=$5,workflow_status=$6,updated_date=$7 WHERE id = $8"
	res, err := w.client.Exec(updateWorkflowSql, workflow.Name, workflow.Workflow_Run_Name, workflow.Project_Id, workflow.Updated_By, workflow.Config, workflow.WorkflowStatus, workflow.UpdatedDate, id)
	if err != nil {
		return errs.NewUnexpectedError("Unexpected error from database")
	}
	count, err := res.RowsAffected()
	if err != nil {
		return errs.NewUnexpectedError("Unexpected error from database")
	}
	fmt.Println(count)
	return nil
}

func (w WorkflowRepositoryDb) AddWorkflow(workflow Workflow) (*Workflow, *errs.AppError) {

	log.Info(workflow.Config)

	var id string
	sqlInsert := "INSERT INTO public.workflows (workflowname,workflow_run_name ,project_id,created_by,config,created_date,workflow_status,updated_by,updated_date,last_execution_date) values ($1, $2, $3,$4,$5,$6,$7,$8,$9,$10) RETURNING id"

	err := w.client.QueryRow(sqlInsert, workflow.Name, workflow.Workflow_Run_Name, workflow.Project_Id, workflow.Created_By, workflow.Config, workflow.CreatedDate, workflow.WorkflowStatus, workflow.Updated_By, workflow.UpdatedDate, workflow.LastExecutedDate).Scan(&id)

	// in case of error Rollback, and changes from both the tables will be reverted
	if err != nil {
		logger.Error("Error while creating new component: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected error from database")
	}

	workflow.Workflow_Id = id
	logrus.Info(id)

	return &workflow, nil
}

func (w WorkflowRepositoryDb) DeleteWorkflow(id string) *errs.AppError {
	log.Info(id)
	deleteSql := "DELETE FROM workflows WHERE id = $1"
	res, err := w.client.Exec(deleteSql, id)
	if err != nil {
		panic(err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}
	fmt.Println(count)

	return nil
}

func (d WorkflowRepositoryDb) GetWorkflowDetail(workflowName string) (*Workflow, *errs.AppError) {
	var err error
	var workflow Workflow
	logrus.Info(workflowName)
	findAllSql := "select id,workflowname,config from public.workflows where workflowname=$1"
	err = d.client.Get(&workflow, findAllSql, workflowName)

	if err != nil {
		fmt.Println("Error while querying workflow table table " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	return &workflow, nil

}

func (d WorkflowRepositoryDb) AllWorkflows(projectKey string, pageId int) ([]Workflow, *errs.AppError) {
	var err error
	workflows := make([]Workflow, 0)
	logrus.Info(projectKey)
	findAllSql := "select workflow_status,id,workflowname, project_id,u.username,workflow_run_name,created_by,updated_by,updated_date,created_date,last_execution_date from public.workflows w join public.users u on u.users_id= w.created_by where project_id=$1 LIMIT $2"
	err = d.client.Select(&workflows, findAllSql, projectKey, pageId)

	if err != nil {
		fmt.Println("Error while querying workflow table " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	return workflows, nil

}

func (w WorkflowRepositoryDb) UpdateReSubmitedWorkflowRun(oldWorkflowName string, newWorkflowName string, status string, lastExecutedDate string, triggeredBy string) (*Workflow, *errs.AppError) {
	var err error
	var workflow Workflow
	logrus.Info(oldWorkflowName)
	// findAllSql := "select id,workflowname,config from public.workflows where workflow_run_name=$1"
	// err = w.client.Get(&workflow, findAllSql, oldWorkflowName)

	// if err != nil {
	// 	fmt.Println("Error while querying workflow table table " + err.Error())
	// 	return nil, errs.NewUnexpectedError("Unexpected database error")
	// }

	tx, err := w.client.Begin()
	if err != nil {
		logger.Error("Error while starting a new transaction for test status transaction: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	sqlInsert := "insert into workflow_runs (workflow_id,name,status,last_executed_date,executed_by) values ((select id from workflows where workflow_run_name=$1),$2,$3,$4,$5) RETURNING id"

	_, err = tx.Exec(sqlInsert, oldWorkflowName, newWorkflowName, status, lastExecutedDate, triggeredBy)

	// in case of error Rollback, and changes from both the tables will be reverted
	if err != nil {
		tx.Rollback()
		logger.Error("Error while saving transaction into test_status_records: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}
	update_workflow_status := "UPDATE workflows SET workflow_status = (select status from public.workflow_runs wr where name =$1 order by last_executed_date desc limit 1  ), workflow_run_name =(select name from public.workflow_runs wr where name =$2 order by last_executed_date desc limit 1  ) WHERE id=$3"
	_, err = tx.Exec(update_workflow_status, newWorkflowName, newWorkflowName, workflow.Workflow_Id)

	// in case of error Rollback, and changes from both the tables will be reverted
	if err != nil {
		tx.Rollback()
		logger.Error("Error while saving transaction into workflows table: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	// commit the transaction when all is good
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		logger.Error("Error while commiting transaction for workflows: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	return &workflow, nil
}

func (w WorkflowRepositoryDb) GetWorkflowNamefromWorkflowRuns(workflowId string) (*WorkflowRuns, *errs.AppError) {
	var err error
	var workflowRuns WorkflowRuns

	findAllSql := "select id,name,status from public.workflow_runs where workflow_id=$1"
	err = w.client.Get(&workflowRuns, findAllSql, workflowId)

	if err != nil {
		fmt.Println("Error while querying workflow table table " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	return &workflowRuns, nil
}

func (w WorkflowRepositoryDb) UpdateWorkflowRun(workflowId string, status string, lastExecutedDate string, triggeredBy string) *errs.AppError {

	tx, err := w.client.Begin()
	if err != nil {
		logger.Error("Error while starting a new transaction for test status transaction: " + err.Error())
		return errs.NewUnexpectedError("Unexpected database error")
	}

	sqlInsert := "insert into workflow_runs (workflow_id,name,status,last_executed_date,executed_by) values ($1,(select workflowname from workflows where id=$2),$3,$4,$5) RETURNING id"

	_, err = tx.Exec(sqlInsert, workflowId, workflowId, status, lastExecutedDate, triggeredBy)

	// in case of error Rollback, and changes from both the tables will be reverted
	if err != nil {
		tx.Rollback()
		logger.Error("Error while saving transaction into test_status_records: " + err.Error())
		return errs.NewUnexpectedError("Unexpected database error")
	}
	update_workflow_status := "UPDATE workflows SET workflow_status = (select status from public.workflow_runs wr where workflow_id =$1 order by last_executed_date desc limit 1  ) WHERE id=$2"
	_, err = tx.Exec(update_workflow_status, workflowId, workflowId)

	// in case of error Rollback, and changes from both the tables will be reverted
	if err != nil {
		tx.Rollback()
		logger.Error("Error while saving transaction into workflows table: " + err.Error())
		return errs.NewUnexpectedError("Unexpected database error")
	}

	// commit the transaction when all is good
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		logger.Error("Error while commiting transaction for workflows: " + err.Error())
		return errs.NewUnexpectedError("Unexpected database error")
	}

	return nil
}

func (w WorkflowRepositoryDb) UpdateWorkflowStatus(workflowRuns WorkflowRuns) *errs.AppError {
	tx, err := w.client.Begin()
	if err != nil {
		logger.Error("Error while starting a new transaction for test status transaction: " + err.Error())
		return errs.NewUnexpectedError("Unexpected database error")
	}

	sqlInsert := "insert into workflow_runs (workflow_id,name,status,last_executed_date,executed_by) values ($1,$2,$3,$4,$5) RETURNING id"

	_, err = tx.Exec(sqlInsert, workflowRuns.WorkflowId, workflowRuns.WorkflowName, workflowRuns.Status, workflowRuns.LastExecutedDate, workflowRuns.UserId)

	// in case of error Rollback, and changes from both the tables will be reverted
	if err != nil {
		tx.Rollback()
		logger.Error("Error while saving transaction into test_status_records: " + err.Error())
		return errs.NewUnexpectedError("Unexpected database error")
	}
	update_workflow_status := "UPDATE workflows SET workflow_status = (select status from public.workflow_runs wr where workflow_id =$1 order by last_executed_date desc limit 1  ) , last_execution_date = (select last_executed_date from public.workflow_runs wr where workflow_id =$2 order by last_executed_date desc limit 1) WHERE id=$3"
	_, err = tx.Exec(update_workflow_status, workflowRuns.WorkflowId, workflowRuns.WorkflowId, workflowRuns.WorkflowId)

	// in case of error Rollback, and changes from both the tables will be reverted
	if err != nil {
		tx.Rollback()
		logger.Error("Error while saving transaction into workflows table: " + err.Error())
		return errs.NewUnexpectedError("Unexpected database error")
	}

	// commit the transaction when all is good
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		logger.Error("Error while commiting transaction for workflows: " + err.Error())
		return errs.NewUnexpectedError("Unexpected database error")
	}

	return nil
}
func (d WorkflowRepositoryDb) RunWorkflow(workflowId string, userId string) (string, *errs.AppError) {
	var err error
	var testWorkflow string
	log.Info("Workflow Run for " + workflowId)
	workflow := make([]Workflow, 0)
	parameters := []Parameters{}
	templates := []Templates{}

	//"select id,title,description,type,priority from testcase where component_id=$1 LIMIT $2"
	findAllSql := "select id,workflowname,project_id,created_by,config from public.workflows w where id=$1"
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

		gitRepo, err := url.Parse(c.Repository)
		if err != nil {
			return "Error", errs.NewUnexpectedError("Unexpected repository parsing error")
		}

		// dnsPing := fmt.Sprint("set -e;\n sh -c 'echo "%s+" | tee /etc/resolv.conf > /dev/null';","nameserver 8.8.8.8\nnameserver 8.8.4.4\nnameserver 1.1.1.1\noptions attempts:5 timeout:2 rotate")

		res := fmt.Sprintf("https://%s:x-oauth-basic@github.com%s", "token", gitRepo.Path)
		// dns_ping := "set -e;\n sh -c 'echo 'nameserver 8.8.8.8\nnameserver 8.8.4.4\nnameserver 1.1.1.1\noptions attempts:5 timeout:2 rotate' | tee /etc/resolv.conf > /dev/null'"
		source := "\ngit clone -b " + c.Branch + " " + res + ";\n" + c.Source

		log.Info("c.DockerImage", c.DockerImage)
		script = Script{
			Image:   c.DockerImage,
			Command: commands,
			Source:  source,
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
		GenerateName: workflow[0].Name,
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
