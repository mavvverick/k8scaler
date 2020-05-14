package internal

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

func ListenKafka(kube KubernetesClient) {
	kafkaURL := os.Getenv("KAFKA_BROKERS")
	topic := os.Getenv("KAFKA_TOPIC")
	groupID := os.Getenv("KAFKA_GROUPID")
	reader := getKafkaReader(kafkaURL, topic, groupID)
	defer reader.Close()
	fmt.Println("start consuming ... !!")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sig := <-signals
		fmt.Println("Got signal: ", sig)
		cancel()
	}()

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Fatalln(err)
		}
		err = Process(kube, m.Value)
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Printf("message at topic:%v partition:%v offset:%v	%s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
	}
}

func getKafkaReader(kafkaURL, topic, groupID string) *kafka.Reader {
	brokers := strings.Split(kafkaURL, ",")
	dialer := &kafka.Dialer{
		Timeout: 10 * time.Second,
	}
	if os.Getenv("KAFKA_USERNAME") != "" {
		mechanism, _ := scram.Mechanism(scram.SHA256, os.Getenv("KAFKA_USERNAME"), os.Getenv("KAFKA_PASS"))
		dialer.SASLMechanism = mechanism
		dialer.DualStack = true
		dialer.TLS = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	readerConfig := kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		Dialer:   dialer,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	}
	return kafka.NewReader(readerConfig)
}
