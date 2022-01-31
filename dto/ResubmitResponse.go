package dto

import "time"

type ReSubmitResponse struct {
	Metadata struct {
		Name              string    `json:"name"`
		GenerateName      string    `json:"generateName"`
		Namespace         string    `json:"namespace"`
		UID               string    `json:"uid"`
		ResourceVersion   string    `json:"resourceVersion"`
		Generation        int       `json:"generation"`
		CreationTimestamp time.Time `json:"creationTimestamp"`
		Labels            struct {
			WorkflowsArgoprojIoResubmittedFromWorkflow string `json:"workflows.argoproj.io/resubmitted-from-workflow"`
		} `json:"labels"`
		Annotations struct {
			WorkflowsArgoprojIoPodNameFormat string `json:"workflows.argoproj.io/pod-name-format"`
		} `json:"annotations"`
		ManagedFields []struct {
			Manager    string    `json:"manager"`
			Operation  string    `json:"operation"`
			APIVersion string    `json:"apiVersion"`
			Time       time.Time `json:"time"`
			FieldsType string    `json:"fieldsType"`
			FieldsV1   struct {
				FMetadata struct {
					FAnnotations struct {
						NAMING_FAILED struct {
						} `json:"."`
						FWorkflowsArgoprojIoPodNameFormat struct {
						} `json:"f:workflows.argoproj.io/pod-name-format"`
					} `json:"f:annotations"`
					FGenerateName struct {
					} `json:"f:generateName"`
					FLabels struct {
						NAMING_FAILED struct {
						} `json:"."`
						FWorkflowsArgoprojIoResubmittedFromWorkflow struct {
						} `json:"f:workflows.argoproj.io/resubmitted-from-workflow"`
					} `json:"f:labels"`
				} `json:"f:metadata"`
				FSpec struct {
				} `json:"f:spec"`
				FStatus struct {
				} `json:"f:status"`
			} `json:"fieldsV1"`
		} `json:"managedFields"`
	} `json:"metadata"`
	Spec struct {
		Templates []struct {
			Name   string `json:"name"`
			Inputs struct {
			} `json:"inputs"`
			Outputs struct {
			} `json:"outputs"`
			Metadata struct {
			} `json:"metadata"`
			Dag struct {
				Tasks []struct {
					Name      string `json:"name"`
					Template  string `json:"template"`
					Arguments struct {
						Parameters []struct {
							Name  string `json:"name"`
							Value string `json:"value"`
						} `json:"parameters"`
					} `json:"arguments"`
				} `json:"tasks"`
			} `json:"dag,omitempty"`
			Script struct {
				Name      string   `json:"name"`
				Image     string   `json:"image"`
				Command   []string `json:"command"`
				Resources struct {
				} `json:"resources"`
				Source string `json:"source"`
			} `json:"script,omitempty"`
		} `json:"templates"`
		Entrypoint string `json:"entrypoint"`
		Arguments  struct {
			Parameters []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"parameters"`
		} `json:"arguments"`
	} `json:"spec"`
	Status struct {
		StartedAt  interface{} `json:"startedAt"`
		FinishedAt interface{} `json:"finishedAt"`
	} `json:"status"`
}
