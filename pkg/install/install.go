/*
Copyright 2019 the Velero contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package install

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/heptio/velero/pkg/discovery"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Install creates resources on the Kubernetes cluster.
// Need to get a client.DynamicFactory in, then produce a client per resource type.
func Install(client dynamic.Interface, helper discovery.Helper, resources *unstructured.UnstructuredList, logger *logrus.Logger) error {
	for _, r := range resources.Items {
		logger.WithField("resource", fmt.Sprintf("%s/%s", r.GetKind(), r.GetName())).Info("Creating resource")

		gvr := schema.ParseGroupResource(r.GetResourceVersion()).WithVersion("")
		_, err := client.Resource(gvr).Create(&r, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrapf(err, "Error creating resource %s/%s", r.GetKind(), r.GetName())
		}
	}
	return nil
}
