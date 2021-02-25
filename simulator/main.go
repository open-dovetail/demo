/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/golang/glog"
	"github.com/open-dovetail/demo/simulator/impl"
)

var configFile, httpPort string

func init() {
	flag.StringVar(&httpPort, "port", "7980", "HTTP REST service listen port")
	flag.StringVar(&configFile, "config", "./config.json", "Server configuration file")
}

// Starts simulator service that listens to HTTP service requests.
// Turn on verbose logging using option -v 2
// Log to stderr using option -logtostderr
// or log to specified file using option -log_dir="mylogfile"

// send sample request
// curl -X PUT -H "Content-Type: application/json" -d @package.json http://localhost:7980/packages/create
// curl -X PUT -H "Content-Type: application/json" http://localhost:7980/packages/pickup?uid=4730f2294a6156c8
// curl -X GET -H "Content-Type: application/json" http://localhost:7980/packages/timeline?uid=4730f2294a6156c8

func main() {
	flag.Parse()
	if flag.Lookup("logtostderr").Value.String() != "true" {
		// Set folder for log files
		if flag.Lookup("log_dir").Value.String() == "" {
			flag.Lookup("log_dir").Value.Set("./log")
		}
		if err := os.MkdirAll(flag.Lookup("log_dir").Value.String(), 0777); err != nil {
			fmt.Printf("Error creating log folder %s: %+v\n", flag.Lookup("log_dir").Value.String(), err)
			flag.Lookup("logtostderr").Value.Set("true")
		}
	}

	// configure carriers and routes
	if err := impl.Initialize(configFile); err != nil {
		glog.Error(err)
		panic(err)
	}
	graph, err := impl.GetTGConnection()
	if err != nil {
		glog.Error(err)
		panic(err)
	}

	// initalize graph only if carriers have not been created yet
	for k := range impl.Carriers {
		query := fmt.Sprintf("gremlin://g.V().has('Carrier', 'name', '%s');", k)
		result, err := graph.Query(query)
		if err != nil {
			glog.Error(err)
			panic(err)
		}
		if len(result) == 0 {
			if err = impl.InitializeGraph(graph); err != nil {
				glog.Error(err)
				panic(err)
			}
			break
		}
	}

	// start HTTP listener
	http.HandleFunc("/", handler)
	glog.Info("Starting HTTP listener on port ", httpPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", httpPort), nil); err != nil {
		glog.Error(err)
		panic(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		resp, status, err := handleQueryRequest(r)
		if err != nil {
			http.Error(w, err.Error(), status)
		}
		w.WriteHeader(status)
		w.Write(resp)
	case "PUT", "POST":
		resp, status, err := handleShippingRequest(r)
		if err != nil {
			http.Error(w, err.Error(), status)
		}
		w.WriteHeader(status)
		w.Write(resp)
	default:
		http.Error(w, "Method is not supported", http.StatusBadRequest)
	}
}

func handleShippingRequest(r *http.Request) ([]byte, int, error) {
	fmt.Println("handling shipping")
	if r.URL.Path == "/packages/create" {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		glog.Info("Create shipping label ", string(data))
		resp, err := impl.PrintShippingLabel(string(data))
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return resp, http.StatusOK, nil
	} else if r.URL.Path == "/packages/pickup" {
		uid := r.URL.Query().Get("uid")
		if len(uid) == 0 {
			return nil, http.StatusBadRequest, errors.New("package uid is not specified as query parameter")
		}
		glog.Info("pickup package", uid)
		err := impl.PickupPackage(uid)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return []byte("pikup and delivery completed for package " + uid), http.StatusOK, nil
	}
	return []byte("to be implemented"), http.StatusOK, nil
}

func handleQueryRequest(r *http.Request) ([]byte, int, error) {
	fmt.Println("handling query")
	if r.URL.Path == "/packages/timeline" {
		uid := r.URL.Query().Get("uid")
		if len(uid) == 0 {
			return nil, http.StatusBadRequest, errors.New("package uid is not specified as query parameter")
		}
		glog.Info("timeline of package", uid)
		data, err := impl.QueryPackageTimeline(uid)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return data, http.StatusOK, nil
	}
	return []byte("to be implemented"), http.StatusOK, nil
}
