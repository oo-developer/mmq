package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	mmq "github.com/oo-developer/mmq/pkg"
	testtools "github.com/oo-developer/mmq/test"
)

const (
	TOPIC_TEST1 = "test/test1/#"
)

func main() {
	read := flag.Bool("read", false, "read topic")
	serverConfigFile := flag.String("server-config", "server_config.json", "Path to server config file")
	clientConfigFile := flag.String("client-config", "client_config.json", "Path to client config file")
	flag.Parse()

	server := testtools.StartServer(*serverConfigFile)

	clientConfig, err := mmq.LoadConfig(*clientConfigFile)
	if err != nil {
		panic(err)
	}
	client, err := mmq.NewClient(clientConfig)
	if err != nil {
		panic(err)
	}
	err = client.Connect()
	if err != nil {
		panic(err)
	}

	if !*read {
		client.Publish(TOPIC_TEST1, []byte("Retained test"), mmq.Persistent)
	} else {
		wg := sync.WaitGroup{}
		wg.Add(1)
		client.Subscribe(TOPIC_TEST1, func(topic string, payload []byte) {
			fmt.Printf("[OK] topic:%s,payload:%s\n", topic, string(payload))
			wg.Done()
		})
		wg.Wait()
		time.Sleep(5 * time.Second)
	}
	client.Disconnect()
	client.Unsubscribe(TOPIC_TEST1)
	server.Shutdown()
}
