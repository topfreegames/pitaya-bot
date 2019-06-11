package models

// Spec defines the bots' spec
type Spec struct {
	Name                 string              `json:"name"`
	NumberOfInstances    int                 `json:"numberOfInstances"`
	PreRun               *InitialDefinitions `json:"preRun,omitempty"`
	SequentialOperations []*Operation        `json:"sequentialOperations,omitempty"`
	PostRun              *FinalDefinitions   `json:"postRun,omitempty"`
}

// NewSpec returns a new spec
func NewSpec(name string) *Spec {
	return &Spec{
		Name:                 name,
		NumberOfInstances:    1,
		SequentialOperations: []*Operation{},
	}
}

// InitialDefinitions are set before running each bot
type InitialDefinitions struct {
	Function string `json:"function,omitempty"`
}

// FinalDefinitions are run after finishing running each bot
type FinalDefinitions struct {
	Function string `json:"function,omitempty"`
}

// StoreSpecEntry ...
type StoreSpecEntry struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// StoreSpec ...
type StoreSpec map[string]StoreSpecEntry

// ExpectSpecEntry ...
type ExpectSpecEntry struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// ExpectSpec  ...
type ExpectSpec map[string]ExpectSpecEntry

// Operation defines an operation the bot may execute
type Operation struct {
	Type    string                 `json:"type"`
	Timeout int                    `json:"timeout,omitempty"`
	Wait    bool                   `json:"wait,omitempty"`
	URI     string                 `json:"uri"`
	Args    map[string]interface{} `json:"args"`
	Expect  ExpectSpec             `json:"expect,omitempty"`
	Store   StoreSpec              `json:"store,omitempty"`
	Change  map[string]interface{} `json:"change,omitempty"`
}
