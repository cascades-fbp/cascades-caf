package caf

// Describes a context property
type Property struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"n"`
	Timestamp   int                    `json:"t"`
	Value       float64                `json:"v"`
	BoolValue   bool                   `json:"bv"`
	StringValue string                 `json:"sv"`
	JsonValue   map[string]interface{} `"json:"json"`
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
	Name     string `json:"name"`
	Type     string `json:"type"`
	Template string `json:"template"`
}
