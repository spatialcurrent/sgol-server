package sgol

type Operation interface{}

type AbstractOperation struct{}

type OperationInit struct {
	*AbstractOperation
	Key string
}

type OperationAdd struct {
	*AbstractOperation
	Key string
}

type OperationDiscard struct {
	*AbstractOperation
	Key string
}

type OperationRelate struct {
	*AbstractOperation
	Keys []string
}

type OperationLimit struct {
	*AbstractOperation
	Limit int
}

type OperationSelect struct {
	*AbstractOperation
	Seeds                 []string
	Entities              []string
	Edges                 []string
	FilterFunctions       map[string]string
	UpdateKey             string
	UpdateFilterFunctions map[string][]FilterFunction
}

type OperationNav struct {
	*AbstractOperation
	Source                  string
	Destination             string
	Entities                []string
	Edges                   []string
	Direction               string
	EdgeIdentifierToExtract string
	SeedMatching            string
	Seeds                   []string
	Depth                   int
	FilterFunctions         map[string]string
	UpdateKey               string
	UpdateFilterFunctions   map[string][]FilterFunction
}

type OperationHas struct {
	*AbstractOperation
}

type OperationFetch struct {
	*AbstractOperation
	Key string
}

type OperationSeed struct {
	*AbstractOperation
}

type OperationRun struct {
	*AbstractOperation
	Operations []string
}
