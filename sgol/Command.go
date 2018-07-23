package sgol

import (
	"time"
)

type Command interface {
	GetName() string
	Parse(args []string) error
	Run(start time.Time, version string) error
}
