package agent

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/renatus-cartesius/metricserv/internal/logger"
	"github.com/renatus-cartesius/metricserv/internal/monitor"
	"go.uber.org/zap"
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
	url := fmt.Sprintf("%s/update/counter/%s/%d", a.serverURL, "PollCount", 1)
	req, err := http.NewRequest(
		http.MethodPost,
		url,
		nil,
	)
	if err != nil {
		logger.Log.Error(
			"error on preparing report request",
			zap.String("metric", "PollCount"),
			zap.Error(err),
		)
	}

	req.Header.Set("Content-Type", "text/plain")
	res, err := a.httpClient.Do(req)
	if err != nil {
		logger.Log.Error(
			"error on sending metric",
			zap.String("metric", "PollCount"),
			zap.Error(err),
		)
		time.Sleep(time.Duration(a.pollInterval) * time.Second)
		return
	}
	err = res.Body.Close()
	if err != nil {
		logger.Log.Error(
			"error on closing response body",
			zap.String("metric", "PollCount"),
			zap.Error(err),
		)
		return
	}
	logger.Log.Debug(
		"metric sended",
		zap.String("metric", "PollCount"),
		zap.Int("value", 1),
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

		url := fmt.Sprintf("%s/update/gauge/%s/%s", a.serverURL, m, v)
		req, err := http.NewRequest(
			http.MethodPost,
			url,
			nil,
		)
		req.Header.Set("Content-Type", "text/plain")
		if err != nil {
			logger.Log.Error(
				"error on preparing report request",
				zap.String("metric", m),
				zap.Error(err),
			)
		}

		res, err := a.httpClient.Do(req)
		if err != nil {
			logger.Log.Error(
				"error on sending metric",
				zap.String("metric", m),
				zap.Error(err),
			)
			continue
		}
		res.Body.Close()

		logger.Log.Debug(
			"metric sended",
			zap.String("metric", m),
			zap.String("value", v),
			zap.Int("status", res.StatusCode),
		)

	}
}
