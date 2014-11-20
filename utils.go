package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
)

var (
	decimapAbbrs = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
)

type QueueOpts struct {
	LookupdAddr string
	Topic       string
	Channel     string
	Concurrent  int
	Signals     []os.Signal
}

func QueueOptsFromContext(topic, channel, lookupd string) QueueOpts {
	return QueueOpts{
		Signals:     []os.Signal{syscall.SIGTERM, syscall.SIGINT},
		LookupdAddr: lookupd,
		Topic:       topic,
		Channel:     channel,
		Concurrent:  1,
	}
}

func ProcessQueue(handler nsq.Handler, opts QueueOpts) error {
	if opts.Concurrent == 0 {
		opts.Concurrent = 1
	}
	s := make(chan os.Signal, 64)
	signal.Notify(s, opts.Signals...)

	consumer, err := nsq.NewConsumer(opts.Topic, opts.Channel, nsq.NewConfig())
	if err != nil {
		return err
	}
	consumer.AddConcurrentHandlers(handler, opts.Concurrent)
	if err := consumer.ConnectToNSQLookupd(opts.LookupdAddr); err != nil {
		return err
	}

	for {
		select {
		case <-consumer.StopChan:
			return nil
		case sig := <-s:
			log.WithField("signal", sig).Debug("received signal")
			consumer.Stop()
		}
	}
	return nil
}

func getBinaryVersion(temp string) (version string, err error) {
	file, err := ioutil.ReadFile(path.Join(temp, "VERSION"))
	if err != nil {
		return version, err
	}

	return strings.TrimSpace(string(file)), nil
}

// HumanSize returns a human-readable approximation of a size
// using SI standard (eg. "44kB", "17MB")
func humanSize(size int64) string {
	return intToString(float64(size), 1000.0, decimapAbbrs)
}

func intToString(size, unit float64, _map []string) string {
	i := 0
	for size >= unit {
		size = size / unit
		i++
	}
	return fmt.Sprintf("%.4g %s", size, _map[i])
}
