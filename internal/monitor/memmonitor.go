package monitor

import (
	"fmt"
	"runtime"
)

type MemMonitor struct {
	memStats runtime.MemStats
}

func (m *MemMonitor) Flush() error {
	runtime.ReadMemStats(&m.memStats)
	return nil
}

func (m *MemMonitor) Get() map[string]string {
	stats := map[string]string{
		"Alloc":         fmt.Sprintf("%f", float64(m.memStats.Alloc)),
		"BuckHashSys":   fmt.Sprintf("%f", float64(m.memStats.BuckHashSys)),
		"Frees":         fmt.Sprintf("%f", float64(m.memStats.Frees)),
		"GCCPUFraction": fmt.Sprintf("%f", float64(m.memStats.GCCPUFraction)),
		"GCSys":         fmt.Sprintf("%f", float64(m.memStats.GCSys)),
		"HeapAlloc":     fmt.Sprintf("%f", float64(m.memStats.HeapAlloc)),
		"HeapIdle":      fmt.Sprintf("%f", float64(m.memStats.HeapIdle)),
		"HeapInuse":     fmt.Sprintf("%f", float64(m.memStats.HeapInuse)),
		"HeapObjects":   fmt.Sprintf("%f", float64(m.memStats.HeapObjects)),
		"HeapReleased":  fmt.Sprintf("%f", float64(m.memStats.HeapReleased)),
		"HeapSys":       fmt.Sprintf("%f", float64(m.memStats.HeapSys)),
		"LastGC":        fmt.Sprintf("%f", float64(m.memStats.LastGC)),
		"Lookups":       fmt.Sprintf("%f", float64(m.memStats.Lookups)),
		"MCacheInuse":   fmt.Sprintf("%f", float64(m.memStats.MCacheInuse)),
		"MCacheSys":     fmt.Sprintf("%f", float64(m.memStats.MCacheSys)),
		"MSpanInuse":    fmt.Sprintf("%f", float64(m.memStats.MSpanInuse)),
		"MSpanSys":      fmt.Sprintf("%f", float64(m.memStats.MSpanSys)),
		"Mallocs":       fmt.Sprintf("%f", float64(m.memStats.Mallocs)),
		"NextGC":        fmt.Sprintf("%f", float64(m.memStats.NextGC)),
		"NumForcedGC":   fmt.Sprintf("%f", float64(m.memStats.NumForcedGC)),
		"NumGC":         fmt.Sprintf("%f", float64(m.memStats.NumGC)),
		"OtherSys":      fmt.Sprintf("%f", float64(m.memStats.OtherSys)),
		"PauseTotalNs":  fmt.Sprintf("%f", float64(m.memStats.PauseTotalNs)),
		"StackInuse":    fmt.Sprintf("%f", float64(m.memStats.StackInuse)),
		"StackSys":      fmt.Sprintf("%f", float64(m.memStats.StackSys)),
		"Sys":           fmt.Sprintf("%f", float64(m.memStats.Sys)),
		"TotalAlloc":    fmt.Sprintf("%f", float64(m.memStats.TotalAlloc)),
	}

	return stats
}
