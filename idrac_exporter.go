package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

func collectMetrics(target string) (string, bool) {
	host, ok := config.Hosts[target]
	if !ok {
		config.Hosts[target] = new(HostConfig)
		host = config.Hosts[target]
		host.Token = config.Hosts["default"].Token
		host.Hostname = target
		host.Initialized = false
		host.Retries = 0
	}

	if !host.Initialized {
		ok = redfishFindAllEndpoints(host)
		host.Retries++
		host.Initialized = (ok || (host.Retries >= config.Retries))
		host.Reachable = ok
	}

	if !host.Reachable {
		return "", false
	}

	metricsClear(host)

	if collectSystem {
		ok = redfishSystem(host)
		if !ok {
			return "", false
		}
	}

	if collectSensors {
		ok = redfishSensors(host)
		if !ok {
			return "", false
		}
	}

	if collectSEL {
		ok = redfishSEL(host)
		if !ok {
			return "", false
		}
	}

	if collectPower {
		ok = redfishPower(host)
		if !ok {
			return "", false
		}
	}

	return metricsGet(host), true
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	args := r.URL.Query()
	target, ok := args["target"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metrics, ok := collectMetrics(target[0])
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, metrics)
}

type JsonError struct {
	Error string `json:"error"`
}

func targetsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if config.Ome == nil {
		j, _ := json.Marshal(JsonError{Error: "could not fetch targets: OpenManage Enterprise service discovery not enabled"})
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprint(w, string(j))
		return
	}

	sc := ScrapeConfig{
		Targets:        []string{},
		Labels:         map[string]string{},
		targetsFetched: 0,
	}

	ok := omeFindAllTargets(config.Ome, &sc)
	if !ok {
		j, _ := json.Marshal(JsonError{Error: "could not fetch targets"})
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, string(j))
		return
	}
	w.WriteHeader(http.StatusOK)
	j, err := json.Marshal(sc)
	if err != nil {
		log.Print(err)
		fmt.Print(w, "{'error': 'could not render targets. check the logs'}")
		return
	}
	fmt.Fprint(w, []string{string(j)})
}

func main() {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting...")
	var configFile string
	flag.StringVar(&configFile, "config", "/etc/prometheus/idrac.yml", "path to idrac exporter configuration file")
	flag.Parse()
	readConfigFile(configFile)

	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/targets", targetsHandler)
	http.HandleFunc("/metrics", metricsHandler)
	bind := fmt.Sprintf("%s:%d", config.Address, config.Port)
	logger.Println("Server is ready to handle requests at", bind)
	err := http.ListenAndServe(bind, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
	if err != nil {
		logger.Fatalf("Could not listen on %s: %v\n", bind, err)
	}
}
