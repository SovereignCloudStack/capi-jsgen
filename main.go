package main

import (
	"flag"
	cache "github.com/victorspringer/http-cache"
	"log"
	"log/slog"
	"net/http"
	"os"
)

var (
	localMode             = flag.Bool("local", false, "run in local mode")
	requiredOnly          = flag.Bool("required", false, "only include required variables into the schema")
	listen                = flag.String("listen", ":8080", "listen address")
	baseSchemaFile        = flag.String("baseschema", "data/baseschema.json", "path to baseschema file")
	sbDefaultNamespace    = flag.Bool("default-namespace", true, "When set, the namespace of the clusterclass is used as default for .metadata.namespace")
	sbDefaultClusterClass = flag.Bool("default-clusterclass", true, "When set, the clusterclass is used as default for .spec.topology.class")
	sbDefaultMachineClass = flag.Bool("default-machineclass", true, "When set, the clusterclass name is used as default for .spec.topology.workers.machineDeployments.class")
	cacheTime             = flag.String("cache", "1h", "Duration how long API-responses are cached")

	version = "dev"

	sb          SchemaBuilder
	cacheClient *cache.Client
)

func init() {
	flag.Parse()
	// no date, no time, no nothing
	log.SetFlags(0)
}

func main() {
	//Hello
	slog.Info("Starting capi-jsgen", "version", version)
	// read baseSchema from File
	baseSchema, err := os.ReadFile(*baseSchemaFile)
	if err != nil {
		slog.Error("reading baseschema file", "err", err)
		os.Exit(1)
	}
	// and create schemabuilder of it (and the CLI args)
	sb = SchemaBuilder{
		requiredOnly:           *requiredOnly,
		opSkeleton:             varOpSkeleton,
		setDefaultNamespace:    *sbDefaultNamespace,
		setDefaultClusterClass: *sbDefaultClusterClass,
		setDefaultMachineClass: *sbDefaultMachineClass,
		baseSchema:             baseSchema,
	}

	// setup HTTP response cache
	configureCache(*cacheTime)

	http.Handle("GET /namespaces", cacheClient.Middleware(http.HandlerFunc(handleHTTPNamespaces)))
	http.Handle("GET /clusterschema/{namespace}/{clusterclass}", cacheClient.Middleware(http.HandlerFunc(handleHTTPClusterSchema)))
	slog.Info("Starting HTTP server on", "address", *listen)
	_ = http.ListenAndServe(*listen, nil)
}