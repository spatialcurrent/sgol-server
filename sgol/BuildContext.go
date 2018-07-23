package sgol

import (
	"strings"
)

import (
	"github.com/pkg/errors"
)

func BuildContext(args []string, required []string, optional []string) (map[string]interface{}, error) {
	ctx := map[string]interface{}{}
	for _, a := range args {
		parts := strings.SplitN(a, "=", 2)
		ctx[parts[0]] = parts[1]
	}
	for _, key := range required {
		if _, ok := ctx[key]; !ok {
			return ctx, errors.New("Error: Missing required parameter " + key + ".")
		}
	}
	for key, _ := range ctx {
		if (!StringSliceContains(required, key)) && (!StringSliceContains(optional, key)) {
			return ctx, errors.New("Error: Unknown parameter " + key + ".")
		}
	}

	return ctx, nil
}
