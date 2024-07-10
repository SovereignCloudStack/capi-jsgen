package main

import (
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/invopop/jsonschema"
	"log"
	"log/slog"
	"os"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
)

func init() {
	// no date, no time, no nothing
	log.SetFlags(0)
	// for dev purposes
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func main() {
	// Create reflector, DoNotReference inlines referenced schemas
	inLineRef := jsonschema.Reflector{DoNotReference: true}

	// Create Schema of v1beta1 Cluster Type
	fullClusterSchema := inLineRef.Reflect(&capi.Cluster{})

	// serialize...
	schemaData, err := json.MarshalIndent(fullClusterSchema, "", " ")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// load patch file
	patchData, err := os.ReadFile("defaultpatch.json")
	if err != nil {
		slog.Error("reading json patch", "err", err)
		os.Exit(1)
	}

	// parse patch
	patch, err := jsonpatch.DecodePatch(patchData)
	if err != nil {
		slog.Error("parsing json patch", "err", err)
		os.Exit(1)
	}

	schemaData, err = patch.Apply(schemaData)
	if err != nil {
		slog.Error("applying json patch", "err", err)
		os.Exit(1)
	}

	// and write to disk
	err = os.WriteFile("baseschema.json", schemaData, 0644)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}