/*
Copyright 2017, 2019 the Velero contributors.

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

package client

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	v1 "github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/cmd/util/flag"
	"github.com/heptio/velero/pkg/features"
	clientset "github.com/heptio/velero/pkg/generated/clientset/versioned"
)

// Factory knows how to create a VeleroClient and Kubernetes client.
type Factory interface {
	// BindFlags binds common flags (--kubeconfig, --namespace) to the passed-in FlagSet.
	BindFlags(flags *pflag.FlagSet)
	// Client returns a VeleroClient. It uses the following priority to specify the cluster
	// configuration: --kubeconfig flag, KUBECONFIG environment variable, in-cluster configuration.
	Client() (clientset.Interface, error)
	// KubeClient returns a Kubernetes client. It uses the following priority to specify the cluster
	// configuration: --kubeconfig flag, KUBECONFIG environment variable, in-cluster configuration.
	KubeClient() (kubernetes.Interface, error)
	// DynamicClient returns a Kubernetes dynamic client. It uses the following priority to specify the cluster
	// configuration: --kubeconfig flag, KUBECONFIG environment variable, in-cluster configuration.
	DynamicClient() (dynamic.Interface, error)
	// SetBasename changes the basename for an already-constructed client.
	// This is useful for generating clients that require a different user-agent string below the root `velero`
	// command, such as the server subcommand.
	SetBasename(string)
	// SetClientQPS sets the Queries Per Second for a client.
	SetClientQPS(float32)
	// SetClientBurst sets the Burst for a client.
	SetClientBurst(int)
	// ClientConfig returns a rest.Config struct used for client-go clients.
	ClientConfig() (*rest.Config, error)
	// Namespace returns the namespace which the Factory will create clients for.
	Namespace() string
}

type factory struct {
	flags       *pflag.FlagSet
	features    *features.FeatureFlagSet
	kubeconfig  string
	kubecontext string
	baseName    string
	namespace   string
	clientQPS   float32
	clientBurst int
}

// NewFactory returns a Factory.
func NewFactory(baseName string) Factory {
	f := &factory{
		flags:    pflag.NewFlagSet("", pflag.ContinueOnError),
		baseName: baseName,
	}

	f.namespace = os.Getenv("VELERO_NAMESPACE")
	var config VeleroConfig
	if config, err := LoadConfig(); err == nil {
		if config.Namespace() != "" {
			f.namespace = config.Namespace()
		}

	} else {
		fmt.Fprintf(os.Stderr, "WARNING: error retrieving namespace from config file: %v\n", err)
	}

	// We didn't get the namespace via env var or config file, so use the default.
	// Command line flags will override when BindFlags is called.
	if f.namespace == "" {
		f.namespace = v1.DefaultNamespace
	}

	f.flags.StringVar(&f.kubeconfig, "kubeconfig", "", "Path to the kubeconfig file to use to talk to the Kubernetes apiserver. If unset, try the environment variable KUBECONFIG, as well as in-cluster configuration")
	f.flags.StringVarP(&f.namespace, "namespace", "n", f.namespace, "The namespace in which Velero should operate")
	f.flags.StringVar(&f.kubecontext, "kubecontext", "", "The context to use to talk to the Kubernetes apiserver. If unset defaults to whatever your current-context is (kubectl config current-context)")
	// Use a separate StringArray to collect the features because we want to combine the ones in the config file with the ones from the command line, not override them.
	var cmdFeatures flag.StringArray
	f.flags.Var(&cmdFeatures, "features", "Comma-separated list of features to enable for this Velero process. Combines with values from $HOME/.config/velero/config.json if present")

	allFeatures := append(config.Features(), cmdFeatures...)

	f.features = features.NewFeatureFlagSet(allFeatures...)

	return f
}

func (f *factory) BindFlags(flags *pflag.FlagSet) {
	flags.AddFlagSet(f.flags)
}

func (f *factory) ClientConfig() (*rest.Config, error) {
	return Config(f.kubeconfig, f.kubecontext, f.baseName, f.clientQPS, f.clientBurst)
}

func (f *factory) Client() (clientset.Interface, error) {
	clientConfig, err := f.ClientConfig()
	if err != nil {
		return nil, err
	}

	veleroClient, err := clientset.NewForConfig(clientConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return veleroClient, nil
}

func (f *factory) KubeClient() (kubernetes.Interface, error) {
	clientConfig, err := f.ClientConfig()
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return kubeClient, nil
}

func (f *factory) DynamicClient() (dynamic.Interface, error) {
	clientConfig, err := f.ClientConfig()
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(clientConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return dynamicClient, nil
}

func (f *factory) SetBasename(name string) {
	f.baseName = name
}

func (f *factory) SetClientQPS(qps float32) {
	f.clientQPS = qps
}

func (f *factory) SetClientBurst(burst int) {
	f.clientBurst = burst
}

func (f *factory) Namespace() string {
	return f.namespace
}
