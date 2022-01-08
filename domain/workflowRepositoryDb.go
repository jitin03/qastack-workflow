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
)

type WorkflowRepositoryDb struct {
	client *sqlx.DB
}

func (w WorkflowRepositoryDb) AddWorkflow(workflow Workflow) (*Workflow, *errs.AppError) {

	// starting the database transaction block
	tx, err := w.client.Begin()

	if err != nil {
		logger.Error("Error while starting a new transaction for bank account transaction: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	sqlInsert := "INSERT INTO public.workflows (name, project_id,created_by) values ($1, $2, $3) RETURNING id"

	_, err = tx.Exec(sqlInsert, workflow.Name, workflow.Project_Id, workflow.Created_By)

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

	if err != nil {
		logger.Error("Error while creating new workflow: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected error from database")
	}
	workflow.Workflow_Id = id

	for index, d := range workflow.Config {
		fmt.Println(index, d)
		x := d.Name
		fmt.Println(x)

		sqlTestStepInsert := "INSERT INTO workflow_steps (workflow_id,name, repository,branch,token) values ($1, $2,$3,$4,$5) RETURNING id"

		_, err := tx.Exec(sqlTestStepInsert, id, d.Name, d.Repository, d.Branch, d.Git_Token)
		if err != nil {
			tx.Rollback()
			logger.Error("Error while saving transaction into workflow: " + err.Error())
			return nil, errs.NewUnexpectedError("Unexpected database error")
		}
	}

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
