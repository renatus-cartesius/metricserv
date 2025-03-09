// Package agent providing types and methods for running metrics agent
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/renatus-cartesius/metricserv/pkg/encryption"
	"github.com/renatus-cartesius/metricserv/pkg/workerpool"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/go-resty/resty/v2"
	"github.com/renatus-cartesius/metricserv/pkg/logger"
	"github.com/renatus-cartesius/metricserv/pkg/metrics"
	"github.com/renatus-cartesius/metricserv/pkg/monitor"
	"github.com/renatus-cartesius/metricserv/pkg/server/models"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/shirou/gopsutil/v4/cpu"
)

const (
	updateURI  = "/update"
	updatesURI = "/updates/"
)

type Agent struct {
	monitor        monitor.Monitor
	reportInterval int
	pollInterval   int
	serverURL      string
	httpClient     *resty.Client
	hashKey        string
	reportCh       chan *monitor.RuntimeMetric
	workersPool    *workerpool.Pool[*monitor.RuntimeMetric]
	encProcessor   encryption.Processor
}

func NewAgent(repoInterval, pollInterval int, serverURL string, mon monitor.Monitor, hashKey string, encP encryption.Processor) (*Agent, error) {

	httpClient := resty.New()
	httpClient.
		SetRetryCount(3).
		AddRetryCondition(
			func(r *resty.Response, err error) bool {
				return r.StatusCode() == http.StatusTooManyRequests
			},
		)

	pool, err := workerpool.NewPool[*monitor.RuntimeMetric]()
	if err != nil {
		return nil, err
	}

	return &Agent{
		monitor:        mon,
		reportInterval: repoInterval,
		pollInterval:   pollInterval,
		serverURL:      serverURL,
		httpClient:     httpClient,
		hashKey:        hashKey,
		reportCh:       make(chan *monitor.RuntimeMetric, runtime.NumCPU()),
		workersPool:    pool,
		encProcessor:   encP,
	}, nil
}

func (a *Agent) Serve(ctx context.Context, reportWorkers int) {

	logger.Log.Info("starting agent")

	logger.Log.Debug(
		"creating workers",
		zap.Int("count", reportWorkers),
	)
	a.workersPool.Listen(ctx, reportWorkers, a.ReportHandler)

	defer func() {
		a.workersPool.Stop()
		a.workersPool.Wait()
	}()

	reportTicker := time.NewTicker(time.Duration(a.reportInterval) * time.Second)
	defer reportTicker.Stop()

	pollTicker := time.NewTicker(time.Duration(a.pollInterval) * time.Second)
	defer pollTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info(
				"shutting down agent",
			)
			return
		case <-pollTicker.C:
			a.Poll()
		case <-reportTicker.C:
			a.Report()
		}
	}
}

func (a *Agent) Poll() {

	metric := &models.Metric{
		ID:    "PollCount",
		MType: metrics.TypeCounter,
		Delta: new(int64),
	}

	*metric.Delta = 1

	resp, err := a.SendUpdate(metric)

	if err != nil {
		logger.Log.Error(
			"error on making report request",
			zap.String("metric", metric.ID),
			zap.Error(err),
		)
		return
	}

	logger.Log.Debug(
		"metric sended",
		zap.String("metric", metric.ID),
		zap.Int("status", resp.StatusCode()),
	)
}

func (a *Agent) ReportHandler(runtimeMetric *monitor.RuntimeMetric) {
	value, err := strconv.ParseFloat(runtimeMetric.Value, 64)
	if err != nil {
		logger.Log.Error(
			"error when parsing float value",
			zap.Error(err),
		)
		return
	}

	metric := &models.Metric{
		ID:    runtimeMetric.Name,
		MType: metrics.TypeGauge,
		Value: &value,
	}

	resp, err := a.SendUpdate(metric)

	if err != nil {
		logger.Log.Error(
			"error on making metric request",
			// zap.String("worker", uuid),
			zap.Error(err),
		)
		return
	}

	logger.Log.Debug(
		"metric sended",
		zap.Int("status", resp.StatusCode()),
		zap.String("metric", metric.ID),
		// zap.String("worker", uuid),
	)
}

func (a *Agent) Report() {
	if err := a.monitor.Flush(); err != nil {
		logger.Log.Error(
			"error when flusing monitor",
			zap.Error(err),
		)
	}

	stats := a.monitor.Get()

	for m, v := range stats {
		a.workersPool.AddJob(&monitor.RuntimeMetric{Name: m, Value: v})
	}

	v, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Error(
			"error on calling gopsutil",
			zap.Error(err),
		)
	}

	a.workersPool.AddJob(&monitor.RuntimeMetric{Name: "TotalMemory", Value: fmt.Sprintf("%v", float64(v.Total))})
	a.workersPool.AddJob(&monitor.RuntimeMetric{Name: "FreeMemory", Value: fmt.Sprintf("%v", float64(v.Free))})

	c, err := cpu.Percent(0, false)
	if err != nil {
		logger.Log.Error(
			"error on calling gopsutil",
			zap.Error(err),
		)
	}
	a.workersPool.AddJob(&monitor.RuntimeMetric{Name: "CPUutilization1", Value: fmt.Sprintf("%v", float64(c[0]))})

}

func (a *Agent) SendUpdate(metric *models.Metric) (*resty.Response, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)

	if err := json.NewEncoder(gzWriter).Encode(metric); err != nil {
		return nil, err
	}

	if err := gzWriter.Flush(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%s", a.serverURL, updateURI)
	payload, err := a.encProcessor.Encrypt(buf.Bytes())
	if err != nil {
		return nil, err
	}

	req := a.httpClient.R()

	if a.hashKey != "" {
		hash := hmac.New(sha256.New, []byte(a.hashKey))
		hash.Write(payload)

		req.SetHeader("HashSHA256", base64.StdEncoding.EncodeToString(hash.Sum(nil)))
	}

	req.SetHeader("Content-Encoding", "gzip").SetBody(payload)

	return req.Post(url)
}

func (a *Agent) SendUpdates(metric *models.Metric) (*resty.Response, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)

	if err := json.NewEncoder(gzWriter).Encode(metric); err != nil {
		return nil, err
	}

	if err := gzWriter.Flush(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%s", a.serverURL, updatesURI)
	payload, err := a.encProcessor.Encrypt(buf.Bytes())
	if err != nil {
		return nil, err
	}
	req := a.httpClient.R()

	if a.hashKey != "" {
		hash := hmac.New(sha256.New, []byte(a.hashKey))
		hash.Write(buf.Bytes())

		req.SetHeader("HashSHA256", base64.StdEncoding.EncodeToString(hash.Sum(nil)))
	}

	req.SetHeader("Content-Encoding", "gzip").SetBody(payload)

	return req.Post(url)
}
