package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	parser "github.com/randomowo/servermock-go/config_parser"
)

type MockServerHandler struct {
	config *parser.ConfigFile
}

func (h *MockServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		ok             bool
		routeConfig    parser.RouteConfig
		responseConfig parser.ResponseConfig
	)

	routeConfig, ok = (*h.config)[r.URL.Path]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	responseConfig, ok = routeConfig[r.Method]
	if !ok {
		responseConfig, ok = routeConfig[parser.DEF]
		if !ok {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}

	if responseConfig.Body.Echo {
		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err == nil {
			w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
			w.WriteHeader(responseConfig.Code)
			w.Write(body)
		}
	} else {
		w.Header().Set("Content-Type", responseConfig.Body.ContentType)
		w.WriteHeader(responseConfig.Code)
		_, _ = w.Write(responseConfig.Body.Value.([]byte))
	}
}

func main() {
	fileName := flag.String("config", os.Getenv("CONFIG_FILE"), "path to config file")
	flag.Parse()

	fmt.Println("filename:", *fileName)
	handler := &MockServerHandler{
		config: parser.ParseConfig(*fileName),
	}

	server := http.Server{
		Addr:    ":http",
		Handler: handler,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
