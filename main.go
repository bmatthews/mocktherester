package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var flags = struct {
	port  string
	mocks string
}{
	port:  "port",
	mocks: "mocks",
}

var logs = make(map[string][]*Log)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// using standard library "flag" package
	flag.Int(flags.port, 8080, "port number")
	flag.String(flags.mocks, "mocks.yaml", "mock file")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatal("binding flags")
	}

	port := viper.GetInt(flags.port)
	mocks := viper.GetString(flags.mocks)

	r := mux.NewRouter()

	config := getRoutes(mocks)
	for _, v := range config.Routes {
		r.HandleFunc(v.Path, func(w http.ResponseWriter, r *http.Request) {
			l := func(w http.ResponseWriter, r *http.Request) {
				r.ParseForm()
				logs[v.Name] = append(logs[v.Name], &Log{
					Path:        r.URL.String(),
					RequestTime: time.Now(),
					Headers:     r.Header,
					Form:        r.Form,
					Method:      r.Method,
				})
				w.WriteHeader(v.Result.Code)
				sendJSON(w, v.Result.Data)
			}
			if v.Auth.Type == BASIC {
				fmt.Println("handling basic auth")
				BasicAuth(l, v.Auth.Username, v.Auth.Password)(w, r)
			} else {
				l(w, r)
			}
		}).Methods(string(v.Method))
	}

	r.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		sendJSON(w, logs)
	})

	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
	}()

	log.Printf("server started on localhost:%d\n", port)
	<-stop
}

// BasicAuth wraps a handler requiring HTTP basic auth for it
func BasicAuth(handler http.HandlerFunc, username, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if (ok && (user != username || pass != password)) || !ok { // dont care about timings of the checks
			fmt.Println("failed auth check")
			w.WriteHeader(401)
			_, err := w.Write([]byte("Unauthorised.\n"))
			if err != nil {
				log.Println(err)
			}
			return
		}
		fmt.Println("passed auth check")
		handler(w, r)
	}
}

func sendJSON(w http.ResponseWriter, p interface{}) {
	err := json.NewEncoder(w).Encode(p)
	if err != nil {
		http.Error(w, "marshalling json response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func getRoutes(mocks string) *Config {
	routes := &Config{}
	yamlFile, err := ioutil.ReadFile(mocks)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, routes)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	fmt.Println(routes)
	return routes
}

type Method string

const (
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
	GET    = "GET"
)

type Config struct {
	Routes []Route `yaml:"routes"`
}

type Route struct {
	Name   string `yaml:"name"`
	Method Method `yaml:"method"`
	Path   string `yaml:"path"`
	Auth   Auth   `yaml:"auth"`
	Result Result `yaml:"result"`
}

type AuthType string

const (
	BASIC = "BASIC"
)

type Auth struct {
	Type     AuthType `yaml:"type"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
}

type Result struct {
	Code int                    `yaml:"code"`
	Data map[string]interface{} `yaml:"data"`
}

type Log struct {
	Path        string              `json:"path"`
	RequestTime time.Time           `json:"reqeust_time"`
	Method      string              `json:"method"`
	Headers     map[string][]string `json:"headers"`
	Form        map[string][]string `json:"form"`
}
