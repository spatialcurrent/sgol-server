package sgol

import (
	"encoding/json"
	//"fmt"
	"net/http"
	//"net/url"
	"os"
	"path"
	//"strings"
	//"regexp"
	"strconv"
	"time"
)

import (
	"github.com/gorilla/mux"
	//"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	//"github.com/sirupsen/logrus"
	//"github.com/ttacon/chalk"
	"gopkg.in/yaml.v2"
)

import (
	"github.com/spatialcurrent/sgol-codec/codec"
)

import (
	"github.com/spatialcurrent/go-auth-backend/authbackend"
	"github.com/spatialcurrent/go-composite-logger/compositelogger"
	"github.com/spatialcurrent/go-graph/graph"
	"github.com/spatialcurrent/go-graph/graph/elements"
	//"github.com/spatialcurrent/sgol-codec/codec"
)

type AddElementsInput struct {
	Entities []elements.Entity `json:"entities" bson:"entities" yaml:"entities" hcl:"entities"`
	Edges    []elements.Edge   `json:"edges" bson:"edges" yaml:"edges" hcl:"edges"`
}

type ServeCommand struct {
	*HttpCommand
	server_port int
}

func (cmd *ServeCommand) GetName() string {
	return "serve"
}

func (cmd *ServeCommand) GetParam(r *http.Request, params map[string]string, name string, fallback string) string {
	value := r.URL.Query().Get(name)
	if len(value) == 0 {
		value, ok := params[name]
		if !ok {
			return fallback
		} else {
			return value
		}
	} else {
		return value
	}
}

func (cmd *ServeCommand) Parse(args []string) error {

	fs := cmd.NewFlagSet(cmd.GetName())

	port_default_int := 8005
	if port_default_text := os.Getenv("SGOL_SERVER_PORT"); len(port_default_text) > 0 {
		if i, err := strconv.Atoi(port_default_text); err == nil {
			port_default_int = i
		}
	}
	//c.sgol_config_path = flagSet.String("c", os.Getenv("SGOL_CONFIG_PATH"), "path to SGOL config.hcl")
	fs.IntVar(&cmd.server_port, "p", port_default_int, "Server port.")
	//fs.StringVar(&cmd.backend_url, "u", os.Getenv("SGOL_BACKEND_URL"), "Backend url.")
	fs.BoolVar(&cmd.verbose, "verbose", false, "Provide verbose output")
	fs.BoolVar(&cmd.dry_run, "dry_run", false, "Connect to destination, but don't import any data.")
	fs.BoolVar(&cmd.version, "version", false, "Version")
	fs.BoolVar(&cmd.help, "help", false, "Print help")

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *ServeCommand) Run(start time.Time, version string) error {

	if cmd.help {
		cmd.PrintHelp(cmd.GetName(), version)
	}

	var router = mux.NewRouter()

	logger, err := compositelogger.NewCompositeLogger(cmd.config.Logs)
	if err != nil {
		return err
	}

	var auth_backend authbackend.Backend
	if cmd.config.AuthenticationConfig != nil && cmd.config.AuthenticationConfig.Enabled {
		auth_backend, err = authbackend.LoadFromPlugin(
			cmd.config.AuthenticationConfig.PluginPath,
			cmd.config.AuthenticationConfig.Symbol,
		)
		if err != nil {
			return err
		}
		err = auth_backend.Connect(cmd.config.AuthenticationConfig.Options)
		if err != nil {
			return err
		}
	}

	if cmd.config.GraphBackendConfig == nil {
		return errors.New("Graph config missing from config file.")
	}

	if !cmd.config.GraphBackendConfig.Enabled {
		return errors.New("Graph backend is not enabled.")
	}

	graph_backend_options := cmd.config.GraphBackendConfig.Options

	if sgol_backend_url := os.Getenv("SGOL_BACKEND_URL"); len(sgol_backend_url) > 0 {
		graph_backend_options["url"] = sgol_backend_url
	}

	if sgol_backend_timeout := os.Getenv("SGOL_BACKEND_TIMEOUT"); len(sgol_backend_timeout) > 0 {
		graph_backend_options["timeout"] = sgol_backend_timeout
	}

	graph_backend, err := graph.ConnectToBackend(
		cmd.config.GraphBackendConfig.PluginPath,
		cmd.config.GraphBackendConfig.Symbol,
		graph_backend_options,
	)
	if err != nil {
		return err
	}

	//results := cache.New(60*time.Minute, 1*time.Minute)
	var qc *QueryCacheInstance
	if cmd.config.QueryCache.Enabled {
		qc = NewQueryCache(cmd.config.QueryCache)
		if cmd.verbose {
			logger.Info("Created new query cache.")
		}
	}

	router.Methods("GET").Name("proxy").Path("/{pathtofile:.*}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		pathtofile := cmd.GetParam(r, params, "pathtofile", "")

		filename, ext := ParseFilename(path.Base(pathtofile), true)

		if filename != "exec" {
			logger.WarnWithFields("unkown path", map[string]interface{}{"path": pathtofile})
			w.WriteHeader(404)
			return
		}

		authenticated := false
		var user authbackend.User
		var team authbackend.Team

		if auth_backend != nil {
			authenticated2, user2, team2, err := auth_backend.AuthenticateRequest(r)
			if err != nil {
				logger.Warn(err)
				graph.RespondWithError(w, ext, "Error with Authentication Backend.")
				return
			}

			if !authenticated2 {
				logger.Warn(err)
				graph.RespondWithError(w, ext, "Could not authenticate.")
				return
			}

			authenticated = authenticated2
			user = user2
			team = team2

			logger.InfoWithFields("Authenticated", map[string]interface{}{
				"authenticated": authenticated,
				"user":          user.GetId(),
				"team":          team.GetId(),
			})
		}

		if filename == "exec" {

			q := cmd.GetParam(r, params, "q", "")

			if len(q) == 0 {
				logger.Warn(errors.New("Error: Missing query"))
				w.WriteHeader(500)
				return
			}

			chain, err := cmd.config.Parser.ParseQuery(q)
			if err != nil {
				logger.WarnWithFields(errors.New("Error: Could not parse query into operation chain"), map[string]interface{}{
					"q":             q,
					"original":      err,
					"authenticated": authenticated,
				})
				w.WriteHeader(500)
				return
			}

			hash, err := chain.Hash()
			if err != nil {
				logger.WarnWithFields(errors.New("Error: Could not hash operation chain"), map[string]interface{}{
					"chain": chain,
				})
				w.WriteHeader(500)
				return
			}

			chain_yml, err := yaml.Marshal(chain)
			if err != nil {
				logger.WarnWithFields(errors.New("Error: Can not encode operation chain as yaml."), map[string]interface{}{
					"q":             q,
					"original":      err,
					"authenticated": authenticated,
				})
				w.WriteHeader(500)
				return
			}

			logger.Info("Operation Chain:\n" + string(chain_yml))

			cacheable := false

			qr := graph.QueryResponse{}
			cache_hit := false
			//output_text := ""
			if qc != nil {
				cacheable = qc.CacheOperationChain(chain)

				if cmd.verbose {
					if cacheable {
						logger.Info("Cacheable")
					} else {
						logger.Info("Not cacheable")
					}
				}

				if cacheable {
					if data, found, err := qc.GetHash(hash); err == nil && found {
						if cmd.verbose {
							logger.InfoWithFields("Cache hit!", map[string]interface{}{
								"chain": chain,
								"hash":  hash,
							})
						}
						qr = data
						cache_hit = true

						qr.WriteToResponse(w, ext)
					}
				}
			}

			if !cache_hit {

				options := map[string]string{
					"auth_token": r.Header.Get("X-Auth-Token"),
					"cookie":     r.Header.Get("Cookie"),
				}

				err = graph_backend.Validate(chain, options)
				if err != nil {
					logger.WarnWithFields("Invalid operation chain", map[string]interface{}{"chain": chain, "message": err})
					w.WriteHeader(600)
					return
				}

				qr, err = graph_backend.Execute(chain, options)
				if err != nil {
					logger.WarnWithFields(err, map[string]interface{}{
						"q":             q,
						"original":      err,
						"authenticated": authenticated,
					})
					w.WriteHeader(500)
				}

				if qc != nil && cacheable {
					if cmd.verbose {
						logger.InfoWithFields("Saving to cache", map[string]interface{}{
							"chain": chain,
							"hash":  hash,
						})
					}
					qc.SetHash(hash, qr)
				}

				qr.WriteToResponse(w, ext)

			}

		} else {
			logger.WarnWithFields("unkown path", map[string]interface{}{"path": pathtofile})
			w.WriteHeader(404)
			return
		}

	})

	router.Methods("POST").Name("proxy").Path("/{pathtofile:.*}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		pathtofile := cmd.GetParam(r, params, "pathtofile", "")

		filename, ext := ParseFilename(path.Base(pathtofile), true)

		if filename != "add" {
			logger.WarnWithFields("unkown path", map[string]interface{}{"path": pathtofile})
			w.WriteHeader(404)
			return
		}

		authenticated := false
		var user authbackend.User
		var team authbackend.Team

		if auth_backend != nil {
			authenticated2, user2, team2, err := auth_backend.AuthenticateRequest(r)
			if err != nil {
				logger.Warn(err)
				w.WriteHeader(500)
				return
			}

			if !authenticated2 {
				logger.Warn(err)
				w.WriteHeader(500)
				return
			}

			authenticated = authenticated2
			user = user2
			team = team2

			logger.InfoWithFields("Authenticated", map[string]interface{}{
				"authenticated": authenticated,
				"user":          user.GetId(),
				"team":          team.GetId(),
			})
		}

		if filename == "add" {

			options := map[string]string{
				"auth_token": r.Header.Get("X-Auth-Token"),
				"cookie":     r.Header.Get("Cookie"),
			}

			decoder := json.NewDecoder(r.Body)
			input := AddElementsInput{}
			err := decoder.Decode(&input)
			if err != nil {
				logger.WarnWithFields("Invalid input", map[string]interface{}{"message": err})
				w.WriteHeader(600)
				return
			}
			defer r.Body.Close()

			op := codec.NewOperationAdd(false)
			op.Entities = input.Entities
			op.Edges = input.Edges
			operations := []graph.Operation{op}
			chain := codec.NewOperationChain("chain", operations, "void")

			err = graph_backend.Validate(chain, options)
			if err != nil {
				logger.WarnWithFields("Invalid operation chain", map[string]interface{}{"chain": chain, "message": err})
				w.WriteHeader(600)
				return
			}

			qr, err := graph_backend.Execute(chain, options)
			if err != nil {
				logger.WarnWithFields(err, map[string]interface{}{
					"original":      err,
					"authenticated": authenticated,
				})
				w.WriteHeader(500)
			}

			qr.WriteToResponse(w, ext)

		} else {
			logger.WarnWithFields("unkown path", map[string]interface{}{"path": pathtofile})
			w.WriteHeader(404)
			return
		}

	})

	logger.Info("Listening on port " + strconv.Itoa(cmd.server_port))
	logger.Fatal(http.ListenAndServe(":"+strconv.Itoa(cmd.server_port), router))

	elapsed := time.Since(start)
	if cmd.verbose {
		logger.Info("Exited after " + elapsed.String())
	}

	return nil
}

func NewServeCommand(config *Config) *ServeCommand {
	return &ServeCommand{
		HttpCommand: &HttpCommand{
			BasicCommand: &BasicCommand{
				config: config,
			},
		},
	}
}
