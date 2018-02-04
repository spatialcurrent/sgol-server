package sgol

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

import (
	"github.com/pkg/errors"
)

type BasicCommand struct {
	evars []string
	config             *Config
	flagSet            *flag.FlagSet
	output_uri         string
	output_format_text string
	output_format      *Format
	output_append      bool
	output_overwrite   bool
	verbose            bool
	dry_run            bool
	version            bool
	help               bool
}

func (cmd *BasicCommand) GetEnvironmentVariables() []string {
	return cmd.evars
}

func (cmd *BasicCommand) NewFlagSet(name string) *flag.FlagSet {
	cmd.flagSet = flag.NewFlagSet(name, flag.ExitOnError)
	return cmd.flagSet
}

func (cmd *BasicCommand) ParseOutputFormat() error {
	if len(cmd.output_format_text) == 0 {
		return errors.New("Error: missing output format")
	}

	cmd.output_format = &Format{}
	output_format_text_lc := strings.ToLower(cmd.output_format_text)
	for _, format := range cmd.config.Formats {
		for _, alias := range format.Aliases {
			if output_format_text_lc == alias {
				cmd.output_format = format
				break
			}
		}
		if len(cmd.output_format.Id) > 0 {
			break
		}
	}

	if len(cmd.output_format.Id) == 0 {
		return errors.New("Error: Could not resolve output format")
	}

	return nil
}

func (cmd *BasicCommand) PrintHelp(name string, version string) {
	fmt.Println("SGOL CLI - Version " + version)
	fmt.Println("Usage: sgol " + name)
	fmt.Println("")
	fmt.Println("Environment Variables: " + strings.Join(cmd.evars, ", ")+"\n")
	fmt.Println("Options:")
	cmd.flagSet.PrintDefaults()
}

func (cmd *BasicCommand) WriteOutput(output_uri string, output_text string) error {

	if output_uri == "stdout" {
		fmt.Println(output_text)
	} else if output_uri == "stderr" {
		fmt.Fprintf(os.Stderr, output_text)
	} else {
		output_file := &os.File{}
		if cmd.output_append {
			f, err := os.OpenFile(output_uri, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				return err
			}
			output_file = f
		} else {
			f, err := os.OpenFile(output_uri, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			output_file = f
		}
		w := bufio.NewWriter(output_file)
		w.WriteString(output_text)
		w.Flush()
	}
	return nil
}
