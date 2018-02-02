package sgol

import (
	//"os"
	"strings"
	"time"
)

import (
	//"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spatialcurrent/go-simple-serializer/simpleserializer"
)

type QueriesCommand struct {
	*BasicCommand
}

func (cmd *QueriesCommand) GetName() string {
	return "queries"
}

func (cmd *QueriesCommand) Parse(args []string) error {

	fs := cmd.NewFlagSet(cmd.GetName())

	output_format_help := "Output format.  Options: " + strings.Join(cmd.config.GetFormatIds(), ", ")
	fs.StringVar(&cmd.output_format_text, "f", "", output_format_help)

	fs.StringVar(&cmd.output_uri, "output_uri", "stdout", "stdout, stderr, or filepath")
	fs.BoolVar(&cmd.output_append, "append", false, "Append to existing file")
	fs.BoolVar(&cmd.output_overwrite, "overwrite", false, "Overwrite existing file")
	fs.BoolVar(&cmd.verbose, "verbose", false, "Provide verbose output")
	fs.BoolVar(&cmd.dry_run, "dry_run", false, "Connect to destination, but don't import any data.")
	fs.BoolVar(&cmd.version, "version", false, "Version")
	fs.BoolVar(&cmd.help, "help", false, "Print help")

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	if !cmd.help {
		err = cmd.ParseOutputFormat()
		if err != nil {
			return err
		}
	}

	return nil
}

func (cmd *QueriesCommand) Run(log *logrus.Logger, start time.Time, version string) error {

	if cmd.help {
		cmd.PrintHelp(cmd.GetName(), version)
	} else {

		output_format := cmd.output_format.Id

		var output_object interface{}
		if output_format == "toml" {
			output_object_string_map := map[string]interface{}{}
			output_object_string_map[cmd.GetName()] = cmd.config.NamedQueries
			output_object = output_object_string_map
		} else {
			output_object = cmd.config.NamedQueries
		}
		output_text, err := simpleserializer.Serialize(output_object, output_format)
		if err != nil {
			return err
		}

		err = cmd.WriteOutput(cmd.output_uri, output_text)
		if err != nil {
			return err
		}

		elapsed := time.Since(start)
		if cmd.verbose {
			log.Info("Done in " + elapsed.String())
		}
	}

	return nil
}

func NewQueriesCommand(config *Config) *QueriesCommand {
	return &QueriesCommand{
		BasicCommand: &BasicCommand{
			config: config,
		},
	}
}
