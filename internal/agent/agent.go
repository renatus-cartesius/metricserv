package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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
	hashKey        string
}

func NewAgent(repoInterval, pollInterval int, serverURL string, monitor monitor.Monitor, exitCh chan os.Signal, hashKey string) *Agent {

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
		hashKey:        hashKey,
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
	payload := metricJSON.Bytes()
	req := a.httpClient.R()

	if a.hashKey != "" {
		hash := hmac.New(sha256.New, []byte(a.hashKey))
		hash.Write(metricJSON.Bytes())

		req.SetHeader("HashSHA256", base64.StdEncoding.EncodeToString(hash.Sum(nil)))
	}

	req.SetHeader("Content-Encoding", "gzip").SetBody(payload)

	resp, err := req.Post(url)

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
		zap.String("sum", req.Header.Get("HashSHA256")),
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
				"error when parsing float value",
				zap.Error(err),
			)
			continue
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
	payload := metricsBatchJSON.Bytes()
	req := a.httpClient.R()

	if a.hashKey != "" {
		hash := hmac.New(sha256.New, []byte(a.hashKey))
		hash.Write(metricsBatchJSON.Bytes())

		req.SetHeader("HashSHA256", base64.StdEncoding.EncodeToString(hash.Sum(nil)))
	}

	req.SetHeader("Content-Encoding", "gzip").SetBody(payload)

	resp, err := req.Post(url)

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
		zap.String("sum", req.Header.Get("HashSHA256")),
	)
}
