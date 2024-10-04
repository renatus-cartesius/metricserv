package agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/renatus-cartesius/metricserv/internal/monitor"
)

type Agent struct {
	monitor      monitor.Monitor
	pollInterval int
	serverUrl    string
	httpClient   *http.Client
}

func NewAgent(pollInterval int, serverUrl string, monitor monitor.Monitor) *Agent {
	return &Agent{
		monitor:      monitor,
		pollInterval: pollInterval,
		serverUrl:    serverUrl,
		httpClient:   &http.Client{},
	}
}

func (a *Agent) Serve() error {

	for {
		if err := a.monitor.Flush(); err != nil {
			fmt.Errorf("%v")
		}
		stats := a.monitor.Get()
		for m, v := range stats {

			url := fmt.Sprintf("%s/update/gauge/%s/%s", a.serverUrl, m, v)
			req, err := http.NewRequest(
				http.MethodPost,
				url,
				nil,
			)
			if err != nil {
				return fmt.Errorf("%v", err)
			}

			res, err := a.httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("%v", err)
			}
			res.Body.Close()

			fmt.Println("Sended: ", url, res.StatusCode)

		}

		url := fmt.Sprintf("%s/update/counter/%s/%d", a.serverUrl, "PollCount", 1)
		req, err := http.NewRequest(
			http.MethodPost,
			url,
			nil,
		)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		res, err := a.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		res.Body.Close()
		fmt.Println("Sended: ", url, res.StatusCode)

		time.Sleep(time.Duration(a.pollInterval) * time.Second)
	}
}
