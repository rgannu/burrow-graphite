package burrow_graphite

import (
	"context"

	"sync"
	"time"
	"strconv"

  log "github.com/Sirupsen/logrus"
	"github.com/marpaia/graphite-golang"
)

type BurrowExporter struct {
	client            *BurrowClient
	graphiteHost      string
	graphitePort      int
	interval          int
	wg                sync.WaitGroup
}

func (be *BurrowExporter) processGroup(cluster, group string, gh *graphite.Graphite) {
	lag, err := be.client.ConsumerGroupLag(cluster, group)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"group": group,
		}).Error("error getting lag for consumer group. returning.")
		return
	}
  metrics := make([]graphite.Metric)

	for _, partition := range lag.Status.Partitions {
	  metricNamePrefix := "kafka" + "." + lag.Status.Cluster + "." + "group" + "." + lag.Status.Group + "." + "topic" + "." + partition.Topic + "." + strconv.Itoa(int(partition.Partition))
	  metrics = append(metrics, graphite.NewMetric(metricNamePrefix + "." + "Lag", strconv.Itoa(int(partition.End.Lag)), time.Now().Unix()))
	  metrics = append(metrics, graphite.NewMetric(metricNamePrefix + "." + "Offset", strconv.Itoa(int(partition.End.Offset)), time.Now().Unix()))
	}

  totalLagMetricName := "kafka" + "." + lag.Status.Cluster + "." + "group" + "." + lag.Status.Group + "." + "TotalLag"
  metrics = append(metrics, graphite.NewMetric(totalLagMetricName, strconv.Itoa(int(lag.Status.TotalLag))))
  SendMetrics(gh, metrics)
  if err != nil {
      log.WithFields(log.Fields{
        "err": err,
      }).Error("Error in sending metrics to Graphite. returning.")
      return
  }
}

func (be *BurrowExporter) processTopic(cluster, topic string, gh *graphite.Graphite) {
	details, err := be.client.ClusterTopicDetails(cluster, topic)
	if err != nil {
		log.WithFields(log.Fields{
			"err":   err,
			"topic": topic,
		}).Error("error getting status for cluster topic. returning.")
		return
	}

	for i, offset := range details.Offsets {
    metricName := "kafka" + "." + cluster + "." + "topic" + "." + topic + "." + strconv.Itoa(i) + "." + "offset"
    err := gh.SimpleSend(metricName, strconv.Itoa(int(offset)))
    if err != nil {
        log.WithFields(log.Fields{
          "err": err,
        }).Error("Error in sending metrics to Graphite. returning.")
        return
    }
	}
}

func (be *BurrowExporter) processCluster(cluster string, gh *graphite.Graphite) {
	groups, err := be.client.ListConsumers(cluster)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"cluster": cluster,
		}).Error("error listing consumer groups. returning.")
		return
	}

	topics, err := be.client.ListClusterTopics(cluster)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"cluster": cluster,
		}).Error("error listing cluster topics. returning.")
		return
	}

	wg := sync.WaitGroup{}

	for _, group := range groups.ConsumerGroups {
		wg.Add(1)

		go func(g string) {
			defer wg.Done()
			be.processGroup(cluster, g, gh)
		}(group)
	}

	for _, topic := range topics.Topics {
		wg.Add(1)

		go func(t string) {
			defer wg.Done()
			be.processTopic(cluster, t, gh)
		}(topic)
	}

	wg.Wait()
}

func (be *BurrowExporter) Close() {
	be.wg.Wait()
}

func (be *BurrowExporter) Start(ctx context.Context) {
	be.wg.Add(1)
	defer be.wg.Done()

	be.mainLoop(ctx)
}

func (be *BurrowExporter) scrape() {
	start := time.Now()
	log.WithField("timestamp", start.UnixNano()).Info("Scraping burrow...")

	log.Info("Getting connection to the Graphite server")
	gh, err := GetGraphiteConnection(be.graphiteHost, be.graphitePort)
  if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("error Connnecting to Graphite server. Continuing.")
		return
	}

	clusters, err := be.client.ListClusters()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("error listing clusters. Continuing.")
		return
	}

	wg := sync.WaitGroup{}

	for _, cluster := range clusters.Clusters {
		wg.Add(1)

		go func(c string) {
			defer wg.Done()
			be.processCluster(c, gh)
		}(cluster)
	}

	wg.Wait()

  log.Info("Sleeping for 5 secs to ensure that metrics are sent to Graphite...")
  time.Sleep(5 * time.Second)
  log.Info("Closing the connection to the Graphite server")
	err = CloseGraphiteConnection(gh)
	if err != nil {
  		log.WithFields(log.Fields{
  			"err": err,
  		}).Error("error Closing the connection to Graphite server. Continuing.")
  		return
  }

	end := time.Now()
	log.WithFields(log.Fields{
		"timestamp": end.UnixNano(),
		"took":      end.Sub(start),
	}).Info("Finished scraping burrow.")
}

func (be *BurrowExporter) mainLoop(ctx context.Context) {
	timer := time.NewTicker(time.Duration(be.interval) * time.Second)

	// scrape at app start without waiting for the first interval to elapse
	be.scrape()

	for {
		select {
		case <-ctx.Done():
			log.Info("Shutting down exporter.")
			timer.Stop()
			return

		case <-timer.C:
			be.scrape()
		}
	}
}

func MakeBurrowExporter(burrowUrl string, graphiteHost string, graphitePort int, interval int) *BurrowExporter {
	return &BurrowExporter{
		client:            MakeBurrowClient(burrowUrl),
		graphiteHost:      graphiteHost,
		graphitePort:      graphitePort,
		interval:          interval,
	}
}
