package logrus_influxdb

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"os"
	"time"

	influxdb "github.com/influxdata/influxdb-client-go/v2"
	"github.com/sirupsen/logrus"
)

var (
	defaultHost            = "localhost"
	defaultPort            = 8086
	defaultBucket          = "logrus"
	defaultBatchIntervalMs = uint(5000)
	defaultMeasurement     = "logrus"
	defaultBatchCount      = uint(200)
	defaultPrecision       = time.Nanosecond
	defaultSyslog          = false
)

// InfluxDBHook delivers logs to an InfluxDB cluster.
type InfluxDBHook struct {
	client                   influxdb.Client
	writeAPI                 api.WriteAPI
	precision                time.Duration
	org, bucket, measurement string
	tagList                  []string
	lastBatchUpdate          time.Time
	batchIntervalMs          uint
	batchCount               uint
	syslog                   bool
	facility                 string
	facilityCode             int
	appName                  string
	version                  string
	minLevel                 string
	ErrorChannel             <-chan error
}

// NewInfluxDB returns a new InfluxDBHook.
func NewInfluxDB(config *Config, clients ...influxdb.Client) (hook *InfluxDBHook, err error) {
	if config == nil {
		config = &Config{}
	}
	config.defaults()

	var client influxdb.Client
	if len(clients) == 0 {
		client = newInfluxDBClient(config)
	} else if len(clients) == 1 {
		client = clients[0]
	} else {
		return nil, fmt.Errorf("NewInfluxDB: Error creating InfluxDB Client, %d is too many influxdb clients", len(clients))
	}

	ready, err := client.Ready(context.Background())
	if !ready || err != nil {
		return nil, fmt.Errorf("client is not available. ready=%v,err=%v", ready, err)
	}

	writeAPI := client.WriteAPI(config.Organization, config.Bucket)

	hook = &InfluxDBHook{
		client:          client,
		bucket:          config.Bucket,
		org:             config.Organization,
		measurement:     config.Measurement,
		tagList:         config.Tags,
		batchIntervalMs: config.BatchIntervalMs,
		batchCount:      config.BatchCount,
		precision:       config.Precision,
		syslog:          config.Syslog,
		facility:        config.Facility,
		facilityCode:    config.FacilityCode,
		appName:         config.AppName,
		version:         config.Version,
		minLevel:        config.MinLevel,
		writeAPI:        writeAPI,
		ErrorChannel:    writeAPI.Errors(),
	}

	return hook, nil
}

func (hook *InfluxDBHook) Close() {
	hook.writeAPI.Flush()
	hook.client.Close()
}

func parseSeverity(level string) (string, int) {
	switch level {
	case "info":
		return "info", 6
	case "error":
		return "err", 3
	case "debug":
		return "debug", 7
	case "panic":
		return "panic", 0
	case "fatal":
		return "crit", 2
	case "warning":
		return "warning", 4
	}

	return "none", -1
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (hook *InfluxDBHook) hasMinLevel(level string) bool {
	if len(hook.minLevel) > 0 {
		if hook.minLevel == "debug" {
			return true
		}

		if hook.minLevel == "info" {
			return stringInSlice(level, []string{"info", "warning", "error", "fatal", "panic"})
		}

		if hook.minLevel == "warning" {
			return stringInSlice(level, []string{"warning", "error", "fatal", "panic"})
		}

		if hook.minLevel == "error" {
			return stringInSlice(level, []string{"error", "fatal", "panic"})
		}

		if hook.minLevel == "fatal" {
			return stringInSlice(level, []string{"fatal", "panic"})
		}

		if hook.minLevel == "panic" {
			return level == "panic"
		}

		return false
	}

	return true
}

// Fire adds a new InfluxDB point based off of Logrus entry
func (hook *InfluxDBHook) Fire(entry *logrus.Entry) (err error) {
	if hook.hasMinLevel(entry.Level.String()) {
		measurement := hook.measurement
		if result, ok := getTag(entry.Data, "measurement"); ok {
			measurement = result
		}

		tags := make(map[string]string)
		data := make(map[string]interface{})

		if hook.syslog {
			hostname, err := os.Hostname()

			if err != nil {
				return err
			}

			severity, severityCode := parseSeverity(entry.Level.String())

			tags["appname"] = hook.appName
			tags["facility"] = hook.facility
			tags["host"] = hostname
			tags["hostname"] = hostname
			tags["severity"] = severity

			data["facility_code"] = hook.facilityCode
			data["message"] = entry.Message
			data["procid"] = os.Getpid()
			data["severity_code"] = severityCode
			data["timestamp"] = entry.Time.UnixNano()
			data["version"] = hook.version
		} else {
			// If passing a "message" field then it will be overridden by the entry Message
			entry.Data["message"] = entry.Message

			// Set the level of the entry
			tags["level"] = entry.Level.String()
			// getAndDel and getAndDelRequest are taken from https://github.com/evalphobia/logrus_sentry
			if logger, ok := getTag(entry.Data, "logger"); ok {
				tags["logger"] = logger
			}

			for k, v := range entry.Data {
				data[k] = v
			}

			for _, tag := range hook.tagList {
				if tagValue, ok := getTag(entry.Data, tag); ok {
					tags[tag] = tagValue
					delete(data, tag)
				}
			}
		}

		hook.writeAPI.WritePoint(influxdb.NewPoint(measurement, tags, data, entry.Time))
	}

	return nil
}
