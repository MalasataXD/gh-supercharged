package projects

type FieldOption struct {
	ID   string
	Name string
}

type Field struct {
	ID      string
	Name    string
	Type    string
	Options []FieldOption
}

type IDs struct {
	ItemNodeID    string
	ProjectNodeID string
	FieldID       string
	OptionID      string // empty for non-single-select fields
}
