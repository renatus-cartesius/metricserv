package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/go-resty/resty/v2"
	"github.com/renatus-cartesius/metricserv/internal/logger"
	"github.com/renatus-cartesius/metricserv/internal/metrics"
	"github.com/renatus-cartesius/metricserv/internal/monitor"
	"github.com/renatus-cartesius/metricserv/internal/server/models"
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
	exitCh         chan os.Signal
}

func NewAgent(repoInterval, pollInterval int, serverURL string, monitor monitor.Monitor, exitCh chan os.Signal) *Agent {

	httpClient := resty.New()
	httpClient.
		SetRetryCount(3).
		AddRetryCondition(
			func(r *resty.Response, err error) bool {
				return r.StatusCode() == http.StatusTooManyRequests
			},
		)

	return &Agent{
		monitor:        monitor,
		reportInterval: repoInterval,
		pollInterval:   pollInterval,
		serverURL:      serverURL,
		httpClient:     httpClient,
		exitCh:         exitCh,
	}
}

func (a *Agent) Serve() {

	logger.Log.Info("starting agent")

	reportTicker := time.NewTicker(time.Duration(a.reportInterval) * time.Second)
	defer reportTicker.Stop()

	pollTicker := time.NewTicker(time.Duration(a.pollInterval) * time.Second)
	defer pollTicker.Stop()

	for {
		select {
		case <-a.exitCh:
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

	var metricJSON bytes.Buffer
	gzWriter := gzip.NewWriter(&metricJSON)

	if err := json.NewEncoder(gzWriter).Encode(metric); err != nil {
		logger.Log.Error(
			"error on marshaling metric",
			zap.String("metricID", metric.ID),
			zap.Error(err),
		)
		return
	}
	gzWriter.Flush()

	url := fmt.Sprintf("%s%s", a.serverURL, updateURI)
	resp, err := a.httpClient.R().
		SetHeader("Content-Encoding", "gzip").
		SetBody(metricJSON.Bytes()).
		Post(url)

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

func (a *Agent) Report() {
	if err := a.monitor.Flush(); err != nil {
		logger.Log.Error(
			"error when flusing monitor",
			zap.Error(err),
		)
	}

	var metricsBatch models.MetricsBatch

	stats := a.monitor.Get()
	for m, v := range stats {

		value, err := strconv.ParseFloat(v, 64)
		if err != nil {
			logger.Log.Error(
				"error parsing float",
				zap.Error(err),
			)
		}

		metric := &models.Metric{
			ID:    m,
			MType: metrics.TypeGauge,
			Value: &value,
		}

		metricsBatch = append(metricsBatch, metric)

	}

	var metricsBatchJSON bytes.Buffer
	gzWriter := gzip.NewWriter(&metricsBatchJSON)

	if err := json.NewEncoder(gzWriter).Encode(metricsBatch); err != nil {
		logger.Log.Error(
			"error on marshaling metrics batch",
			zap.Error(err),
		)
		return
	}
	gzWriter.Flush()

	url := fmt.Sprintf("%s%s", a.serverURL, updatesURI)
	resp, err := a.httpClient.R().
		SetHeader("Content-Encoding", "gzip").
		SetBody(metricsBatchJSON.Bytes()).
		Post(url)

	if err != nil {
		logger.Log.Error(
			"error on making metrics batch request",
			zap.Error(err),
		)
		return
	}

	logger.Log.Debug(
		"metrics batch sended",
		zap.Int("status", resp.StatusCode()),
	)
}
