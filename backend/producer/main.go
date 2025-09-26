package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/lmittmann/tint"
	"github.com/spf13/viper"
)

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			AddSource:  true,
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))
	viper.AutomaticEnv()

	for i := range 4 {
		time.Sleep(3 * time.Second)
		slog.Info("starting in ", "sec", 12-4*i)
	}

	slog.Info("envs", "BROKER", viper.GetString("BROKER"), "TOPIC_NAME", viper.GetString("TOPIC_NAME"))

	create_topic(viper.GetString("TOPIC_NAME"))

	producer(viper.GetString("TOPIC_NAME"))
}

func producer(topic string) {

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":   viper.GetString("BROKER"),
		"acks":                "all",
		"retries":             3,
		"delivery.timeout.ms": 10000,
	})
	if err != nil {
		slog.Error("producer create failed", "error", err)
		return
	}
	defer p.Close()

	// Create delivery report channel
	deliveryChan := make(chan kafka.Event)

	for i := range 10 {
		message := &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(fmt.Sprintf("%v %v", topic, i)),
		}

		// Produce with delivery channel
		err = p.Produce(message, deliveryChan)
		if err != nil {
			slog.Error("produce failed", "error", err)
			return
		}

		// slog.Info("message sent", "topic", topic, "index", i)

		// Wait for delivery report
		e := <-deliveryChan
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			slog.Error("delivery failed", "error", m.TopicPartition.Error)
			return
		} else {
			slog.Info("delivered",
				"topic", *m.TopicPartition.Topic,
				"partition", m.TopicPartition.Partition,
				"offset", m.TopicPartition.Offset)
		}

		time.Sleep(1 * time.Second)
	}

	slog.Info("flushing remaining messages...")
	p.Flush(15 * 1000)
}

func create_topic(topic string) {

	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": viper.GetString("BROKER")})
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		return
	}
	defer adminClient.Close()

	// Define topic specification
	topicSpec := kafka.TopicSpecification{
		Topic:             topic,
		NumPartitions:     10,
		ReplicationFactor: 1,
		// You can add more configurations here, e.g., "cleanup.policy": "compact"
		Config: map[string]string{
			"retention.ms": "604800000", // 7 days retention
		},
	}

	// Create a context with a timeout for the operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = adminClient.DeleteTopics(ctx, []string{topic}, kafka.SetAdminOperationTimeout(5*time.Second))
	if err != nil {
		fmt.Printf("Failed to delete topic %s: %s\n", topic, err)
		return
	}

	// Create the topic
	results, err := adminClient.CreateTopics(ctx, []kafka.TopicSpecification{topicSpec})
	if err != nil {
		fmt.Printf("Failed to create topic '%s': %s\n", topic, err)
		return
	}

	// Process the results
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			fmt.Printf("Topic creation failed for '%s': %s\n", result.Topic, result.Error)
		} else if result.Error.Code() == kafka.ErrTopicAlreadyExists {
			fmt.Printf("Topic '%s' already exists.\n", result.Topic)
		} else {
			fmt.Printf("Topic '%s' created successfully.\n", result.Topic)
		}
	}
}
