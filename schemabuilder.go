package main

import (
	"encoding/json"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"log/slog"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"strings"
)

type SchemaBuilder struct {
	baseSchema             []byte // the base Schema to work on, all patches are applied on top of that base Schema
	requiredOnly           bool   // if set only variables with required: true are included
	opSkeleton             string // set Skeleton that is used to template the patchOps
	setDefaultNamespace    bool   // sets the namespace default to the namespace of the ClusterClass
	setDefaultClusterClass bool   // sets the clusterclass default to the name of the ClusterClass
	setDefaultMachineClass bool   // sets workers.machineDeployments[0].class to the name of the ClusterClass
}

const (
	varOpSkeleton = `
{ 
	"op": "add",
    "path": "/properties/spec/properties/topology/properties/variables/items/-",
	"value": {
		"type": "object",
			"properties": {
				"name": {
					"type": "string",
					"default": "%[1]s",
					"const": "%[1]s"
				},
			"value": %[2]s
		}
	}
}
`
	spaceOpSkeleton = `
    {
        "op": "add",
        "path": "/properties/metadata/properties/namespace/default",
        "value": "%s"
    }
`
	classOpSkeleton = `
    {
        "op": "add",
        "path": "/properties/spec/properties/topology/properties/class/default",
        "value": "%s"
    }
`
	machineOpSkeleton = `
    {
        "op": "add",
        "path": "/properties/spec/properties/topology/properties/workers/properties/machineDeployments/default",
        "value": [
            {
                "class": "%[1]s",
                "failureDomain": "nova",
                "name": "md0",
                "replicas": 1
            }
            ]
    }
`
)

// Build takes a ClusterClass and returns a customized version of the baseSchema
// that validates Clusters of ClusterClass cc. It looks up the variables under
// `spec.variables` and transforms them to patches, depending on the settings of the SchemaBuilder
// To take the variables from `spec.variables` is officially not the correct way to retrieve variables
// of a ClusterClass (the correct way is to parse `.spec.Status.Variables` there you can find inline
// and RuntimeSDK variables), but since we are only interested in inline patches this is
// the easiest way to go. Otherwise that would involve many more checks and config flags
// Apart from the variables, the defaults for clusterclass, namespace, machinedeployment
// are set, depending on the settings
func (sb SchemaBuilder) Build(cc capi.ClusterClass) ([]byte, error) {
	var patchOps []string // collect patches over the function

	// start with the namespace
	if sb.setDefaultNamespace {
		patchOps = append(patchOps, fmt.Sprintf(spaceOpSkeleton, cc.ObjectMeta.Namespace))
	}

	// next is the ClusterClass
	if sb.setDefaultClusterClass {
		patchOps = append(patchOps, fmt.Sprintf(classOpSkeleton, cc.ObjectMeta.Name))
	}

	// next is the workers machineDeployments[0] group
	if sb.setDefaultMachineClass {
		patchOps = append(patchOps, fmt.Sprintf(machineOpSkeleton, cc.ObjectMeta.Name))
	}

	// add the variables
	for _, v := range cc.Spec.Variables {
		if sb.requiredOnly {
			if v.Required {
				schema, err := json.Marshal(v.Schema.OpenAPIV3Schema)
				if err != nil {
					slog.Warn(err.Error())
					// give other variables the chance to succeed
					break
				}
				// still in game? let's build the patchOp
				patchOps = append(patchOps, fmt.Sprintf(sb.opSkeleton, v.Name, string(schema)))
			}
		} else {
			schema, err := json.Marshal(v.Schema.OpenAPIV3Schema)
			if err != nil {
				slog.Warn(err.Error())
				// give other variables the chance to succeed
				break
			}
			// still in game? let's build the patchOp
			patchOps = append(patchOps, fmt.Sprintf(sb.opSkeleton, v.Name, string(schema)))
		}
	}

	// compile the patch
	patch, err := jsonpatch.DecodePatch([]byte("[" + strings.Join(patchOps, ",") + "]"))
	if err != nil {
		return nil, err
	}

	// apply the patch and return the new Schema
	return patch.ApplyIndent(sb.baseSchema, " ")
}