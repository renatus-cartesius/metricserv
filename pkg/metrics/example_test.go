package metrics

import "fmt"

func Example() {

	metric := &GaugeMetric{
		ID:    "cpu_usage",
		Value: float64(10),
	}

	fmt.Println(metric)

	//Output
	//gauge:cpu_usage:10.000000
}
