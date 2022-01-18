package dto

type AddWorkflowRequest struct {
	Project_id string `json:"project_Id"`
	Name       string `json:"name"`
	Config     []struct {
		Name         string   `json:"name"`
		Repository   string   `json:"repository"`
		Branch       string   `json:"branch"`
		Git_Token    string   `json:"token"`
		DockerImage  string   `json:"docker_image"`
		EntryPath    []string `json:"entrypath"`
		Dependencies []string `json:"dependencies"`
		Parameters   []struct {
			Name  string `json:"name",omitempty"`
			Value string `json:"value",omitempty"`
		}
		// Parameters []Parameter `json:""`
		Source string `json:"input_command"`
	} `json:"config"`
	Created_By int `json:"user_Id"`
}

type Parameter struct {
	Name  string `json:"name",omitempty"`
	Value string `json:"value",omitempty"`
}
