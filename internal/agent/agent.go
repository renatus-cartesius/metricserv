package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/renatus-cartesius/metricserv/internal/logger"
	"github.com/renatus-cartesius/metricserv/internal/metrics"
	"github.com/renatus-cartesius/metricserv/internal/monitor"
	"github.com/renatus-cartesius/metricserv/internal/server/models"
)

const (
	updateURI = "/update"
)

type Agent struct {
	monitor        monitor.Monitor
	reportInterval int
	pollInterval   int
	serverURL      string
	httpClient     *http.Client
	exitCh         chan os.Signal
}

func NewAgent(repoInterval, pollInterval int, serverURL string, monitor monitor.Monitor, exitCh chan os.Signal) *Agent {
	return &Agent{
		monitor:        monitor,
		reportInterval: repoInterval,
		pollInterval:   pollInterval,
		serverURL:      serverURL,
		httpClient:     &http.Client{},
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

	if err := json.NewEncoder(&metricJSON).Encode(metric); err != nil {
		logger.Log.Error(
			"error on marshaling metric",
			zap.String("metric", metricJSON.String()),
			zap.String("metricID", metric.ID),
			zap.Error(err),
		)
		return
	}

	metricsDebug := metricJSON.Bytes()

	url := fmt.Sprintf("%s%s", a.serverURL, updateURI)
	req, err := http.NewRequest(
		http.MethodPost,
		url,
		&metricJSON,
	)

	if err != nil {
		logger.Log.Error(
			"error on preparing report request",
			zap.String("metric", metricJSON.String()),
			zap.Error(err),
		)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := a.httpClient.Do(req)
	if err != nil {
		logger.Log.Error(
			"error on sending metric",
			zap.String("metric", metricJSON.String()),
			zap.Error(err),
		)
		time.Sleep(time.Duration(a.pollInterval) * time.Second)
		return
	}
	defer res.Body.Close()

	logger.Log.Debug(
		"metric sended",
		zap.String("metric", string(metricsDebug)),
		zap.Int("status", res.StatusCode),
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

		var metricJSON bytes.Buffer

		if err := json.NewEncoder(&metricJSON).Encode(metric); err != nil {
			logger.Log.Error(
				"error on marshaling metric",
				zap.String("metric", metricJSON.String()),
				zap.String("metricID", metric.ID),
				zap.Error(err),
			)
			return
		}

		metricsDebug := metricJSON.Bytes()

		url := fmt.Sprintf("%s%s", a.serverURL, updateURI)
		req, err := http.NewRequest(
			http.MethodPost,
			url,
			&metricJSON,
		)

		if err != nil {
			logger.Log.Error(
				"error on preparing report request",
				zap.String("metric", string(metricsDebug)),
				zap.Error(err),
			)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		res, err := a.httpClient.Do(req)
		if err != nil {
			logger.Log.Error(
				"error on sending metric",
				zap.String("metric", string(metricsDebug)),
				zap.Error(err),
			)
			continue
		}
		defer res.Body.Close()

		logger.Log.Debug(
			"metric sended",
			zap.String("metric", string(metricsDebug)),
			zap.Int("status", res.StatusCode),
		)

	}
}
