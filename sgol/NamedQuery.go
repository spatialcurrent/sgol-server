package sgol

type NamedQuery struct {
	Name     string   `json:"name" hcl:"name"`
	Sgol     string   `json:"sgol" hcl:"sgol"`
	Required []string `json:"required,omitempty" hcl:"required,omitempty"`
	Optional []string `json:"optional,omitempty" hcl:"optional,omitempty"`
}
