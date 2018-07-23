package sgol

import (
	//"fmt"
	//"net/url"
	"os"
	"strings"
	"time"
)

import (
	"github.com/pkg/errors"
	//"github.com/sirupsen/logrus"
	//"github.com/mattn/go-colorable"
)

import (
	//"github.com/spatialcurrent/go-auth-backend/authbackend"
	"github.com/spatialcurrent/go-composite-logger/compositelogger"
	"github.com/spatialcurrent/go-graph/graph"
	//"github.com/spatialcurrent/sgol-codec/codec"
)

type SchemaCommand struct {
	*HttpCommand
}

func (cmd *SchemaCommand) GetName() string {
	return "schema"
}

func (cmd *SchemaCommand) Parse(args []string) error {

	fs := cmd.NewFlagSet(cmd.GetName())

	output_format_help := "Output format.  Options: " + strings.Join(cmd.config.GetFormatIds(), ", ")
	fs.StringVar(&cmd.output_format_text, "f", "", output_format_help)

	//c.sgol_config_path = flagSet.String("c", os.Getenv("SGOL_CONFIG_PATH"), "path to SGOL config.hcl")
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

func (cmd *SchemaCommand) Run(start time.Time, version string) error {

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

	schema, err := graph_backend.Schema(map[string]string{
		"auth_token": cmd.auth_token,
	})
	if err != nil {
		return err
	}

	output_text := ""
	if cmd.output_format.Id == "json" {
		output_text, err = schema.Json()
	} else if cmd.output_format.Id == "yaml" {
		output_text, err = schema.Yaml()
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

func NewSchemaCommand(config *Config) *SchemaCommand {
	return &SchemaCommand{
		HttpCommand: &HttpCommand{
			BasicCommand: &BasicCommand{
				evars:  []string{"SGOL_BACKEND_URL", "SGOL_AUTH_TOKEN"},
				config: config,
			},
		},
	}
}
