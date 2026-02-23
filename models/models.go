package models

type OSWindow struct {
	Id   int   `json:"id"`
	Tabs []Tab `json:"tabs"`
}

type Tab struct {
	Id      int      `json:"id"`
	Title   string   `json:"title"`
	Layout  string   `json:"layout"`
	Windows []Window `json:"windows"`
}

type Window struct {
	Id  int    `json:"id"`
	Cwd string `json:"cwd"`
}
