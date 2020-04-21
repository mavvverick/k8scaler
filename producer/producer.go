package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/YOVO-LABS/k8scaler/internal"
	_ "github.com/joho/godotenv/autoload"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

func getKafkaWriter(kafkaURL, topic, groupID string) *kafka.Writer {
	brokers := strings.Split(kafkaURL, ",")
	mechanism, _ := scram.Mechanism(scram.SHA256, os.Getenv("username"), os.Getenv("pass"))
	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		SASLMechanism: mechanism,
		DualStack:     true,
		TLS: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Dialer:   dialer,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		Async:    true,
	})
}

func main() {
	// to produce messages
	kafkaURL := os.Getenv("kafkaURL")
	topic := os.Getenv("topic")
	groupID := os.Getenv("groupID")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := <-signals
		fmt.Println("Got signal: ", sig)
		cancel()
	}()

	writer := getKafkaWriter(kafkaURL, topic, groupID)

	scaler := internal.Message{
		Completed:  220,
		Failed:     46,
		Open:       799,
		Cancelled:  0,
		Timeout:    0,
		Terminated: 0,
		Total:      269,
	}

	data, _ := json.Marshal(scaler)

	err := writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte("video"),
			Value: data,
		},
	)

	if err != nil {
		fmt.Println(err)
	}

	writer.Close()
}
