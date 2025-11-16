package main

import (
	"flag"
	"log"
	"sync"
	"time"

	tinymq "github.com/oo-developer/tinymq/pkg"
	"github.com/oo-developer/tinymq/src/application"
	"github.com/oo-developer/tinymq/src/config"
)

const (
	TOPIC_TEST1 = "test/test1/#"
	TOPIC_TEST2 = "test/test2/value"
)

func main() {
	serverConfigFile := flag.String("server-config", "server_config.json", "Path to server config file")
	clientConfigFile := flag.String("client-config", "client_config.json", "Path to client config file")
	flag.Parse()

	startServer(*serverConfigFile)

	clientConfig, err := tinymq.LoadConfig(*clientConfigFile)
	if err != nil {
		panic(err)
	}
	start := time.Now().UnixMilli()
	client, err := tinymq.NewClient(clientConfig)
	if err != nil {
		panic(err)
	}
	err = client.Connect()
	if err != nil {
		panic(err)
	}

	countReceived := 0
	client.Subscribe(TOPIC_TEST1, func(topic string, payload []byte) {
		countReceived++
		if countReceived%100000 == 0 {
			log.Printf("Received %d messages", countReceived)
		}
		if countReceived >= 1000000 {
			end := time.Now().UnixMilli()
			span := end - start
			msPerMsg := float64(span) / float64(1000000)
			log.Printf("Received %d messages in %.3f ms per message", countReceived, msPerMsg)
			client.Unsubscribe(TOPIC_TEST1)

		}
	})

	msg := "This is a short message."
	msg += msg
	log.Printf("Message size: %d ", len(msg))
	for ii := 0; ii < 1000000; ii++ {
		client.Publish(TOPIC_TEST1, []byte(msg))
		if ii%100000 == 0 {
			log.Printf("Sent %d messages", ii)
		}
	}

	for {
		time.Sleep(10 * time.Second)
	}
}

func startServer(configFile string) {
	configuration := config.Load(configFile)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		app := application.NewApplication(configuration)
		wg.Done()
		app.Start()
	}()
	wg.Wait()
	time.Sleep(1 * time.Second)
}
