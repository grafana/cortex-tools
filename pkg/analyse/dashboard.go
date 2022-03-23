package analyse

type (
	Board struct {
		ID         uint       `json:"id,omitempty"`
		UID        string     `json:"uid,omitempty"`
		Slug       string     `json:"slug"`
		Title      string     `json:"title"`
		Panels     []*Panel   `json:"panels"`
		Rows       []*Row     `json:"rows"`
		Templating Templating `json:"templating"`
	}
	Templating struct {
		List []TemplateVar `json:"list"`
	}
	TemplateVar struct {
		Name  string      `json:"name"`
		Query interface{} `json:"query"`
		Type  string      `json:"type"`
	}
	Panel struct {
		Targets []Target `json:"targets,omitempty"`
		Title   string   `json:"title"`
		Panels  []Panel  `json:"panels"` // row panel type
		Type    string   `json:"type"`
	}
	Target struct {
		RefID      string `json:"refId"`
		Datasource string `json:"datasource,omitempty"`
		// For Prometheus
		Expr string `json:"expr,omitempty"`
	}
	Row struct {
		Panels []Panel `json:"panels"`
	}
)
