package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"strings"

	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v2"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Config struct {
	Responses map[string]string `yaml:"responses"`
}

func readConfigFromFile(filePath string) (*Config, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, config *Config) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	fmt.Println("Client Connected")

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		clientMessage := strings.TrimSuffix(string(p), "\n")
		fmt.Printf("Received message: %s\n", clientMessage)

		response, exists := config.Responses[clientMessage]
		if !exists {
			response = "No response found for the given message."
		}

		err = conn.WriteMessage(messageType, []byte(response))
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func main() {
	port := "8080"
	filePath := "responses.yaml"

	config, err := readConfigFromFile(filePath)
	if err != nil {
		log.Fatal("Error reading config from file:", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			handleWebSocket(w, r, config)
		})

		fmt.Printf("Server is listening on :%s\n", port)
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	// Wait for the goroutine to finish (this will never happen unless there's an error)
	wg.Wait()
}

