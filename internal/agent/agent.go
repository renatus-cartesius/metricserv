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
	"runtime"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/renatus-cartesius/metricserv/internal/logger"
	"github.com/renatus-cartesius/metricserv/internal/metrics"
	"github.com/renatus-cartesius/metricserv/internal/monitor"
	"github.com/renatus-cartesius/metricserv/internal/server/models"
	"github.com/renatus-cartesius/metricserv/pkg/workerpool"
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
	exitCh         chan os.Signal
	hashKey        string
	reportCh       chan *monitor.RuntimeMetric
	workersPool    *workerpool.Pool
}

func NewAgent(repoInterval, pollInterval int, serverURL string, mon monitor.Monitor, exitCh chan os.Signal, hashKey string) (*Agent, error) {

	httpClient := resty.New()
	httpClient.
		SetRetryCount(3).
		AddRetryCondition(
			func(r *resty.Response, err error) bool {
				return r.StatusCode() == http.StatusTooManyRequests
			},
		)

	pool, err := workerpool.NewPool()
	if err != nil {
		return nil, err
	}

	return &Agent{
		monitor:        mon,
		reportInterval: repoInterval,
		pollInterval:   pollInterval,
		serverURL:      serverURL,
		httpClient:     httpClient,
		exitCh:         exitCh,
		hashKey:        hashKey,
		reportCh:       make(chan *monitor.RuntimeMetric, runtime.NumCPU()),
		workersPool:    pool,
	}, nil
}

func (a *Agent) Serve(reportWorkers int) {

	logger.Log.Info("starting agent")

	logger.Log.Debug(
		"creating workers",
		zap.Int("count", reportWorkers),
	)
	a.workersPool.Listen(reportWorkers, func() {
		a.StartReportWorker()
	})

	defer a.workersPool.Wait()

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
			close(a.reportCh)
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

	resp, err := a.Send(metric, updateURI)

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

func (a *Agent) StartReportWorker() {
	uuid := uuid.NewString()
	for runtimeMetric := range a.reportCh {
		value, err := strconv.ParseFloat(runtimeMetric.Value, 64)
		if err != nil {
			logger.Log.Error(
				"error when parsing float value",
				zap.String("worker", uuid),
				zap.Error(err),
			)
			continue
		}

		metric := &models.Metric{
			ID:    runtimeMetric.Name,
			MType: metrics.TypeGauge,
			Value: &value,
		}

		resp, err := a.Send(metric, updateURI)

		if err != nil {
			logger.Log.Error(
				"error on making metric request",
				zap.String("worker", uuid),
				zap.Error(err),
			)
			continue
		}

		logger.Log.Debug(
			"metric sended",
			zap.Int("status", resp.StatusCode()),
			zap.String("metric", metric.ID),
			zap.String("worker", uuid),
		)
	}
	logger.Log.Debug(
		"shutting down report worker",
		zap.String("worker", uuid),
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
		a.reportCh <- &monitor.RuntimeMetric{Name: m, Value: v}
	}

	v, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Error(
			"error on calling gopsutil",
			zap.Error(err),
		)
	}

	a.reportCh <- &monitor.RuntimeMetric{Name: "TotalMemory", Value: fmt.Sprintf("%v", float64(v.Total))}
	a.reportCh <- &monitor.RuntimeMetric{Name: "FreeMemory", Value: fmt.Sprintf("%v", float64(v.Free))}

	c, err := cpu.Percent(0, false)
	if err != nil {
		logger.Log.Error(
			"error on calling gopsutil",
			zap.Error(err),
		)
	}
	a.reportCh <- &monitor.RuntimeMetric{Name: "CPUutilization1", Value: fmt.Sprintf("%v", float64(c[0]))}

}

func (a *Agent) Send(obj interface{}, uri string) (*resty.Response, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)

	if err := json.NewEncoder(gzWriter).Encode(obj); err != nil {
		return nil, err
	}

	if err := gzWriter.Flush(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%s", a.serverURL, uri)
	payload := buf.Bytes()
	req := a.httpClient.R()

	if a.hashKey != "" {
		hash := hmac.New(sha256.New, []byte(a.hashKey))
		hash.Write(buf.Bytes())

		req.SetHeader("HashSHA256", base64.StdEncoding.EncodeToString(hash.Sum(nil)))
	}

	req.SetHeader("Content-Encoding", "gzip").SetBody(payload)

	return req.Post(url)
}
