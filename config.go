package logrus_influxdb

import (
	"time"
)

// Config handles InfluxDB configuration, Logrus tags and batching inserts to InfluxDB
type Config struct {
	// InfluxDB Configurations
	Host         string        `json:"influxdb_host"`
	Port         int           `json:"influxdb_port"`
	Timeout      time.Duration `json:"influxdb_timeout"`
	Bucket       string        `json:"influxdb_bucket"`
	Organization string        `json:"influx_organization"`
	Token        string        `json:"influx_token"`
	UseHTTPS     bool          `json:"influxdb_https"`
	Precision    time.Duration `json:"influxdb_precision"`

	// Enable syslog format for chronograf logviewer usage
	Syslog       bool   `json:"syslog_enabled"`
	Facility     string `json:"syslog_facility"`
	FacilityCode int    `json:"syslog_facility_code"`
	AppName      string `json:"syslog_app_name"`
	Version      string `json:"syslog_app_version"`

	// Minimum level for push
	MinLevel string `json:"syslog_min_level"`

	// Logrus tags
	Tags []string `json:"logrus_tags"`

	// Defaults
	Measurement string `json:"measurement"`

	// Batching
	BatchIntervalMs uint `json:"batch_interval"` // Defaults to 5s.
	BatchCount      uint `json:"batch_count"`    // Defaults to 200.
}

// Set the default configurations
func (c *Config) defaults() {
	if c.Host == "" {
		c.Host = defaultHost
	}
	if c.Port == 0 {
		c.Port = defaultPort
	}
	if c.Timeout == 0 {
		c.Timeout = 100 * time.Millisecond
	}
	if c.Bucket == "" {
		c.Bucket = defaultBucket
	}
	//if c.Username == "" {
	//	c.Username = os.Getenv("INFLUX_USER")
	//}
	//if c.Password == "" {
	//	c.Password = os.Getenv("INFLUX_PWD")
	//}
	if c.Precision == time.Duration(0) {
		c.Precision = defaultPrecision
	}
	if c.Tags == nil {
		c.Tags = []string{}
	}
	if c.Measurement == "" {
		c.Measurement = defaultMeasurement
	}
	if c.BatchCount <= 0 {
		c.BatchCount = defaultBatchCount
	}
	if c.BatchIntervalMs <= 0 {
		c.BatchIntervalMs = defaultBatchIntervalMs
	}
}
