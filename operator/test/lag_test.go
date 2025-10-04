package test

import (
	"fmt"
	"testing"
	"time"

	"retail-scalar/utils"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/require"
)

func TestLag(t *testing.T) {
	ctx := t.Context()
	client, err := utils.NewClient(ctx, kafka.ConfigMap{
		"bootstrap.servers":  "localhost:9094",
		"group.id":           "retail",
		"enable.auto.commit": false,
	})
	require.NoError(t, err)

	for {
		lag, err := client.Lag(ctx)
		require.NoError(t, err)
		fmt.Println(lag)
		time.Sleep(time.Second)
	}
}
