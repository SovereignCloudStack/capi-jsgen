package main

import (
	"io"
	"k8s.io/client-go/rest"
	"log/slog"
	"net/http"
	"os"
)

// getFromK8s tries to access an autodetect the k8s apiserver if this program runs inside a cluster. If it as able to
// connect to an API Server it will issue a get request to the specified URL and will return the response as the
// []byte
func getFromK8s(path string) ([]byte, error) {
	// creates the in-cluster config
	config := rest.Config{
		Host: "https://moin.k8s.scs.community",
	}
	/*
		config, err := rest.InClusterConfig()
		if err != nil {
			slog.Error("getting kubernetes config failed", "err", err)
			os.Exit(1)
		}
	*/
	slog.Info("HTTP request", "host", config.Host, "path", path)
	resp, err := http.Get(config.Host + path)
	if err != nil {
		slog.Error("http request failed", "err", err)
		os.Exit(1)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error("close failed", "err", err)
		}
	}(resp.Body)
	return io.ReadAll(resp.Body)

}

func getNamespacesFromK8s() ([]byte, error) {
	return getFromK8s("/apis/cluster.x-k8s.io/v1beta1/clusterclasses")
}

func getClusterClassFromK8s(namespace, name string) ([]byte, error) {
	return getFromK8s("/apis/cluster.x-k8s.io/v1beta1/namespaces/" + namespace + "/clustersclasses/" + name)
}