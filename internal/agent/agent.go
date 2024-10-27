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
	"github.com/renatus-cartesius/metricserv/internal/monitor"
	"github.com/renatus-cartesius/metricserv/internal/server/models"
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
	pollTicker := time.NewTicker(time.Duration(a.pollInterval) * time.Second)

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
		MType: "counter",
		Delta: new(int64),
	}

	*metric.Delta = 1

	var metricJSON bytes.Buffer

	if err := json.NewEncoder(&metricJSON).Encode(metric); err != nil {
		logger.Log.Error(
			"error on marshaling metric",
			zap.String("metric", metricJSON.String()),
			zap.Error(err),
		)
		return
	}

	metricsDebug := metricJSON.Bytes()

	url := fmt.Sprintf("%s/update", a.serverURL)
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
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := a.httpClient.Do(req)
	if err != nil {
		logger.Log.Error(
			"error on sending metric",
			zap.String("metric", string(metricsDebug)),
			zap.Error(err),
		)
		time.Sleep(time.Duration(a.pollInterval) * time.Second)
		return
	}
	err = res.Body.Close()
	if err != nil {
		logger.Log.Error(
			"error on closing response body",
			zap.String("metric", string(metricsDebug)),
			zap.Error(err),
		)
		return
	}
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

		value, _ := strconv.ParseFloat(v, 64)

		metric := &models.Metric{
			ID:    m,
			MType: "gauge",
			Value: &value,
		}

		var metricJSON bytes.Buffer

		if err := json.NewEncoder(&metricJSON).Encode(metric); err != nil {
			logger.Log.Error(
				"error on marshaling metric",
				zap.String("metric", metricJSON.String()),
				zap.Error(err),
			)
			return
		}

		metricsDebug := metricJSON.Bytes()

		url := fmt.Sprintf("%s/update", a.serverURL)
		req, err := http.NewRequest(
			http.MethodPost,
			url,
			&metricJSON,
		)

		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			logger.Log.Error(
				"error on preparing report request",
				zap.String("metric", string(metricsDebug)),
				zap.Error(err),
			)
		}

		res, err := a.httpClient.Do(req)
		if err != nil {
			logger.Log.Error(
				"error on sending metric",
				zap.String("metric", string(metricsDebug)),
				zap.Error(err),
			)
			continue
		}
		res.Body.Close()

		logger.Log.Debug(
			"metric sended",
			zap.String("metric", string(metricsDebug)),
			zap.Int("status", res.StatusCode),
		)

	}
}
