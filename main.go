// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"strings"
	"time"

	nsq "github.com/nsqio/go-nsq"
	"github.com/openfaas-incubator/kafka-connector/types"
)

type connectorConfig struct {
	gatewayURL      string
	upstreamTimeout time.Duration
	topics          []string
	printResponse   bool
	rebuildInterval time.Duration
	broker          string
}

type ConsumerMessage struct {
	Topic string
	Value string
}

func main() {
	config := buildConnectorConfig()

	topicMap := types.NewTopicMap()

	lookupBuilder := types.FunctionLookupBuilder{
		GatewayURL: config.gatewayURL,
		Client:     types.MakeClient(config.upstreamTimeout),
	}

	ticker := time.NewTicker(config.rebuildInterval)
	go synchronizeLookups(ticker, &lookupBuilder, &topicMap)

	nsqAddr := config.broker + ":4150"
	makeConsumer(nsqAddr, config, &topicMap)
}

func makeConsumer(nsqAddr string, config connectorConfig, topicMap *types.TopicMap) {
	num := 0

	r, err := nsq.NewConsumer("test_topic", "openfaas-channel", nsq.NewConfig())
	if err != nil {
		log.Fatalf(err.Error())
	}

	mcb := makeMessageHandler(topicMap, config)

	r.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {

		var msg ConsumerMessage
		err := json.Unmarshal(m.Body, &msg)
		if err != nil {
			log.Println("consumer error: ", err)
			return nil
		}

		num = (num + 1) % math.MaxInt32
		log.Printf("[#%d] Received on [%v]: '%s'\n",
			num,
			msg.Topic,
			msg.Value)
		mcb(&msg)

		return nil
	}))

	err = r.ConnectToNSQD(nsqAddr)
	if err != nil {
		log.Fatalf(err.Error())
	}

	<-r.StopChan
}

func makeMessageHandler(topicMap *types.TopicMap, config connectorConfig) func(msg *ConsumerMessage) {

	invoker := types.Invoker{
		PrintResponse: config.printResponse,
		Client:        types.MakeClient(config.upstreamTimeout),
		GatewayURL:    config.gatewayURL,
	}

	mcb := func(msg *ConsumerMessage) {
		val := []byte(msg.Value)
		invoker.Invoke(topicMap, msg.Topic, &val)
	}
	return mcb
}

func buildConnectorConfig() connectorConfig {

	broker := "127.0.0.1"
	if val, exists := os.LookupEnv("broker_host"); exists {
		broker = val
	}

	topics := []string{}
	if val, exists := os.LookupEnv("topics"); exists {
		for _, topic := range strings.Split(val, ",") {
			if len(topic) > 0 {
				topics = append(topics, topic)
			}
		}
	}
	if len(topics) == 0 {
		log.Fatal(`Provide a list of topics i.e. topics="payment_published,slack_joined"`)
	}

	gatewayURL := "http://gateway:8080"
	if val, exists := os.LookupEnv("gateway_url"); exists {
		gatewayURL = val
	}

	upstreamTimeout := time.Second * 30
	rebuildInterval := time.Second * 3

	if val, exists := os.LookupEnv("upstream_timeout"); exists {
		parsedVal, err := time.ParseDuration(val)
		if err == nil {
			upstreamTimeout = parsedVal
		}
	}

	if val, exists := os.LookupEnv("rebuild_interval"); exists {
		parsedVal, err := time.ParseDuration(val)
		if err == nil {
			rebuildInterval = parsedVal
		}
	}

	printResponse := false
	if val, exists := os.LookupEnv("print_response"); exists {
		printResponse = (val == "1" || val == "true")
	}

	return connectorConfig{
		gatewayURL:      gatewayURL,
		upstreamTimeout: upstreamTimeout,
		topics:          topics,
		rebuildInterval: rebuildInterval,
		broker:          broker,
		printResponse:   printResponse,
	}
}

func synchronizeLookups(ticker *time.Ticker,
	lookupBuilder *types.FunctionLookupBuilder,
	topicMap *types.TopicMap) {

	for {
		<-ticker.C
		lookups, err := lookupBuilder.Build()
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("Syncing topic map")
		topicMap.Sync(&lookups)
	}
}