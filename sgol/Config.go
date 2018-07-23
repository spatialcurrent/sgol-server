package sgol

import (
	"github.com/spatialcurrent/go-composite-logger/compositelogger"
	"github.com/spatialcurrent/sgol-codec/codec"
)

type ClientConfig struct {
	Token string `json:"token" hcl:"token"`
}

type AuthenticationConfig struct {
	Enabled    bool              `json:"enabled" bson:"enabled" yaml:"enabled" hcl:"enabled"`
	PluginPath string            `json:"plugin" bson:"plugin" yaml:"plugin" hcl:"plugin"`
	Symbol     string            `json:"symbol" bson:"symbol" yaml:"symbol" hcl:"symbol"`
	Options    map[string]string `json:"options" bson:"options" yaml:"options" hcl:"options"`
}

type GraphBackendConfig struct {
	Enabled    bool              `json:"enabled" bson:"enabled" yaml:"enabled" hcl:"enabled"`
	PluginPath string            `json:"plugin" bson:"plugin" yaml:"plugin" hcl:"plugin"`
	Symbol     string            `json:"symbol" bson:"symbol" yaml:"symbol" hcl:"symbol"`
	Options    map[string]string `json:"options" bson:"options" yaml:"options" hcl:"options"`
}

type Config struct {
	Client               *ClientConfig                `hcl:"client"`
	Server               *Server                      `hcl:"server"`
	AuthenticationConfig *AuthenticationConfig        `hcl:"auth"`
	Logs                 []*compositelogger.LogConfig `hcl:"logs"`
	GraphBackendConfig   *GraphBackendConfig          `hcl:"backend"`
	Parser               *codec.Parser                `hcl:"parser"`
	QueryCache           *QueryCache                  `hcl:"cache"`
	Formats              []*Format                    `hcl:"formats"`
	NamedQueries         []*codec.NamedQuery          `hcl:"queries"`
}

func (config *Config) GetFormatIds() []string {
	formatIds := make([]string, len(config.Formats))
	for i := 0; i < len(formatIds); i++ {
		formatIds[i] = config.Formats[i].Id
	}
	return formatIds
}
