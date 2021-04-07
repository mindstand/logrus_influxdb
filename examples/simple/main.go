package main

import (
	"bufio"
	"github.com/Abramovic/logrus_influxdb"
	"github.com/sirupsen/logrus"
	_log "log"
	"os"
)

func main() {
	log := logrus.New()

	hook, err := logrus_influxdb.NewInfluxDB(&logrus_influxdb.Config{
		Host:         "0.0.0.0",
		Port:         8086,
		Bucket:       "some_bucket",
		Organization: "some_org",
		Token:        "random_token",
	})
	if err == nil {
		log.Hooks.Add(hook)
	} else {
		_log.Fatal(err)
	}

	defer hook.Close()

	go func() {
		for err := range hook.ErrorChannel {
			_log.Println(err)
		}
	}()

	log.Info("test")

	in := bufio.NewScanner(os.Stdin)
	in.Scan()
}
