package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

func RunWebSocketExample() {
	fmt.Println("Connecting to tx-indexer WebSocket...")

	// Build WebSocket URL - replace localhost:8546 with your indexer's address
	u := url.URL{Scheme: "ws", Host: "localhost:8546", Path: "/graphql/query"}

	// Establish WebSocket connection with GraphQL-WS protocol
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("WebSocket connection failed:", err)
	}
	defer conn.Close() // Clean up connection when function exits

	fmt.Println("Connected! Initializing GraphQL-WS connection...")

	// Step 1: Send connection_init message
	// GraphQL-WS protocol requires this handshake before subscriptions
	initMsg := map[string]interface{}{
		"type": "connection_init",
	}
	initBytes, _ := json.Marshal(initMsg)
	conn.WriteMessage(websocket.TextMessage, initBytes)

	// Step 2: Wait for connection_ack from server
	// Server must acknowledge our connection before we can subscribe
	_, ackMessage, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("Failed to receive connection ack:", err)
	}

	var ackResponse map[string]interface{}
	json.Unmarshal(ackMessage, &ackResponse)

	// Verify server sent the correct acknowledgment
	if ackResponse["type"] != "connection_ack" {
		log.Fatalf("Expected connection_ack, got: %+v", ackResponse)
	}

	fmt.Println("Connection acknowledged! Setting up subscription...")

	// Step 3: Send subscription message
	// This give the query to the server
	subscription := map[string]interface{}{
		"id":   "1",     // Unique ID for this subscription
		"type": "start", // GraphQL-WS message type for subscriptions
		"payload": map[string]interface{}{
			"query": `subscription { ... }`, // Your GraphQL subscription here
		},
	}

	subscriptionBytes, _ := json.Marshal(subscription)
	conn.WriteMessage(websocket.TextMessage, subscriptionBytes)

	fmt.Println("📡 Listening for new send transactions...")

	// Step 4: Listen for incoming messages in an infinite loop
	for {
		// Read next message from WebSocket
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			continue
		}

		// Debug: Show raw message (remove in production)
		fmt.Printf("Raw message: %s", string(message))

		// Parse JSON message from server
		var response map[string]interface{}
		err = json.Unmarshal(message, &response)
		if err != nil {
			log.Printf("JSON parse error: %v\n", err)
			continue
		}

		// Handle different message types from GraphQL-WS protocol
		switch response["type"] {
		case "data":
			// New transaction data received!
			// Extract payload and process transaction data
			data := response["payload"]
			dataBytes, _ := json.Marshal(data)
			parsedData, err := parseTransactions(dataBytes)
			if err != nil {
				log.Printf("Error parsing transactions: %v", err)
				continue
			}
			displayTransactions(parsedData) // Show formatted transaction details
		case "error":
			// GraphQL query/subscription error
			fmt.Printf("GraphQL error: %+v\n", response["payload"])

		case "complete":
			// Subscription finished
			fmt.Println("Subscription completed")

		case "connection_error":
			// WebSocket connection issue
			fmt.Printf("Connection error: %+v\n", response["payload"])

		case "ka":
			// Keep-alive message from server - ignore silently
			// Server sends these periodically to prevent connection timeouts
			continue

		default:
			// Unknown message type
			fmt.Printf("Unknown message type: %s\n", response["type"])
		}
	}
}
