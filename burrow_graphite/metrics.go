package burrow_graphite

import (
	"fmt"
	"github.com/marpaia/graphite-golang"
)

func GetGraphiteConnection(graphiteHost string, graphitePort int) (*graphite.Graphite, error) {
	gh, err := graphite.NewGraphite(graphiteHost, graphitePort)
  if err != nil {
    fmt.Println("Error in getting the Graphite connection", err)
    return nil, err
  }
  return gh, nil
}

func CloseGraphiteConnection(graphite *graphite.Graphite) error {
  err := graphite.Disconnect()
  if err != nil {
    fmt.Println("Error in closing the Graphite connection", err)
    return err
  }
  return nil
}

func SendMetrics(graphite *graphite.Graphite, metrics []Metric) error {
  return graphite.sendMetrics(metrics)
}
