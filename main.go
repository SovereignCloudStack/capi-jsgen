package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
)

var (
	localMode bool
)

func init() {
	flag.BoolVar(&localMode, "local", false, "run in local mode")
	flag.Parse()
	// no date, no time, no nothing
	log.SetFlags(0)
	// for dev purposes
	slog.SetLogLoggerLevel(slog.LevelDebug) // TODO make configurable
}

func main() {
	// read baseSchema from File
	baseSchema, err := os.ReadFile("data/baseschema.json")
	if err != nil {
		slog.Error("reading baseschema file", "err", err)
		os.Exit(1)
	}

	configureSchemaBuilder(&SchemaBuilder{
		requiredOnly:           true,
		opSkeleton:             varOpSkeleton,
		setDefaultNamespace:    true,
		setDefaultClusterClass: true,
		setDefaultMachineClass: true,
		baseSchema:             baseSchema,
	})

	http.HandleFunc("GET /namespaces", handleHTTPNamespaces)
	http.HandleFunc("GET /clusterschema/{namespace}/{clusterclass}", handleHTTPClusterSchema)
	_ = http.ListenAndServe(":8080", nil)
}