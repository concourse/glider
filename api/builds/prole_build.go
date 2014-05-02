package builds

type ProleBuild struct {
	Guid string `json:"guid"`

	Image  string      `json:"image"`
	Env    [][2]string `json:"env"`
	Script string      `json:"script"`

	LogsURL  string `json:"logs_url"`
	Callback string `json:"callback"`

	Source ProleBuildSource `json:"source"`

	Status string `json:"status"`
}

type ProleBuildSource struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
	Ref  string `json:"ref"`
	Path string `json:"path"`
}
