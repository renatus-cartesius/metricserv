package storage

import (
	"context"
	"github.com/google/uuid"
	"github.com/renatus-cartesius/metricserv/pkg/metrics"
	"math/rand"
	"testing"
)

func BenchmarkMemStorage_Add(b *testing.B) {
	storage, err := NewMemStorage("/dev/null")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	b.Run("gauge", func(b *testing.B) {
		value := rand.Float64()
		metricName := uuid.NewString()
		metric := metrics.NewGauge(metricName, value)

		//b.StartTimer()
		for i := 0; i < b.N; i++ {
			if err := storage.Add(context.Background(), metricName, metric); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("counter", func(b *testing.B) {
		value := rand.Int63()
		metricName := uuid.NewString()
		metric := metrics.NewCounter(metricName, value)

		//b.StartTimer()
		for i := 0; i < b.N; i++ {
			if err := storage.Add(context.Background(), metricName, metric); err != nil {
				b.Fatal(err)
			}
		}
	})

}

func BenchmarkMemStorage_CheckMetric(b *testing.B) {
	ctx := context.Background()
	storageSize := 10000
	storage, err := NewMemStorage("/dev/null")
	if err != nil {
		b.Fatal(err)
	}

	metricIds := make([]string, storageSize/2)

	for i := 0; i < storageSize; i++ {
		var metric metrics.Metric
		id := uuid.NewString()
		switch rand.Intn(2) {
		case 0:
			value := rand.Float64()
			metric = metrics.NewGauge(id, value)
			if err := storage.Add(ctx, id, metric); err != nil {
				b.Error(err)
			}
		case 1:
			value := rand.Int63()
			metric = metrics.NewCounter(id, value)
			if err := storage.Add(ctx, id, metric); err != nil {
				b.Error(err)
			}
		default:
		}
		if i%2 == 0 {
			metricIds = append(metricIds, id)
		} else {
			metricIds = append(metricIds, uuid.NewString())
		}
	}

	b.ResetTimer()

	for _, id := range metricIds {
		if _, err := storage.CheckMetric(ctx, id); err != nil {
			b.Error(err)
		}
	}
}
