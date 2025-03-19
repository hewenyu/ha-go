package hago

import (
	"fmt"
	"strings"
)

// Entities provides methods to interact with specific Home Assistant entity types
type Entities struct {
	api *API
}

// NewEntities creates a new Entities instance
func NewEntities(api *API) *Entities {
	return &Entities{
		api: api,
	}
}

// LightTurnOn turns on a light entity
func (e *Entities) LightTurnOn(entityID string, options map[string]interface{}) error {
	if !strings.HasPrefix(entityID, "light.") {
		entityID = "light." + entityID
	}

	data := map[string]interface{}{
		"entity_id": entityID,
	}

	// Merge options into data
	for k, v := range options {
		data[k] = v
	}

	return e.api.CallService("light", "turn_on", data)
}

// LightTurnOff turns off a light entity
func (e *Entities) LightTurnOff(entityID string) error {
	if !strings.HasPrefix(entityID, "light.") {
		entityID = "light." + entityID
	}

	data := map[string]interface{}{
		"entity_id": entityID,
	}

	return e.api.CallService("light", "turn_off", data)
}

// SwitchTurnOn turns on a switch entity
func (e *Entities) SwitchTurnOn(entityID string) error {
	if !strings.HasPrefix(entityID, "switch.") {
		entityID = "switch." + entityID
	}

	data := map[string]interface{}{
		"entity_id": entityID,
	}

	return e.api.CallService("switch", "turn_on", data)
}

// SwitchTurnOff turns off a switch entity
func (e *Entities) SwitchTurnOff(entityID string) error {
	if !strings.HasPrefix(entityID, "switch.") {
		entityID = "switch." + entityID
	}

	data := map[string]interface{}{
		"entity_id": entityID,
	}

	return e.api.CallService("switch", "turn_off", data)
}

// ClimateSetTemperature sets the temperature for a climate entity
func (e *Entities) ClimateSetTemperature(entityID string, temperature float64, options map[string]interface{}) error {
	if !strings.HasPrefix(entityID, "climate.") {
		entityID = "climate." + entityID
	}

	data := map[string]interface{}{
		"entity_id":   entityID,
		"temperature": temperature,
	}

	// Merge options into data
	for k, v := range options {
		data[k] = v
	}

	return e.api.CallService("climate", "set_temperature", data)
}

// ClimateSetHVACMode sets the HVAC mode for a climate entity
func (e *Entities) ClimateSetHVACMode(entityID string, hvacMode string) error {
	if !strings.HasPrefix(entityID, "climate.") {
		entityID = "climate." + entityID
	}

	data := map[string]interface{}{
		"entity_id": entityID,
		"hvac_mode": hvacMode,
	}

	return e.api.CallService("climate", "set_hvac_mode", data)
}

// CoverOpen opens a cover entity
func (e *Entities) CoverOpen(entityID string) error {
	if !strings.HasPrefix(entityID, "cover.") {
		entityID = "cover." + entityID
	}

	data := map[string]interface{}{
		"entity_id": entityID,
	}

	return e.api.CallService("cover", "open_cover", data)
}

// CoverClose closes a cover entity
func (e *Entities) CoverClose(entityID string) error {
	if !strings.HasPrefix(entityID, "cover.") {
		entityID = "cover." + entityID
	}

	data := map[string]interface{}{
		"entity_id": entityID,
	}

	return e.api.CallService("cover", "close_cover", data)
}

// CoverSetPosition sets the position of a cover entity
func (e *Entities) CoverSetPosition(entityID string, position int) error {
	if !strings.HasPrefix(entityID, "cover.") {
		entityID = "cover." + entityID
	}

	if position < 0 || position > 100 {
		return fmt.Errorf("position must be between 0 and 100")
	}

	data := map[string]interface{}{
		"entity_id": entityID,
		"position":  position,
	}

	return e.api.CallService("cover", "set_cover_position", data)
}

// MediaPlay plays media on a media player entity
func (e *Entities) MediaPlay(entityID string) error {
	if !strings.HasPrefix(entityID, "media_player.") {
		entityID = "media_player." + entityID
	}

	data := map[string]interface{}{
		"entity_id": entityID,
	}

	return e.api.CallService("media_player", "media_play", data)
}

// MediaPause pauses media on a media player entity
func (e *Entities) MediaPause(entityID string) error {
	if !strings.HasPrefix(entityID, "media_player.") {
		entityID = "media_player." + entityID
	}

	data := map[string]interface{}{
		"entity_id": entityID,
	}

	return e.api.CallService("media_player", "media_pause", data)
}

// MediaStop stops media on a media player entity
func (e *Entities) MediaStop(entityID string) error {
	if !strings.HasPrefix(entityID, "media_player.") {
		entityID = "media_player." + entityID
	}

	data := map[string]interface{}{
		"entity_id": entityID,
	}

	return e.api.CallService("media_player", "media_stop", data)
}

// ScriptRun runs a script entity
func (e *Entities) ScriptRun(entityID string, variables map[string]interface{}) error {
	// Remove "script." prefix if it exists
	scriptName := entityID
	if strings.HasPrefix(entityID, "script.") {
		scriptName = strings.TrimPrefix(entityID, "script.")
	}

	data := map[string]interface{}{}

	// Add variables if provided
	if variables != nil {
		data["variables"] = variables
	}

	return e.api.CallService("script", scriptName, data)
}

// SceneTurnOn activates a scene
func (e *Entities) SceneTurnOn(entityID string) error {
	if !strings.HasPrefix(entityID, "scene.") {
		entityID = "scene." + entityID
	}

	data := map[string]interface{}{
		"entity_id": entityID,
	}

	return e.api.CallService("scene", "turn_on", data)
}

// AutomationTrigger triggers an automation
func (e *Entities) AutomationTrigger(entityID string) error {
	if !strings.HasPrefix(entityID, "automation.") {
		entityID = "automation." + entityID
	}

	data := map[string]interface{}{
		"entity_id": entityID,
	}

	return e.api.CallService("automation", "trigger", data)
}

// GetSensor gets the state of a sensor
func (e *Entities) GetSensor(entityID string) (*State, error) {
	if !strings.HasPrefix(entityID, "sensor.") {
		entityID = "sensor." + entityID
	}

	return e.api.GetState(entityID)
}

// GetBinarySensor gets the state of a binary sensor
func (e *Entities) GetBinarySensor(entityID string) (*State, error) {
	if !strings.HasPrefix(entityID, "binary_sensor.") {
		entityID = "binary_sensor." + entityID
	}

	return e.api.GetState(entityID)
}
