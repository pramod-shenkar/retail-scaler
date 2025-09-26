package main

import (
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
