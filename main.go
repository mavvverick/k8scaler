package main

import (
	"os"

	"github.com/YOVO-LABS/k8scaler/internal"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
)

func main() {
	kubernetes, err := internal.NewKubernetesClient(os.Getenv("KUBERNETES_SERVICE_HOST"),
		os.Getenv("KUBERNETES_SERVICE_PORT"),
		os.Getenv("KUBERNETES_NAMESPACE"), os.Getenv("KUBE_CONFIG"))

	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing Kubernetes client")
	}

	internal.ListenKafka(kubernetes)
}
