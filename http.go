package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sigs.k8s.io/cluster-api/api/v1beta1"
)

var sb SchemaBuilder

const clusterClassesDemoFile = "data/clusterclasses.json"

func configureSchemaBuilder(refSb *SchemaBuilder) {
	sb = *refSb
}

func getNamespaces() ([]byte, error) {
	var ccListResource v1beta1.ClusterClassList
	var ccPerNs = make(map[string][]string)
	var (
		ccFileContent []byte
		err           error
	)
	if localMode {
		ccFileContent, err = os.ReadFile(clusterClassesDemoFile)
	} else {
		ccFileContent, err = getNamespacesFromK8s()
	}
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(ccFileContent, &ccListResource)
	if err != nil {
		return nil, err
	}

	for _, cc := range ccListResource.Items {
		ccPerNs[cc.Namespace] = append(ccPerNs[cc.Namespace], cc.Name)
	}

	// make nice JSON
	return json.MarshalIndent(ccPerNs, "", " ")
}

// getClusterSchema first gets Upstream ClusterClass Resource from the k8s apiserver using  getClusterClassFromK8s
// and afterwards builds a customized ClusterSchema based on the settings in the schemabuilder
func getClusterSchema(namespace, clusterclass string) ([]byte, error) {
	var ccResource v1beta1.ClusterClass
	var (
		ccFileContent []byte
		err           error
	)
	if localMode {
		ccFileContent, err = os.ReadFile("data/demo-clusterclass.json")
	} else {
		ccFileContent, err = getClusterClassFromK8s(namespace, clusterclass)
	}
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(ccFileContent, &ccResource)
	if err != nil {
		return nil, err
	}

	return sb.Build(ccResource)
}

// handleHTTPNamespaces is the http handler for responding to requests for the /namespaces path.
// It calls the next http-independent function getNamespaces to retrieve the namespaces and takes care of setting
// the correct return values and HTTP-Status codes
func handleHTTPNamespaces(w http.ResponseWriter, r *http.Request) {
	res, err := getNamespaces()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("getting namespaces", "err", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(res)
	if err != nil {
		slog.Error("writing response", "err", err)
	}
	return
}

// handleHTTPClusterSchema is the http handler responding to requests to the /clusterschema endpoint. It parses
// the arguments namespace and clusterclass and calls the next, http-independent function getClusterSchema. Afterwards
// it handles the return value and sets the correct values and HTTP-Status code
func handleHTTPClusterSchema(w http.ResponseWriter, r *http.Request) {
	res, err := getClusterSchema(r.PathValue("namespace"), r.PathValue("clusterclass"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("getting cluster schema", "err", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(res)
	if err != nil {
		slog.Error("writing response", "err", err)
	}
	return
}