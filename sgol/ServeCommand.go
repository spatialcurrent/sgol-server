package sgol

import (
	"fmt"
	"net/http"
	"net/url"
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
	"github.com/ttacon/chalk"
	//"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

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
	fs.StringVar(&cmd.backend_url, "u", os.Getenv("SGOL_BACKEND_URL"), "Backend url.")
	fs.BoolVar(&cmd.verbose, "verbose", false, "Provide verbose output")
	fs.BoolVar(&cmd.dry_run, "dry_run", false, "Connect to destination, but don't import any data.")
	fs.BoolVar(&cmd.version, "version", false, "Version")
	fs.BoolVar(&cmd.help, "help", false, "Print help")

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	err = cmd.ParseBackendUrl()
	if err != nil {
		return err
	}

	return nil
}

func (cmd *ServeCommand) Run(log *logrus.Logger, start time.Time, version string) error {

	if cmd.help {
		cmd.PrintHelp(cmd.GetName(), version)
	}

	contentTypes := map[string]string{
		"css":  "text/css",
		"json": "application/json",
		"js":   "application/javascript; charset=utf-8",
		"png":  "image/png",
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"jp2":  "image/jpeg",
	}

	var router = mux.NewRouter()

	//results := cache.New(60*time.Minute, 1*time.Minute)
	var qc *QueryCacheInstance
	if cmd.config.QueryCache.Enabled {
		qc = NewQueryCache(cmd.config.QueryCache)
		if cmd.verbose {
			log.Println("Created new query cache.")
		}
	}

	router.Methods("GET").Name("proxy").Path("/{pathtofile:.*}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		pathtofile := cmd.GetParam(r, params, "pathtofile", "")
		q := cmd.GetParam(r, params, "q", "")

		filename, ext := ParseFilename(path.Base(pathtofile), true)

		contentType := "text/plain"
		if x, ok := contentTypes[ext]; ok {
			contentType = x
		}

		fmt.Println("filename:", filename)
		fmt.Println("ext:", ext)

		if filename == "exec" {

			if len(q) == 0 {
				log.Println(chalk.Red, "Error: Missing query", chalk.Reset)
				w.WriteHeader(500)
				return
			}

			chain, err := cmd.config.Parser.ParseQuery(q)
			if err != nil {
				log.Println(chalk.Red, "Error: Could not parse query into operation chain", chalk.Reset)
				log.Println(chalk.Red, err, chalk.Reset)
				w.WriteHeader(500)
				return
			}

			chain_yml, err := yaml.Marshal(chain)
			if err != nil {
				log.Println("Error: Can not encode operation chain as yaml.")
				w.WriteHeader(500)
				return
			}
			log.Println("Operation Chain:\n", string(chain_yml))

			url := cmd.backend_url + "/exec." + ext + "?q=" + url.QueryEscape(q)
			cookie := r.Header.Get("Cookie")
			auth_token := r.Header.Get("X-Auth-Token")

			if cmd.verbose {
				//	log.Println(r.Header)
				log.Println(chalk.Green, "url:", url, chalk.Reset)
				//	log.Println(chalk.Green, "cookie:", cookie, chalk.Reset)
				//	log.Println(chalk.Green, "auth_token:", auth_token, chalk.Reset)
			}

			cacheable := false

			output_text := ""
			if qc != nil {
				cacheable = qc.CacheOperationChain(chain)
				if cmd.verbose {
					fmt.Println("Op Chain Cacheable", cacheable)
				}
				if cacheable {
					if data, found, err := qc.GetOperationChain(chain); err == nil && found {
						if cmd.verbose {
							log.Println(chalk.Green, "Cache hit!", chalk.Reset)
						}
						output_text = data.(string)
					}
				}
			}

			if len(output_text) == 0 {
				if len(cookie) > 0 {
					response_text, err := cmd.MakeRequestWithCookie(url, cookie, cmd.verbose)
					if err != nil {
						log.Println(chalk.Red, err, chalk.Reset)
						w.WriteHeader(500)
						return
					}
					output_text = response_text
				} else if len(auth_token) > 0 {
					response_text, err := cmd.MakeRequestWithAuthToken(url, auth_token, cmd.verbose)
					if err != nil {
						log.Println(chalk.Red, err, chalk.Reset)
						w.WriteHeader(500)
						return
					}
					output_text = response_text
				}

				//if cmd.verbose {
				//	log.Println(chalk.Green, "output_text:", output_text, chalk.Reset)
				//}
			}

			if len(output_text) > 0 {
				if qc != nil && cacheable {
				  qc.SetOperationChain(chain, output_text)
				}
				w.Header().Set("Content-Type", contentType)
				fmt.Fprintf(w, output_text)
			} else {
				log.Println(chalk.Red, "Response from backend was empty.", chalk.Reset)
				w.WriteHeader(500)
				return
			}

		} else {
			log.Println(chalk.Red, "Unknown path", pathtofile, chalk.Reset)
			w.WriteHeader(404)
			return
		}

	})

	log.Println(chalk.Cyan, "Listening on port", cmd.server_port, chalk.Reset)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(cmd.server_port), router))

	elapsed := time.Since(start)
	if cmd.verbose {
		log.Info("Exited after " + elapsed.String())
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
