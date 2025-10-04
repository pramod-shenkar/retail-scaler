package utils

import (
	"context"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Client struct {
	kafka.ConfigMap
	*kafka.Consumer
	kafka.TopicPartitions
}

func NewClient(ctx context.Context, config kafka.ConfigMap) (*Client, error) {

	admin, err := kafka.NewAdminClient(&config)
	if err != nil {
		slog.ErrorContext(ctx, "error while getting client", "err", err)
		return nil, err
	}

	defer admin.Close()

	meta, err := admin.GetMetadata(nil, true, 50000)
	if err != nil {
		slog.ErrorContext(ctx, "error while getting meta", "err", err)
		return nil, err
	}

	consumer, err := kafka.NewConsumer(&config)
	if err != nil {
		slog.ErrorContext(ctx, "error while getting consumer", "err", err)
		return nil, err
	}

	var all kafka.TopicPartitions
	for _, topic := range meta.Topics {
		for _, part := range topic.Partitions {
			all = append(all, kafka.TopicPartition{
				Topic:     &topic.Topic,
				Partition: part.ID,
			})
		}
	}

	return &Client{config, consumer, all}, nil

}

func (c *Client) Close() {
	defer c.Consumer.Close()
}

func (c *Client) Lag(ctx context.Context) (map[string]int64, error) {

	commited, err := c.Consumer.Committed(c.TopicPartitions, 50000)
	if err != nil {
		slog.ErrorContext(ctx, "error while getting Committed", "err", err)
		return nil, err
	}

	lag := map[string]int64{}

	for _, tp := range commited {
		_, high, err := c.Consumer.QueryWatermarkOffsets(*tp.Topic, tp.Partition, 5000)
		if err != nil {
			slog.ErrorContext(ctx, "error while getting QueryWatermarkOffsets", "err", err)
			return nil, err
		}

		var offset int64

		switch tp.Offset {
		case kafka.OffsetInvalid, kafka.OffsetStored, kafka.OffsetEnd:
			offset = high
		case kafka.OffsetBeginning:
			offset = 0
		}

		lag[*tp.Topic] += high - offset
	}

	return lag, nil
}
