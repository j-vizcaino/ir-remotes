package ui

type Remote struct {
	Name    string   `json:"name"`
	Text    string   `json:"text,omitempty"`
	Buttons []Button `json:"buttons"`
}

type Button struct {
	Name  string `json:"name"`
	Class string `json:"class,omitempty"`
	Icon  string `json:"icon,omitempty"`
	Text  string `json:"text"`
}
