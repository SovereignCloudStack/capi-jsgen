package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
)

var (
	localMode    = flag.Bool("local", false, "run in local mode")
	requiredOnly = flag.Bool("required", false, "only include required variables into the schema")
	listen       = flag.String("listen", ":8080", "listen address")
	version      = "dev"
)

func init() {
	flag.Parse()
	// no date, no time, no nothing
	log.SetFlags(0)
	// for dev purposes
	slog.SetLogLoggerLevel(slog.LevelDebug) // TODO make configurable
}

func main() {
	//Hello
	slog.Info("Starting capi-jsgen", "version", version)
	// read baseSchema from File
	baseSchema, err := os.ReadFile("data/baseschema.json")
	if err != nil {
		slog.Error("reading baseschema file", "err", err)
		os.Exit(1)
	}

	configureSchemaBuilder(&SchemaBuilder{
		requiredOnly:           *requiredOnly,
		opSkeleton:             varOpSkeleton,
		setDefaultNamespace:    true,
		setDefaultClusterClass: true,
		setDefaultMachineClass: true,
		baseSchema:             baseSchema,
	})

	configureCache()

	http.Handle("GET /namespaces", cacheClient.Middleware(http.HandlerFunc(handleHTTPNamespaces)))
	http.Handle("GET /clusterschema/{namespace}/{clusterclass}", cacheClient.Middleware(http.HandlerFunc(handleHTTPClusterSchema)))
	slog.Info("Starting HTTP server on", "address", *listen)
	_ = http.ListenAndServe(*listen, nil)
}