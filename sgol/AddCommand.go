package sgol

import (
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"
)

import (
	//"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

import (
	"github.com/spatialcurrent/go-collector/collector"
	"github.com/spatialcurrent/go-composite-logger/compositelogger"
	"github.com/spatialcurrent/go-simple-serializer/simpleserializer"
)

import (
	"github.com/spatialcurrent/go-graph/graph"
	"github.com/spatialcurrent/go-graph/graph/elements"
	"github.com/spatialcurrent/sgol-codec/codec"
)

type AddCommand struct {
	*HttpCommand
	basepath            string
	vertex_template     string
	properties_template string
	element_group       string
	edges_template      string
	limit               int
}

type InputEdge struct {
	With string `json:"with,omitempty" bson:"with,omitempty" yaml:"with,omitempty" hcl:"with,omitempty"`
	*elements.Edge
}

func (cmd *AddCommand) GetName() string {
	return "add"
}

func (cmd *AddCommand) extract(keyChain []string, node interface{}) (interface{}, error) {
	switch node.(type) {
	case map[string]interface{}:
	}
	return nil, errors.New("Could not type node.")
}

func (cmd *AddCommand) CheckBasePath() error {
	if len(cmd.basepath) == 0 {
		return errors.New("Error: Missing basepath.")
	}
	return nil
}

func (cmd *AddCommand) Parse(args []string) error {

	fs := cmd.NewFlagSet(cmd.GetName())

	//c.sgol_config_path = flagSet.String("c", os.Getenv("SGOL_CONFIG_PATH"), "path to SGOL config.hcl")
	fs.StringVar(&cmd.backend_url, "u", os.Getenv("SGOL_BACKEND_URL"), "Backend url.")
	fs.StringVar(&cmd.auth_token, "t", os.Getenv("SGOL_AUTH_TOKEN"), "Authentication token.  Default: environment variable SGOL_AUTH_TOKEN.")
	fs.StringVar(&cmd.basepath, "basepath", "", "Input base path")
	fs.StringVar(&cmd.vertex_template, "vertex", "", "Vertex template")
	fs.StringVar(&cmd.properties_template, "properties", "", "Properties template")
	fs.StringVar(&cmd.element_group, "group", "", "Element group")
	fs.StringVar(&cmd.edges_template, "edges", "", "Edges template")
	//fs.StringVar(&cmd.input_uri, "input_uri", "stdout", "stdout, stderr, or filepath")
	//input_format_help := "Input format.  Options: " + strings.Join(cmd.config.GetFormatIds(), ", ")
	//fs.StringVar(&cmd.input_format_text, "input_format", "", input_format_help)
	fs.StringVar(&cmd.output_uri, "output_uri", "stdout", "stdout, stderr, or filepath")
	output_format_help := "Output format.  Options: " + strings.Join(cmd.config.GetFormatIds(), ", ")
	fs.StringVar(&cmd.output_format_text, "f", "", output_format_help)
	fs.BoolVar(&cmd.output_append, "append", false, "Append to existing file")
	fs.BoolVar(&cmd.output_overwrite, "overwrite", false, "Overwrite existing file")
	fs.BoolVar(&cmd.verbose, "verbose", false, "Provide verbose output")
	fs.BoolVar(&cmd.dry_run, "dry_run", false, "Connect to destination, but don't import any data.")
	fs.IntVar(&cmd.limit, "limit", 0, "Maximum number of elements to process.")
	fs.BoolVar(&cmd.version, "version", false, "Version")
	fs.BoolVar(&cmd.help, "help", false, "Print help")

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	if !cmd.help {

		err = cmd.CheckBasePath()
		if err != nil {
			return err
		}

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

func (cmd *AddCommand) BuildUrl(q string) (string, error) {

	u := cmd.backend_url
	if !strings.HasSuffix(cmd.backend_url, "/") {
		u += "/"
	}
	u += "exec." + cmd.output_format.Extension + "?q=" + url.QueryEscape(q)

	return u, nil

}

func (cmd *AddCommand) RenderVextex(context_base map[string]interface{}, data map[string]interface{}) (string, error) {
	context_entity := map[string]interface{}{
		"Group": cmd.element_group,
		"Id":    data["id"].(string),
		"data":  data,
	}
	for k, v := range context_base {
		context_entity[k] = v
	}
	return RenderTemplate(cmd.vertex_template, context_entity)
}

func (cmd *AddCommand) RenderProperties(context_base map[string]interface{}, data map[string]interface{}) (map[string]interface{}, error) {

	properties := map[string]interface{}{}

	context_entity := map[string]interface{}{
		"Group": cmd.element_group,
		"Id":    data["id"].(string),
		"data":  data,
	}
	for k, v := range context_base {
		context_entity[k] = v
	}

	if len(cmd.properties_template) > 0 {

		properties_json, err := RenderTemplate(cmd.properties_template, context_entity)
		if err != nil {
			return properties, err
		}

		//fmt.Println("Properties JSON:", properties_json)

		err = json.Unmarshal([]byte(properties_json), &properties)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character 'T' looking for beginning of value") {
				return properties, errors.New("Error processing properties_template.  Likely using True instead of true.")
			}
			return properties, err
		}

	}

	for k, v := range data {
		properties[k] = v
	}

	return properties, nil
}

func (cmd *AddCommand) CreateEdges(b graph.Backend, schema graph.Schema, logger *compositelogger.CompositeLogger, context_base map[string]interface{}, vertex string, entity_data map[string]interface{}) ([]elements.Edge, error) {
	edges := make([]elements.Edge, 0)
	context_edges := map[string]interface{}{
		"Vertex": vertex,
	}
	for k, v := range context_base {
		context_edges[k] = v
	}

	edges_input := make([]InputEdge, 0)
	err := json.Unmarshal([]byte(cmd.edges_template), &edges_input)
	if err != nil {
		if strings.Contains(err.Error(), "invalid character 'T' looking for beginning of value") {
			return edges, errors.New("Error processing edges text.  Likely using True instead of true.")
		}
		return edges, err
	}

	for _, d := range edges_input {

		if len(d.With) > 0 {
			if items, ok := entity_data[d.With]; ok {
				for _, item := range items.([]interface{}) {

					context_edges["Item"] = item.(string)

					group, err := RenderTemplate(d.Group, context_edges)
					if err != nil {
						return edges, err
					}

					source, err := RenderTemplate(d.Source, context_edges)
					if err != nil {
						return edges, err
					}

					destination, err := RenderTemplate(d.Destination, context_edges)
					if err != nil {
						return edges, err
					}

					edges = append(edges, *b.NewEdge(group, source, destination, d.Directed, d.Properties))
				}
			}
		} else {
			group, err := RenderTemplate(d.Group, context_edges)
			if err != nil {
				return edges, err
			}

			source, err := RenderTemplate(d.Source, context_edges)
			if err != nil {
				return edges, err
			}

			destination, err := RenderTemplate(d.Destination, context_edges)
			if err != nil {
				return edges, err
			}

			edges = append(edges, *b.NewEdge(group, source, destination, d.Directed, d.Properties))
		}

	}

	return edges, nil
}

func (cmd *AddCommand) LoadElements(b graph.Backend, schema graph.Schema, logger *compositelogger.CompositeLogger) ([]elements.Entity, []elements.Edge, error) {

	entities := make([]elements.Entity, 0)
	edges := make([]elements.Edge, 0)

	filepaths, err := collector.CollectFilepaths(cmd.basepath, []string{"json", "yaml", "yml"}, false, []string{})
	if err != nil {
		return entities, edges, err
	}

	context_base, err := BuildContext(cmd.flagSet.Args(), []string{}, []string{})
	if err != nil {
		return entities, edges, err
	}

	entityPropertyNames := schema.GetEntityPropertyNames(cmd.element_group)

	for i, f := range filepaths {

		if cmd.verbose {
			logger.InfoWithFields("Processing file", map[string]interface{}{
				"file": f,
			})
		}

		buf := make([]byte, 0)
		buf, err := ioutil.ReadFile(f)
		if err != nil {
			logger.WarnWithFields(errors.New("Error: Could not open file for object from path"), map[string]interface{}{
				"f":        f,
				"original": err,
			})
			continue
		}

		data := map[string]interface{}{}

		if strings.HasSuffix(f, ".json") {
			err := json.Unmarshal(buf, &data)
			if err != nil {
				logger.WarnWithFields(errors.New("Error: Could not unmarshal JSON object from path"), map[string]interface{}{
					"f":        f,
					"original": err,
				})
				continue
			}
		} else if strings.HasSuffix(f, ".yaml") || strings.HasSuffix(f, ".yml") {
			err := yaml.Unmarshal(buf, &data)
			if err != nil {
				logger.WarnWithFields(errors.New("Error: Could not unmarshal YAML object from path"), map[string]interface{}{
					"f":        f,
					"original": err,
				})
				continue
			}
			data = graph.StringifyMapKeys(data).(map[string]interface{})
		}

		vertex, err := cmd.RenderVextex(context_base, data)
		if err != nil {
			return entities, edges, err
		}

		properties, err := cmd.RenderProperties(context_base, data)
		if err != nil {
			return entities, edges, err
		}

		entityProperties := map[string]interface{}{}
		for _, name := range entityPropertyNames {
			if v, ok := properties[name]; ok {
				entityProperties[name] = v
			}
		}

		entities = append(entities, *b.NewEntity(cmd.element_group, vertex, entityProperties))

		if cmd.verbose {
			logger.InfoWithFields("Entity loaded", map[string]interface{}{
				"vertex": vertex,
			})
		}

		newEdges, err := cmd.CreateEdges(b, schema, logger, context_base, vertex, data)
		if err != nil {
			return entities, edges, err
		}
		edges = append(edges, newEdges...)

		if cmd.verbose {
			logger.InfoWithFields("Edges loaded", map[string]interface{}{
				"edges": len(newEdges),
			})
		}

		if cmd.limit > 0 {
			if i >= cmd.limit {
				break
			}
		}

	}

	return entities, edges, nil

}

func (cmd *AddCommand) Run(start time.Time, version string) error {

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

	entities, edges, err := cmd.LoadElements(graph_backend, schema, logger)
	if err != nil {
		return err
	}

	op := codec.NewOperationAddWithElements(entities, edges)
	operations := []graph.Operation{op}
	chain := codec.NewOperationChain("chain", operations, "void")

	err = graph_backend.Validate(chain, map[string]string{
		"auth_token": cmd.auth_token,
	})
	if err != nil {
		return err
	}

	//output_bytes, err := yaml.Marshal(chain)
	//if err != nil {
	//	return err
	//}
	//output_text := string(output_bytes)

	qr, err := graph_backend.Execute(chain, map[string]string{
		"auth_token": cmd.auth_token,
	})
	if err != nil {
		return err
	}

	output_text, err := simpleserializer.Serialize(qr.Results, cmd.output_format.Id)
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

func NewAddCommand(config *Config) *AddCommand {
	return &AddCommand{
		HttpCommand: &HttpCommand{
			BasicCommand: &BasicCommand{
				evars:  []string{"SGOL_BACKEND_URL", "SGOL_AUTH_TOKEN"},
				config: config,
			},
		},
	}
}
