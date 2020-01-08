/*
 * Copyright 2019-2020 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	"github.com/dgraph-io/dgraph-operator/pkg/defaults"
	"github.com/golang/glog"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// CustomResourceDefinitionGroupName is the CRD group name associated with all the dgraph
	// custom resources registered in k8s.
	CustomResourceDefinitionGroupName = "dgraph.io"

	// CustomResourceDefinitionSchemaVersionKey is key to label which holds the CRD schema version
	CustomResourceDefinitionSchemaVersionKey = "io.dgraph.k8s.crd.schema.version"

	// CustomResourceDefinitionSchemaVersion is semver-conformant version of CRD schema
	// Used to determine if CRD needs to be updated in cluster
	CustomResourceDefinitionSchemaVersion = "1.16"

	// CustomResourceDefinitionVersion is the current version of the resource
	CustomResourceDefinitionVersion = "v1alpha1"

	// DgraphClusterKindDefinition is Kind name of the custom resource definition.
	DgraphClusterKindDefinition = "DgraphCluster"
)

var (
	// SchemeBuilder is required by k8s deepcopy generator.
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder

	// AddToScheme adds all types of this clientset into the given scheme.
	// This allows composition of clientsets, like in:
	//
	//   import (
	//     "k8s.io/client-go/kubernetes"
	//     clientsetscheme "k8s.io/client-go/kuberentes/scheme"
	//     aggregatorclientsetscheme "k8s.io/kube-aggregator/pkg/client/
	//        clientset_generated/clientset/scheme"
	//   )
	//
	//   kclientset, _ := kubernetes.NewForConfig(c)
	//   aggregatorclientsetscheme.AddToScheme(clientsetscheme.Scheme)
	AddToScheme = localSchemeBuilder.AddToScheme
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{
	Group:   CustomResourceDefinitionGroupName,
	Version: CustomResourceDefinitionVersion,
}

func init() {
	// We only register manually written functions here. The registration of the
	// generated functions takes place in the generated files. The separation
	// makes the code compile even when the generated files are missing.
	localSchemeBuilder.Register(addKnownTypes)
}

// Resource takes an unqualified resource and returns back a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// addKnownTypes adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&DgraphCluster{},
	)

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

// CreateCustomResourceDefinitions creates our CRD objects in the kubernetes
// cluster using k8s api extension clientset.
func CreateCustomResourceDefinitions(clientset apiextclient.Interface) error {
	if err := createDgraphClusterCRD(clientset); err != nil {
		return err
	}

	return nil
}

// createDgraphClusterCRD creates a new Custom resource definition for kubernetes for type
// DgraphCluster.
func createDgraphClusterCRD(clientset apiextclient.Interface) error {
	var (
		// customResourceDefinitionSingularName is the singular name of custom resource definition
		customResourceDefinitionSingularName = "dgraphcluster"

		// customResourceDefinitionPluralName is the plural name of custom resource definition
		customResourceDefinitionPluralName = "dgraphclusters"

		// customResourceDefinitionShortNames are the abbreviated names to refer to this CRD's instances
		customResourceDefinitionShortNames = []string{"dc"}

		// crdName k8s represented name of the custom resource definition.
		crdName = customResourceDefinitionPluralName + "." + SchemeGroupVersion.Group
	)

	res := &apiextv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crdName,
			Labels: map[string]string{
				CustomResourceDefinitionSchemaVersionKey: CustomResourceDefinitionSchemaVersion,
			},
		},
		Spec: apiextv1.CustomResourceDefinitionSpec{
			Group: SchemeGroupVersion.Group,
			Versions: []apiextv1.CustomResourceDefinitionVersion{
				{
					Name:   SchemeGroupVersion.Version,
					Served: true,
					Subresources: &apiextv1.CustomResourceSubresources{
						Status: &apiextv1.CustomResourceSubresourceStatus{},
					},
					Storage: true,
					Schema:  dgraphClusterCRV,
				},
			},
			Names: apiextv1.CustomResourceDefinitionNames{
				Plural:     customResourceDefinitionPluralName,
				Singular:   customResourceDefinitionSingularName,
				ShortNames: customResourceDefinitionShortNames,
				Kind:       DgraphClusterKindDefinition,
			},

			// DgraphCluster resource is namespace scoped, user can specify the namespace
			// to create the cluster in.
			Scope: apiextv1.NamespaceScoped,
		},
	}

	return createUpdateCRD(clientset, "DgraphCluster/v1alpha1", res)
}

var (
	dgraphClusterCRV = &apiextv1.CustomResourceValidation{
		OpenAPIV3Schema: &apiextv1.JSONSchemaProps{
			Type:       "object",
			Properties: dgraphClusterProperties,
		},
	}

	dgraphClusterProperties = map[string]apiextv1.JSONSchemaProps{
		"ClusterID": clusterIDSchema,
	}

	maxClusterIDLen int64 = 64
	clusterIDSchema       = apiextv1.JSONSchemaProps{
		Description: "Unique ID of the dgraph cluster deployment.",
		Type:        "string",
		MaxLength:   &maxClusterIDLen,
	}
)

// createUpdateCRD ensures the CRD object is created in the k8s cluster. It
// will create or update the CRD.
func createUpdateCRD(clientset apiextclient.Interface, crdName string,
	crd *apiextv1.CustomResourceDefinition) error {
	_, err := clientset.ApiextensionsV1().
		CustomResourceDefinitions().
		Get(crd.ObjectMeta.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		glog.Infof("creating CRD (CustomResourceDefinition): %s", crdName)
		_, err = clientset.ApiextensionsV1().CustomResourceDefinitions().Create(crd)
		// This occurs when multiple operator instances might race to create the CRD.
		// This is a non error situation, as it has already been created.
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
	}
	if err != nil {
		return err
	}

	// Wait for the CRD to be available
	glog.Info("Waiting for CRD (CustomResourceDefinition) to be available...")
	err = wait.Poll(defaults.CRDWaitPollInterval, defaults.K8SAPIServerRequestTimeout,
		func() (bool, error) {
			crd, err := clientset.ApiextensionsV1().
				CustomResourceDefinitions().
				Get(crd.ObjectMeta.Name, metav1.GetOptions{})
			if err != nil {
				return false, err
			}
			for _, cond := range crd.Status.Conditions {
				switch cond.Type {
				case apiextv1.Established:
					if cond.Status == apiextv1.ConditionTrue {
						return true, err
					}
				case apiextv1.NamesAccepted:
					if cond.Status == apiextv1.ConditionFalse {
						glog.Errorf("name conflict for CRD: %s", crdName)
						return false, err
					}
				}
			}
			return false, err
		})

	// In case of an error, try to delete the CRD inorder to keep it clean.
	if err != nil {
		glog.Info("trying to cleanup CRD")
		deleteErr := clientset.ApiextensionsV1().
			CustomResourceDefinitions().
			Delete(crd.ObjectMeta.Name, nil)
		if deleteErr != nil {
			glog.Errorf("unable to delete k8s %s CRD %s. Deleting CRD due to: %s", crdName, deleteErr, err)
			return errors.NewAggregate([]error{err, deleteErr})
		}

		return err
	}

	glog.Infof("CRD (CustomResourceDefinition) %s is installed and up-to-date", crdName)
	return nil
}
