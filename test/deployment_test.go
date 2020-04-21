package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/YOVO-LABS/k8scaler/internal"
	_ "github.com/joho/godotenv/autoload"
)

func getKube() internal.KubernetesClient {
	kubernetes, err := internal.NewKubernetesClient(os.Getenv("KUBERNETES_SERVICE_HOST"),
		os.Getenv("KUBERNETES_SERVICE_PORT"),
		os.Getenv("KUBERNETES_NAMESPACE"), os.Getenv("KUBE_CONFIG"))

	if err != nil {
		fmt.Println("ERROR ", err)
	}

	return kubernetes

}

func TestGetDeployment(t *testing.T) {
	kubernetes := getKube()
	deployment, err := kubernetes.GetDeployment("cadence", "transcoder-activity")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	fmt.Printf("Current replica Number %v \n", deployment.Spec.GetReplicas())
}

func TestScaleDeployment(t *testing.T) {
	kubernetes := getKube()
	deployments, err := kubernetes.ScaleDeployment("stage", "theseus-grpc", 1)
	if err != nil {
		fmt.Println("ERROR ", err)
	}
	fmt.Println(&deployments)
}
