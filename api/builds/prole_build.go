package builds

type ProleBuild struct {
	Guid string `json:"guid"`

	LogsURL string `json:"logs_url"`

	Image  string `json:"image"`
	Script string `json:"script"`

	Callback string `json:"callback"`

	Source ProleBuildSource `json:"source"`

	Parameters       map[string]string `json:"parameters"`
	SecureParameters map[string]string `json:"secure_parameters"`

	Status string `json:"status"`
}

type ProleBuildSource struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
	Ref  string `json:"ref"`
	Path string `json:"path"`
}
