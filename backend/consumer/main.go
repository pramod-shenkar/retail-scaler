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

	create_topic(viper.GetString("TOPIC_NAME"))

	consumer(viper.GetString("TOPIC_NAME"))
}

func consumer(topic string) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": viper.GetString("BROKER"),
		"group.id":          viper.GetString("GROUP_ID"),
		"auto.offset.reset": "earliest",
		// Consumer group coordinator settings
		"session.timeout.ms":            10000,
		"heartbeat.interval.ms":         3000,
		"max.poll.interval.ms":          300000,
		"coordinator.query.interval.ms": 1000,
		// Reduce debug output - remove this line if you want all the logs
		// "debug":             "generic,broker,topic,metadata,feature,queue,msg,protocol,cgrp,security,fetch,interceptor,plugin,consumer,admin,eos,mock,assignor,conf",
	})
	if err != nil {
		slog.Error("consumer create failed", "error", err)
		return
	}
	defer c.Close()

	err = c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		slog.Error("subscribe failed", "error", err)
		return
	}

	slog.Info("consumer started", "topic", topic, "group_id", viper.GetString("GROUP_ID"))

	for {
		msg, err := c.ReadMessage(100 * time.Millisecond) // Add timeout instead of -1
		if err != nil {
			// Check if it's a timeout error (which is normal)
			if err.(kafka.Error).Code() == kafka.ErrTimedOut {
				// slog.Debug("no message received, continuing...")
				continue
			}
			// Log other errors but continue
			slog.Warn("error reading message", "error", err)
			return
		}

		slog.Info("consumed",
			"topic", *msg.TopicPartition.Topic,
			"partition", msg.TopicPartition.Partition,
			"offset", msg.TopicPartition.Offset,
			"value", string(msg.Value))
	}
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
