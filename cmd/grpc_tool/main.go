package main

import (
	"context"
	"fmt"
	"github.com/renatus-cartesius/metricserv/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/rand"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conn, err := grpc.DialContext(ctx, ":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	c := proto.NewMetricsServiceClient(conn)

	// Gauge add
	_, err = c.AddMetric(ctx, &proto.AddMetricRequest{
		MetricID: "test_gauge1",
		Metric: &proto.Metric{
			Type:  proto.MetricType_GAUGE,
			Value: fmt.Sprintf("%v", rand.Float32()),
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	respG, err := c.GetMetric(ctx, &proto.GetMetricRequest{
		MetricID: "test_gauge1",
		Type:     proto.MetricType_GAUGE,
	})

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(respG)

	// Counter add
	_, err = c.AddMetric(ctx, &proto.AddMetricRequest{
		MetricID: "test_counter1",
		Metric: &proto.Metric{
			Type:  proto.MetricType_COUNTER,
			Value: fmt.Sprintf("%v", rand.Int63()),
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	respC, err := c.GetMetric(ctx, &proto.GetMetricRequest{
		MetricID: "test_counter1",
		Type:     proto.MetricType_COUNTER,
	})

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(respC)

}
