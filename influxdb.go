package logrus_influxdb

import (
	"fmt"

	influxdb "github.com/influxdata/influxdb-client-go/v2"
)

// Returns an influxdb client
func newInfluxDBClient(config *Config) influxdb.Client {
	protocol := "http"
	if config.UseHTTPS {
		protocol = "https"
	}

	return influxdb.NewClientWithOptions(fmt.Sprintf("%s://%s:%v", protocol, config.Host, config.Port),
		config.Token,
		influxdb.DefaultOptions().
			SetBatchSize(config.BatchCount).
			SetFlushInterval(config.BatchIntervalMs).
			SetPrecision(config.Precision),
	)
}
