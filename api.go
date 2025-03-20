package hago

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// API provides methods to interact with Home Assistant API endpoints
type API struct {
	client *Client
}

// NewAPI creates a new API instance with the provided client
func NewAPI(client *Client) *API {
	return &API{
		client: client,
	}
}

// GetStates returns the states of all entities
func (a *API) GetStates() ([]State, error) {
	resp, err := a.client.Get("/api/states")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get states: %s", resp.Status)
	}

	var states []State
	if err := json.NewDecoder(resp.Body).Decode(&states); err != nil {
		return nil, err
	}

	return states, nil
}

// GetState returns the state of a specific entity
func (a *API) GetState(entityID string) (*State, error) {
	resp, err := a.client.Get(fmt.Sprintf("/api/states/%s", entityID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get state for %s: %s", entityID, resp.Status)
	}

	var state State
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, err
	}

	return &state, nil
}

// SetState updates the state of an entity
func (a *API) SetState(entityID, state string, attributes map[string]interface{}) (*State, error) {
	body := map[string]interface{}{
		"state":      state,
		"attributes": attributes,
	}

	resp, err := a.client.Post(fmt.Sprintf("/api/states/%s", entityID), body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to set state for %s: %s", entityID, resp.Status)
	}

	var result State
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CallService calls a Home Assistant service
func (a *API) CallService(domain, service string, data map[string]interface{}) error {
	resp, err := a.client.Post(fmt.Sprintf("/api/services/%s/%s", domain, service), data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to call service %s.%s: %s", domain, service, resp.Status)
	}

	return nil
}

// GetConfig returns the current configuration of Home Assistant
func (a *API) GetConfig() (map[string]interface{}, error) {
	resp, err := a.client.Get("/api/config")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get config: %s", resp.Status)
	}

	var config map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

// GetServices returns all available services
func (a *API) GetServices() (map[string]map[string]interface{}, error) {
	resp, err := a.client.Get("/api/services")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get services: %s", resp.Status)
	}

	// 先尝试解析为数组格式（新版Home Assistant API）
	var servicesArray []map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if err := json.Unmarshal(body, &servicesArray); err == nil {
		// 成功解析为数组，将其转换为需要的map格式
		result := make(map[string]map[string]interface{})
		for _, svc := range servicesArray {
			domain, ok := svc["domain"].(string)
			if !ok {
				continue
			}

			// 如果域名不存在，初始化
			if _, exists := result[domain]; !exists {
				result[domain] = make(map[string]interface{})
			}

			// 如果有服务字段，添加该服务
			if services, ok := svc["services"].(map[string]interface{}); ok {
				for serviceName, serviceData := range services {
					result[domain][serviceName] = serviceData
				}
			}
		}
		return result, nil
	}

	// 如果不是数组格式，回退到尝试直接解析为map格式（旧版API）
	var services map[string]map[string]interface{}
	if err := json.Unmarshal(body, &services); err != nil {
		return nil, fmt.Errorf("failed to decode services: %v", err)
	}

	return services, nil
}

// GetEvents returns all available event types
func (a *API) GetEvents() ([]string, error) {
	resp, err := a.client.Get("/api/events")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get events: %s", resp.Status)
	}

	var events []string
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}

	return events, nil
}

// FireEvent fires an event with the given event_type and data
func (a *API) FireEvent(eventType string, data map[string]interface{}) error {
	resp, err := a.client.Post(fmt.Sprintf("/api/events/%s", eventType), data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fire event %s: %s", eventType, resp.Status)
	}

	return nil
}

// GetErrorLog returns the Home Assistant error log
func (a *API) GetErrorLog() (string, error) {
	resp, err := a.client.Get("/api/error_log")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get error log: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// CheckAPI tests if the API is up and running
func (a *API) CheckAPI() error {
	resp, err := a.client.Get("/api/")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API check failed: %s", resp.Status)
	}

	return nil
}
