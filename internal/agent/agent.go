package agent

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/renatus-cartesius/metricserv/internal/monitor"
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

	reportTicker := time.NewTicker(time.Duration(a.reportInterval) * time.Second)
	pollTicker := time.NewTicker(time.Duration(a.pollInterval) * time.Second)

	for {
		select {
		case <-a.exitCh:
			fmt.Println("Shutting down agent...")
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
		fmt.Printf("%v", err)
	}

	req.Header.Set("Content-Type", "text/plain")
	res, err := a.httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error on sending metric %s:%d : %v\n", "PollCount", 1, err)
		time.Sleep(time.Duration(a.pollInterval) * time.Second)
	}
	res.Body.Close()
	fmt.Println("Sended: ", url, res.StatusCode)
}

func (a *Agent) Report() {
	if err := a.monitor.Flush(); err != nil {
		fmt.Printf("%v", err)
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
			fmt.Printf("%v", err)
		}

		res, err := a.httpClient.Do(req)
		if err != nil {
			fmt.Printf("Error on sending metric %s:%s : %v\n", m, v, err)
			continue
		}
		res.Body.Close()

		fmt.Println("Sended: ", url, res.StatusCode)

	}
}
