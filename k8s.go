package main

import (
	"io"
	"k8s.io/client-go/rest"
	"log/slog"
)

// getFromK8s tries to access an autodetect the k8s apiserver if this program runs inside a cluster. If it as able to
// connect to an API Server it will issue a get request to the specified URL and will return the response as the
// []byte
func getFromK8s(path string) ([]byte, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := rest.HTTPClientFor(config)
	if err != nil {
		return nil, err
	}
	slog.Info("Issuing http-request GET", "host", config.Host, "path", path)
	resp, err := client.Get(config.Host + path)
	if err != nil {
		return nil, err
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
	return getFromK8s("/apis/cluster.x-k8s.io/v1beta1/namespaces/" + namespace + "/clusterclasses/" + name)
}