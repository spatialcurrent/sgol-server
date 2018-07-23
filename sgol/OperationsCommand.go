package sgol

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

import (
	"github.com/spatialcurrent/go-composite-logger/compositelogger"
	"github.com/spatialcurrent/go-graph/graph"
)

type OperationsCommand struct {
	*HttpCommand
}

func (cmd *OperationsCommand) GetName() string {
	return "operations"
}

func (cmd *OperationsCommand) Parse(args []string) error {

	fs := cmd.NewFlagSet(cmd.GetName())

	output_format_help := "Output format.  Options: " + strings.Join(cmd.config.GetFormatIds(), ", ")
	fs.StringVar(&cmd.output_format_text, "f", "", output_format_help)

	fs.StringVar(&cmd.backend_url, "u", os.Getenv("SGOL_BACKEND_URL"), "Backend url.")
	fs.StringVar(&cmd.auth_token, "t", os.Getenv("SGOL_AUTH_TOKEN"), "Authentication token.  Default: environment variable SGOL_AUTH_TOKEN.")
	fs.StringVar(&cmd.output_uri, "output_uri", "stdout", "stdout, stderr, or filepath")
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

		err = cmd.ParseBackendUrl()
		if err != nil {
			return err
		}

		err = cmd.ParseAuthToken()
		if err != nil {
			return err
		}

		err = cmd.ParseOutputFormat()
		if err != nil {
			return err
		}

	}

	return nil
}

func (cmd *OperationsCommand) Run(start time.Time, version string) error {

	if cmd.help {
		cmd.PrintHelp(cmd.GetName(), version)
	}

	logger, err := compositelogger.NewCompositeLogger(cmd.config.Logs)
	if err != nil {
		return err
	}

	if cmd.config.GraphBackendConfig == nil {
		return errors.New("Graph config missing from config file.")
	}

	if !cmd.config.GraphBackendConfig.Enabled {
		return errors.New("Graph backend is not enabled.")
	}

	graph_backend, err := graph.ConnectToBackend(
		cmd.config.GraphBackendConfig.PluginPath,
		cmd.config.GraphBackendConfig.Symbol,
		cmd.config.GraphBackendConfig.Options,
	)
	if err != nil {
		return err
	}

	definitions, err := graph_backend.OperationDefinitions(map[string]string{
		"auth_token": cmd.auth_token,
	})
	if err != nil {
		return err
	}

	output_text := ""
	if cmd.output_format.Id == "json" {
		b, err := json.Marshal(definitions)
		if err != nil {
			return err
		}
		output_text = string(b)
	} else if cmd.output_format.Id == "yaml" {
		b, err := yaml.Marshal(definitions)
		if err != nil {
			return err
		}
		output_text = string(b)
	}
	if err != nil {
		return err
	}

	err = cmd.WriteOutput(cmd.output_uri, output_text)
	if err != nil {
		return err
	}

	elapsed := time.Since(start)
	if cmd.verbose {
		logger.Info("Done in " + elapsed.String())
	}

	return nil
}

func NewOperationsCommand(config *Config) *OperationsCommand {
	return &OperationsCommand{
		HttpCommand: &HttpCommand{
			BasicCommand: &BasicCommand{
				evars:  []string{"SGOL_BACKEND_URL", "SGOL_AUTH_TOKEN"},
				config: config,
			},
		},
	}
}
