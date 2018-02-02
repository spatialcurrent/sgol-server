package sgol

type Predicate interface{}

type FilterFunction struct {
	Selection []string
	Predicate Predicate
}
