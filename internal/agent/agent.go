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
	serverURL    string
	httpClient   *http.Client
}

func NewAgent(pollInterval int, serverURL string, monitor monitor.Monitor) *Agent {
	return &Agent{
		monitor:      monitor,
		pollInterval: pollInterval,
		serverURL:    serverURL,
		httpClient:   &http.Client{},
	}
}

func (a *Agent) Serve() {

	for {
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
			continue
		}
		res.Body.Close()
		fmt.Println("Sended: ", url, res.StatusCode)

		time.Sleep(time.Duration(a.pollInterval) * time.Second)
	}
}
