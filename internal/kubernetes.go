package internal

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"log"

	"github.com/ericchiang/k8s"
	appsv1 "github.com/ericchiang/k8s/apis/apps/v1"
	apiv1 "github.com/ericchiang/k8s/apis/core/v1"
	"github.com/ghodss/yaml"
)

type Kubernetes struct {
	Client  *k8s.Client
	Context context.Context
}

type KubernetesClient interface {
	GetNode(string) (*apiv1.Node, error)
	ListDeployments(string) (*appsv1.DeploymentList, error)
	GetDeployment(string, string) (*appsv1.Deployment, error)
	ScaleDeployment(string, string, int32) (*appsv1.Deployment, error)
	//GetNodeList(string) (*apiv1.NodeList, error)
}

// NewKubernetesClient return a Kubernetes client
func NewKubernetesClient(host string, port string, namespace string, kubeConfigPath string) (kubernetes KubernetesClient, err error) {
	var client *k8s.Client

	if len(host) > 0 && len(port) > 0 {
		client, err = k8s.NewInClusterClient()

		if err != nil {
			err = fmt.Errorf("Error loading incluster client:\n%v", err)
			return
		}
	} else if len(kubeConfigPath) > 0 {
		client, err = loadK8sClient(kubeConfigPath)
		if err != nil {
			err = fmt.Errorf("Error loading client using kubeconfig:\n%v", err)
			return
		}
		client.SetHeaders = func(h http.Header) error {
			h.Set("Authorization", "Bearer "+os.Getenv("token"))
			return nil
		}
	} else {
		if namespace == "" {
			namespace = "default"
		}

		client = &k8s.Client{
			Endpoint:  "http://127.0.0.1:8001",
			Namespace: namespace,
			Client:    &http.Client{},
		}
	}

	kubernetes = &Kubernetes{
		Client:  client,
		Context: context.Background(),
	}

	return
}

// GetNode return the node object from given name
func (k *Kubernetes) GetNode(name string) (node *apiv1.Node, err error) {
	var nodes apiv1.NodeList
	if err := k.Client.List(k.Context, "", &nodes); err != nil {
		log.Fatal(err)
	}

	for _, node := range nodes.Items {
		fmt.Printf("name=%q schedulable=%t\n", *node.Metadata.Name, !*node.Spec.Unschedulable)
	}

	return
}

// ListDeployments return the deplpyments object from given name
func (k *Kubernetes) ListDeployments(name string) (*appsv1.DeploymentList, error) {
	var deployments appsv1.DeploymentList
	if err := k.Client.List(k.Context, name, &deployments); err != nil {
		log.Fatal(err)
	}

	for _, item := range deployments.Items {
		fmt.Printf("name=%q %q\n", *item.Metadata.Name, *&item.Metadata.Labels)
	}
	return &deployments, nil
}

// GetDeployment return the deplpyments object from given name
func (k *Kubernetes) GetDeployment(namespace, resourceName string) (*appsv1.Deployment, error) {
	var d appsv1.Deployment
	if err := k.Client.Get(k.Context, namespace, resourceName, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

//ScaleDeployment return the deplpyments object from given name
//https://github.com/ericchiang/k8s/issues/77#issuecomment-362213630
func (k *Kubernetes) ScaleDeployment(namespace string, resourceName string, replicaCount int32) (*appsv1.Deployment, error) {
	var d appsv1.Deployment
	if err := k.Client.Get(k.Context, namespace, resourceName, &d); err != nil {
		return nil, err
	}
	// fmt.Printf("Current replica Number %v \n", d.Spec.GetReplicas())
	d.Spec.Replicas = k8s.Int32(replicaCount)
	if err := k.Client.Update(k.Context, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// loadK8sClient parses a kubeconfig from a file and returns a Kubernetes
// client. It does not support extensions or client auth providers.
func loadK8sClient(kubeconfigPath string) (*k8s.Client, error) {
	data, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("Read kubeconfig error:\n%v", err)
	}

	// Unmarshal YAML into a Kubernetes config object.
	var config k8s.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("Unmarshal kubeconfig error:\n%v", err)
	}

	// fmt.Printf("%#v", config)
	return k8s.NewClient(&config)
}
