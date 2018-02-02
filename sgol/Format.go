package sgol

type Format struct {
	Id        string   `csv:"id" json:"id" hcl:"id"`
	Extension string   `csv:"extension" json:"extension" hcl:"extension"`
	Aliases   []string `csv:"aliases" json:"aliases" hcl:"aliases"`
}
