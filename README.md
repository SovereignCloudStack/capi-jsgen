# Introduction
The tool `capi-jsgen` creates json-schemas for cluster-API Clusters which reference a ClusterClass. The primary use-case for these schemas is to create UI-Forms out of it, but they can be used in other contexts as well.

# Concept
The generic Cluster specification is merged with each ClusterClass to create specialized versions of the Cluster specification. The tailoring takes place by copying the variable schemas from a ClusterClass to the generic Cluster spec. Additional to that fields that are not wanted in a UI (`.status`, `.metadata.gereration`, ...) are deleted and some defaults (`.spec.topology.class` for example) are set.
```
     generic cluster-spec
+-----------------------------+
|apiVersion: cluster.x-k8s.io |          
|kind: Cluster                |                 specialized cluster-spec
|spec:                        |----+        +-----------------------------+
|  topology:                  |    |        |apiVersion: cluster.x-k8s.io |
|    variables: <[]Object>    |    |        |kind: Cluster                |
+-----------------------------+    |        |spec:                        |
                                   |        |  topology:                  |
                                   |        |    class: example-cc        |
                                   +------->|    variables:               |
                                   |        |      - name var1            |
     specific clusterclass         |        |        schema: sc1          |
+-----------------------------+    |        |      - name var2            |
|apiVersion: cluster.x-k8s.io |    |        |        schema: sc2          |
|kind: ClusterClass           |    |        +-----------------------------+
|name: example-cc             |    |        
|spec:                        |    |     
|  variables:                 |----+     
|    - name: var1             |          
|      schema: sc1            |          
|    - name: var2             |          
|      schema: sc2            |          
+-----------------------------+
```
Notice that the above graphic recklessly mixes specification (Cluster) and instances (ClusterClass). It is not correct in every detail and only meant to get a quick idea of the concept.
# API
There are two API endpoints that can be consumed. The first one offers information about the namespaces and the offered cluster-classes in the cluster. The second API offers Cluster schemas that are specific for a clusterclass in a namespace.
## `GET /clusterclasses`
## `GET /clusterschema/{namespace}/{clusterclass}`

# Background
[Kubernetes](https://kubernetes.io) stores structured objects (_resources_) in its database. Before saving a new _resource_ to its database, the _kubernetes-apiserver_ validates whether the _resource_ adheres to a certain structure. You can describe the desired structure (_schema_) for a kubernetes _resource_ using `kubectl explain`. If you do so, `kubectl` retrieves the corresponding _schema_ from the _kubernetes-apiserver_ under `<server>/openapi/v3`. The format of this _schema_ is  [json-schema](https://json-schema.org), a standard to describe the validity of a structured object. This _schema_ can be used to generate documentation, display a help-text (`kubectl explain`), generate a GUI or validate whether an object adheres to the required structure in the _schema_.

If a _resource_ is validated by the _schema_ this is only a necessary condition for it to be accepted by the _kubernetes-apiserver_. The schema validation alone is not sufficient to guarantee that the _resource_ is accepted by the _kubernetes-apiserver_. For example, there can be webhooks that impose additional, stricter requirements upon the structure of the _resource_.

[cluster-API](https://github.com/kubernetes-sigs/cluster-api) is an extension to the _kubernetes-apiserver_. It introduces several new _resource definitions_ ([CRDs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)). The purpose of those _CRDs_ is to define the _schema_ of _resources_ that are used to define kubernetes clusters. An important, central _schema_ is called [Cluster](https://doc.crds.dev/github.com/kubernetes-sigs/cluster-api/cluster.x-k8s.io/Cluster/v1beta1@v1.7.3). It can be used in two ways (simplified):
1. manually set all required values
2. use a templating mechanism called [ClusterClass](https://cluster-api.sigs.k8s.io/tasks/experimental-features/cluster-class/) and only feed variables to customize the output of the template

Both approaches use the same _schema_. The part in the _schema_ (_subschema_) that defines the format of the variables is loosely defined. Basically, it accepts arbitrary key-value pairs. This makes sense because this _schema_ has to validate all thinkable instances of [Cluster](https://doc.crds.dev/github.com/kubernetes-sigs/cluster-api/cluster.x-k8s.io/Cluster/v1beta1@v1.7.3), no matter which variables are used by a ClusterClass or if a ClusterClass is used at all.

The loose definition of the variables makes the generic Cluster _schema_ a bad choice for creating a form or validating if a Cluster adheres to the required variables in a specific ClusterClass. The validation process for a Cluster that uses a ClusterClass is as follows:
1. Read the referenced ClusterClass from `.spec.topology.class` from the Cluster _resource_.
2. Use the referenced ClusterClass obtained in the prior step to read the variables and their _schemas_ from `.status.variables` 
3. Validate all variables set in the Cluster resource under `.spec.topology.variables` against the variable _schemas_ obtained in the prior step
 
<quote>`capi-jsgen` creates a concrete _schema_ out of the generic Cluster _schema_ and the variables of a concrete ClusterClass by embedding the variable _schemas_ of the ClusterClass into the Cluster _schema_.</quote>

# Limitations
* only inline Patches work (no Runtime-SDK support), but configurable behaviour for `definitionsConflict: true` is planned
* CEL is not supported, probably as long as it is not part of json-schema
# Glossary
##### resource
A kubernetes object. Its structure can be validated by its _schema_.
##### schema
A description of the structure of an object.
##### subschema
A description of a part of the structure of an object. Part of a _schema_.