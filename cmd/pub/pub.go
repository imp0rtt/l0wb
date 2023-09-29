package main

import (
	"github.com/nats-io/stan.go"
	"io/ioutil"
	"log"
	"time"
)

const (
	serviceName = "server"
)

func main() {
	sc, _ := stan.Connect("test-cluster", "sm", stan.NatsURL("nats://localhost:4444"))
	log.Println("Connected")
	log.Println("Publishing...")
	if err := publish(sc, serviceName, "./cmd/pub/valid_test.json"); err != nil {
		log.Fatal(err)
	}
	time.Sleep(10000)
	if err := publish(sc, serviceName, "./cmd/pub/valid_test2.json"); err != nil {
		log.Fatal(err)
	}
}

func publish(sc stan.Conn, name string, filePath string) error {
	// Чтение JSON-файла
	jsonPayload, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Ошибка чтения JSON-файла: %v", err)
	}

	err = sc.Publish(name, jsonPayload)
	if err != nil {
		return err
	}
	return nil
}
