package main

import (
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
