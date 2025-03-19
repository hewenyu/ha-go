# ha-go

A Go SDK for interacting with the Home Assistant API. This library enables Go applications to communicate with Home Assistant instances through both REST API and WebSocket connections.

## Features

- Complete REST API client for Home Assistant
- WebSocket support for real-time updates
- Support for all Home Assistant API endpoints:
  - Entity state management
  - Service calls
  - Event subscription
  - Configuration and system information

## Installation

```bash
go get github.com/hewenyu/ha-go
```

## Quick Start

```go
package main

import (
	"log"
	"github.com/hewenyu/ha-go"
)

func main() {
	// Initialize the client with your Home Assistant URL and API token
	client, err := hago.NewClient("http://homeassistant.local:8123", "your_long_lived_access_token")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create an API instance
	api := hago.NewAPI(client)

	// Get states of all entities
	states, err := api.GetStates()
	if err != nil {
		log.Fatalf("Failed to get states: %v", err)
	}

	// Print the state of each entity
	for _, state := range states {
		log.Printf("Entity %s: %s", state.EntityID, state.State)
	}

	// Call a service
	err = api.CallService("light", "turn_on", map[string]interface{}{
		"entity_id": "light.living_room",
		"brightness": 255,
	})
	if err != nil {
		log.Fatalf("Failed to call service: %v", err)
	}
}
```

## Authentication

To use this SDK, you need a long-lived access token from your Home Assistant instance. You can create one by following these steps:

1. In your Home Assistant dashboard, click on your profile icon at the bottom left
2. Scroll down to the "Long-Lived Access Tokens" section
3. Click on "Create Token"
4. Give it a name that helps you identify what it's used for
5. Copy the token (it will only be shown once)

## WebSocket Support

The SDK includes support for Home Assistant's WebSocket API, which allows you to receive real-time updates about state changes and events:

```go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/hewenyu/ha-go"
)

func main() {
	// Initialize WebSocket client
	wsClient := hago.NewWSClient("ws://homeassistant.local:8123/api/websocket", "your_long_lived_access_token")
	
	// Connect to Home Assistant
	err := wsClient.Connect()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()
	
	// Subscribe to state changes
	err = wsClient.SubscribeEvents("state_changed")
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	
	// Register event handler
	wsClient.AddEventHandler("event", func(msg map[string]interface{}) {
		if event, ok := msg["event"].(map[string]interface{}); ok {
			if eventType, ok := event["event_type"].(string); ok && eventType == "state_changed" {
				data := event["data"].(map[string]interface{})
				entityID := data["entity_id"].(string)
				newState := data["new_state"].(map[string]interface{})
				state := newState["state"].(string)
				log.Printf("Entity %s changed to %s", entityID, state)
			}
		}
	})
	
	// Wait for Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
```

## API Reference

### Client

```go
// Create a new client
client, err := hago.NewClient(baseURL, apiToken)

// REST API methods
api := hago.NewAPI(client)
```

### State Management

```go
// Get all states
states, err := api.GetStates()

// Get state for a specific entity
state, err := api.GetState("light.living_room")

// Update an entity's state
newState, err := api.SetState(
    "input_boolean.test", 
    "on", 
    map[string]interface{}{"friendly_name": "Test Boolean"}
)
```

### Service Calls

```go
// Call a service
err := api.CallService("light", "turn_on", map[string]interface{}{
    "entity_id": "light.living_room",
    "brightness": 255,
    "color_name": "blue",
})

// Get all available services
services, err := api.GetServices()
```

### Events

```go
// Get available event types
eventTypes, err := api.GetEvents()

// Fire an event
err := api.FireEvent("my_custom_event", map[string]interface{}{
    "some_data": "value",
})
```

### WebSocket Connection

```go
// Create WebSocket client
wsClient := hago.NewWSClient(wsURL, apiToken)

// Connect to Home Assistant
err := wsClient.Connect()

// Subscribe to events
err := wsClient.SubscribeEvents("state_changed")

// Add event handler
wsClient.AddEventHandler("event", func(msg map[string]interface{}) {
    // Handle event
})

// Send custom message
err := wsClient.Send(map[string]interface{}{
    "type": "custom_type",
    "data": "value",
})

// Close connection when done
wsClient.Close()
```

## Complete Example

See the [example](./example/main.go) directory for a complete working example of the SDK.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 