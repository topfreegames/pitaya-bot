package models

// Spec defines the bots' spec
type Spec struct {
	NumberOfInstances    int                 `json:"numberOfInstances"`
	Initial              *InitialDefinitions `json:"initial,omitempty"`
	SequentialOperations []*Operation        `json:"sequentialOperations,omitempty"`
	Final                *FinalDefinitions   `json:"final,omitempty"`
}

// InitialDefinitions are set before running each bot
type InitialDefinitions struct {
	Function string `json:"function,omitempty"`
}

// FinalDefinitions are run after finishing running each bot
type FinalDefinitions struct {
	Function string `json:"function,omitempty"`
}

// Operation defines an operation the bot may execute
type Operation struct {
	Request string                 `json:"request"`
	Args    map[string]interface{} `json:"args"`
	Expect  map[string]interface{} `json:"expect"`
	Store   map[string]interface{} `json:"store"`
	Change  map[string]interface{} `json:"change"`
}
