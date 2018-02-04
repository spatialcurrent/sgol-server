package sgol

type Operation interface{
	GetTypeName() string
}

type AbstractOperation struct{
	Type string `json:"type" bson:"type" yaml:"type" hcl:"type"`
}

func (op AbstractOperation) GetTypeName() string {
  return op.Type
}

type AbstractOperationKey struct {
	*AbstractOperation
  Key string `json:"key" bson:"key" yaml:"key" hcl:"key"`
}

type OperationInit struct {
	*AbstractOperationKey
}

type OperationAdd struct {
	*AbstractOperationKey
}

type OperationDiscard struct {
	*AbstractOperationKey
}

type OperationRelate struct {
	*AbstractOperation
	Keys []string `json:"keys" bson:"keys" yaml:"keys" hcl:"keys"`
}

type OperationLimit struct {
	*AbstractOperation
	Limit int `json:"limit" bson:"limit" yaml:"limit" hcl:"limit"`
}

type OperationSelect struct {
	*AbstractOperation
	Seeds                 []string `json:"seeds" bson:"seeds" yaml:"seeds" hcl:"seeds"`
	Entities              []string `json:"entities" bson:"entities" yaml:"entities" hcl:"entities"`
	Edges                 []string `json:"edges" bson:"edges" yaml:"edges" hcl:"edges"`
	FilterFunctions       map[string][]FilterFunction `json:"filter_functions" bson:"filter_functions" yaml:"filter_functions" hcl:"filter_functions"`
	UpdateKey             string `json:"update_key" bson:"update_key" yaml:"update_key" hcl:"update_key"`
	UpdateFilterFunctions map[string][]FilterFunction `json:"update_filter_functions" bson:"update_filter_functions" yaml:"update_filter_functions" hcl:"update_filter_functions"`
}

type OperationNav struct {
	*AbstractOperation
	Seeds                 []string `json:"seeds" bson:"seeds" yaml:"seeds" hcl:"seeds"`
	Source                  string `json:"source" bson:"source" yaml:"source" hcl:"source"`
	Destination             string `json:"destination" bson:"destination" yaml:"destination" hcl:"destination"`
	Entities              []string `json:"entities" bson:"entities" yaml:"entities" hcl:"entities"`
	Edges                 []string `json:"edges" bson:"edges" yaml:"edges" hcl:"edges"`
	Direction               string `json:"direction" bson:"direction" yaml:"direction" hcl:"direction"`
	EdgeIdentifierToExtract string `json:"edgeIdentifierToExtract" bson:"edgeIdentifierToExtract" yaml:"edgeIdentifierToExtract" hcl:"edgeIdentifierToExtract"`
	SeedMatching            string `json:"seedMatching" bson:"seedMatching" yaml:"seedMatching" hcl:"seedMatching"`
	Depth                   int `json:"depth" bson:"depth" yaml:"depth" hcl:"depth"`
	FilterFunctions       map[string][]FilterFunction `json:"filter_functions" bson:"filter_functions" yaml:"filter_functions" hcl:"filter_functions"`
	UpdateKey             string `json:"update_key" bson:"update_key" yaml:"update_key" hcl:"update_key"`
	UpdateFilterFunctions map[string][]FilterFunction `json:"update_filter_functions" bson:"update_filter_functions" yaml:"update_filter_functions" hcl:"update_filter_functions"`
}

type OperationHas struct {
	*AbstractOperation
}

type OperationFetch struct {
	*AbstractOperationKey
}

type OperationSeed struct {
	*AbstractOperation
}

type OperationRun struct {
	*AbstractOperation
	Operations []string `json:"operations" bson:"operations" yaml:"operations" hcl:"operations"`
}
