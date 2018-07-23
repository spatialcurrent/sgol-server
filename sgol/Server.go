package sgol

type Server struct {
	Port                 int    `json:"port" hcl:"port"`
	AuthenticationPlugin string `json:"auth_plugin" bson:"auth_plugin" yaml:"auth_plugin" hcl:"auth_plugin"`
}
