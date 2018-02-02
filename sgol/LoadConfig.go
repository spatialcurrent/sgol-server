package sgol

import (
	"fmt"
	"io/ioutil"
)

import (
	"github.com/hashicorp/hcl"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

func LoadConfig(path string, verbose bool) (*Config, error) {

	path_expanded, err := homedir.Expand(path)
	if err != nil {
		return nil, errors.New("Error: Could not expand home directory for path " + path + ".")
	}

	result := &Config{}

	d, err := ioutil.ReadFile(path_expanded)
	if err != nil {
		return result, errors.New(fmt.Sprintf("Error reading %s: %s", path, err))
	}

	obj, err := hcl.Parse(string(d))
	if err != nil {
		return result, errors.New(fmt.Sprintf("Error parsing %s: %s", path, err))
	}

	if err := hcl.DecodeObject(&result, obj); err != nil {
		return result, errors.New(fmt.Sprintf("Error parsing %s: %s", path, err))
	}

	return result, nil

}
