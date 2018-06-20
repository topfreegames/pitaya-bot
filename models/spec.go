package models

// Spec defines the bots' spec
type Spec struct {
	NumberOfInstances    int                 `json:"numberOfInstances"`
	PreRun               *InitialDefinitions `json:"preRun,omitempty"`
	SequentialOperations []*Operation        `json:"sequentialOperations,omitempty"`
	PostRun              *FinalDefinitions   `json:"postRun,omitempty"`
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
	Request  string                 `json:"request"`
	Function string                 `json:"function"`
	Args     map[string]interface{} `json:"args"`
	Expect   ExpectSpec             `json:"expect"`
	Store    StoreSpec              `json:"store"`
	Change   map[string]interface{} `json:"change"`
}
