package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	hago "github.com/hewenyu/ha-go"
)

func main() {
	// Replace with your Home Assistant URL and API token
	baseURL := "http://homeassistant.local:8123"
	apiToken := "your_long_lived_access_token"
	wsURL := "ws://homeassistant.local:8123/api/websocket"

	// Create a new Home Assistant client
	client, err := hago.NewClient(baseURL, apiToken)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create an API instance
	api := hago.NewAPI(client)

	// Check if the API is available
	if err := api.CheckAPI(); err != nil {
		log.Fatalf("API is not available: %v", err)
	}
	log.Println("Home Assistant API is available!")

	// Get configuration
	config, err := api.GetConfig()
	if err != nil {
		log.Fatalf("Failed to get config: %v", err)
	}
	log.Printf("Home Assistant version: %v", config["version"])

	// Get all entity states
	states, err := api.GetStates()
	if err != nil {
		log.Fatalf("Failed to get states: %v", err)
	}
	log.Printf("Found %d entities", len(states))

	// Print the state of a specific entity
	if len(states) > 0 {
		entity := states[0]
		log.Printf("Entity %s state: %s", entity.EntityID, entity.State)
		log.Printf("Entity attributes: %v", entity.Attributes)
	}

	// Get a specific entity state
	// Replace "light.living_room" with an entity that exists in your setup
	entityID := "light.living_room"
	state, err := api.GetState(entityID)
	if err != nil {
		log.Printf("Failed to get state for %s: %v", entityID, err)
	} else {
		log.Printf("Entity %s state: %s", state.EntityID, state.State)
	}

	// Call a service
	// This example turns on a light
	err = api.CallService("light", "turn_on", map[string]interface{}{
		"entity_id":  "light.living_room",
		"brightness": 255,
		"color_name": "blue",
	})
	if err != nil {
		log.Printf("Failed to call service: %v", err)
	} else {
		log.Println("Service called successfully!")
	}

	// Example of WebSocket connection for real-time updates
	log.Println("Connecting to WebSocket...")
	wsClient := hago.NewWSClient(wsURL, apiToken)
	err = wsClient.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer wsClient.Close()

	// Subscribe to state changes
	log.Println("Subscribing to state changes...")
	err = wsClient.SubscribeEvents("state_changed")
	if err != nil {
		log.Fatalf("Failed to subscribe to events: %v", err)
	}

	// Register event handler for state changes
	wsClient.AddEventHandler("event", func(msg map[string]interface{}) {
		if event, ok := msg["event"].(map[string]interface{}); ok {
			if eventType, ok := event["event_type"].(string); ok && eventType == "state_changed" {
				if data, ok := event["data"].(map[string]interface{}); ok {
					entityID, _ := data["entity_id"].(string)
					newState, _ := data["new_state"].(map[string]interface{})
					state, _ := newState["state"].(string)
					log.Printf("Entity %s changed to %s", entityID, state)
				}
			}
		}
	})

	// Wait for interrupt signal to gracefully shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Listening for state changes (Press Ctrl+C to exit)...")
	<-sigCh
	log.Println("Shutting down...")
}
