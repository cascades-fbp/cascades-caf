package caf

// Constants
const (
	PropTypeFloat  = "float"
	PropTypeBool   = "bool"
	PropTypeString = "string"
)

// Property describes a context property
type Property struct {
	ID          string   `json:"id"`
	Name        string   `json:"n"`
	Group       string   `json:"g"`
	Timestamp   *int64   `json:"t,omitempty"`
	Value       *float64 `json:"v,omitempty"`
	BoolValue   *bool    `json:"bv,omitempty"`
	StringValue *string  `json:"sv,omitempty"`
}

// Context describes a context (primary or secondary)
type Context struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Group       string        `json:"group"`
	Description string        `json:"description"`
	Rule        string        `json:"rule"`
	Matching    bool          `json:"matching"`
	Timestamp   int64         `json:"timestamp"`
	Entries     []interface{} `json:"e"`
}

// PrimaryContext describes a primary (consisting of properties) context
type PrimaryContext struct {
	*Context
	Entries []Property `json:"e"`
}

// SecondaryContext describes a primary (consisting of properties) context
type SecondaryContext struct {
	*Context
	Entries []Context `json:"e"`
}

// PropertyTemplate is used to configure *-property components
type PropertyTemplate struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Group    string `json:"group"`
	Template string `json:"template"`
}
