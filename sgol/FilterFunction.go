package sgol

type Predicate interface{}

type FilterFunction interface {
	GetName() string
}

type AbstractFilterFunction struct {
	Name string `json:"name" bson:"name" yaml:"name" hcl:"name"`
}

func (ff *AbstractFilterFunction) GetName() string {
	return ff.Name
}

type FilterFunctionCollectionContains struct {
  *AbstractFilterFunction
	PropertyName string `json:"property_name" bson:"property_name" yaml:"property_name" hcl:"property_name"`
	PropertyValue string `json:"property_value" bson:"property_value" yaml:"property_value" hcl:"property_value"`
}
