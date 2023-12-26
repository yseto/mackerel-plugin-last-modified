package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	mp "github.com/mackerelio/go-mackerel-plugin"
)

type LastModifiedPlugin struct {
	URLs   [][]string
	Prefix string
}

func main() {
	optFilename := flag.String("conf", "", "Config `file`")
	optPrefix := flag.String("metric-key-prefix", "last-modified", "Metric key prefix")
	flag.Parse()

	if *optFilename == "" {
		flag.Usage()
		os.Exit(1)
	}

	r, err := os.Open(*optFilename)
	if err != nil {
		fmt.Printf("config read error : %v\n", err)
		os.Exit(1)
	}

	records, err := csv.NewReader(r).ReadAll()
	if err != nil {
		fmt.Printf("config read error : %v\n", err)
		os.Exit(1)
	}

	n := LastModifiedPlugin{
		URLs:   records,
		Prefix: *optPrefix,
	}
	mp.NewMackerelPlugin(n).Run()
}

func (n LastModifiedPlugin) GraphDefinition() map[string]mp.Graphs {
	return map[string]mp.Graphs{
		"modified": {
			Label:   n.MetricKeyPrefix(),
			Unit:    mp.UnitInteger,
			Metrics: []mp.Metrics{{Name: "*", Label: "%1"}},
		},
		"status": {
			Label:   n.MetricKeyPrefix(),
			Unit:    mp.UnitInteger,
			Metrics: []mp.Metrics{{Name: "*", Label: "%1"}},
		},
	}
}

func (n LastModifiedPlugin) MetricKeyPrefix() string {
	if n.Prefix == "" {
		n.Prefix = "last-modified"
	}
	return n.Prefix
}

func (n LastModifiedPlugin) FetchMetrics() (map[string]float64, error) {
	now := time.Now()
	kv := make(map[string]float64)

	for _, url := range n.URLs {
		resp, err := http.Get(url[0])
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		kv["status."+url[1]] = float64(resp.StatusCode)

		if resp.StatusCode > 399 {
			continue
		}

		t, err := time.Parse(http.TimeFormat, resp.Header.Get("Last-Modified"))
		if err != nil {
			continue
		}

		kv["modified."+url[1]] = float64(now.Sub(t).Round(time.Second).Seconds())
	}
	return kv, nil
}
