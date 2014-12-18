package caf

const (
	PropTypeFloat  = "float"
	PropTypeBool   = "bool"
	PropTypeString = "string"
	PropTypeJSON   = "json"
)

// Describes a context property
type Property struct {
	ID          string                  `json:"id"`
	Group       string                  `json:"group"`
	Name        string                  `json:"n"`
	Timestamp   *int64                  `json:"t,omitempty"`
	Value       *float64                `json:"v,omitempty"`
	BoolValue   *bool                   `json:"bv,omitempty"`
	StringValue *string                 `json:"sv,omitempty"`
	JSONValue   *map[string]interface{} `json:"jv,omitempty"`
}

// Describes a context (primary or secondary)
type Context struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Rule        string        `json:"rule"`
	Matching    bool          `json:"matching"`
	Timestamp   int64         `json:"timestamp"`
	Entries     []interface{} `json:"e"`
}

// Describes a primary (consisting of properties) context
type PrimaryContext struct {
	*Context
	Entries []Property `json:"e"`
}

// Describes a primary (consisting of properties) context
type SecondaryContext struct {
	*Context
	Entries []Context `json:"e"`
}

type PropertyTemplate struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Group    string `json:"group"`
	Template string `json:"template"`
}
