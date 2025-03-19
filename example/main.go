package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	hago "github.com/hewenyu/ha-go"
)

func main() {
	// Replace with your Home Assistant URL and API token
	baseURL := "http://192.168.199.130:8123/"
	apiToken := os.Getenv("HATOKEN")
	// wsURL变量已移除，因为当前不使用WebSocket功能

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

	// 1. 先按照集成（Integration）分类
	integrationMap := make(map[string][]hago.State)

	// 分析每个设备的entity_id，尝试确定其集成源
	for _, state := range states {
		// 从entity_id中推断集成名称
		integrationName := getIntegrationNameFromEntity(state)
		integrationMap[integrationName] = append(integrationMap[integrationName], state)
	}

	// 获取所有集成名称并排序
	var integrationNames []string
	for name := range integrationMap {
		integrationNames = append(integrationNames, name)
	}
	sort.Strings(integrationNames)

	// 创建一个统计摘要
	fmt.Println("\n====== Home Assistant 集成与设备统计摘要 ======")
	fmt.Printf("共找到 %d 个设备，分布在 %d 个集成中\n", len(states), len(integrationNames))

	// 显示每个集成的设备数量
	fmt.Println("\n集成统计:")
	for _, name := range integrationNames {
		count := len(integrationMap[name])
		fmt.Printf("  %-25s: %3d个设备\n", name, count)
	}

	// 询问用户是否要查看详细信息
	fmt.Println("\n按回车键查看集成和设备详细信息，或按Ctrl+C退出...")
	// 模拟等待用户输入（实际使用时可以取消注释）
	// fmt.Scanln()

	// 2. 输出每个集成的设备信息，并按设备类型进一步分类
	fmt.Println("\n====== 按集成列出所有设备 ======")

	// 首先处理重要的集成
	importantIntegrations := []string{"xiaomi", "tuya", "mqtt", "deye", "zhimi", "opple"}
	processedIntegrations := make(map[string]bool)

	// 首先显示重要的集成
	for _, name := range importantIntegrations {
		for _, integration := range integrationNames {
			if strings.Contains(strings.ToLower(integration), strings.ToLower(name)) {
				displayIntegrationDevices(integration, integrationMap[integration])
				processedIntegrations[integration] = true
			}
		}
	}

	// 然后显示其余的集成
	for _, name := range integrationNames {
		if !processedIntegrations[name] {
			displayIntegrationDevices(name, integrationMap[name])
		}
	}

	// 提供获取特定集成的详细信息示例
	fmt.Println("\n====== 获取特定集成下特定设备的详细信息示例 ======")

	// 尝试获取一个典型的集成，如果存在
	var exampleIntegration string
	var exampleEntity string

	// 尝试Xiaomi集成
	for _, name := range integrationNames {
		if strings.Contains(strings.ToLower(name), "xiaomi") && len(integrationMap[name]) > 0 {
			exampleIntegration = name
			exampleEntity = integrationMap[name][0].EntityID
			break
		}
	}

	// 如果没有找到Xiaomi设备，尝试其他集成
	if exampleEntity == "" {
		for _, name := range integrationNames {
			if len(integrationMap[name]) > 0 {
				exampleIntegration = name
				exampleEntity = integrationMap[name][0].EntityID
				break
			}
		}
	}

	if exampleEntity != "" {
		state, err := api.GetState(exampleEntity)
		if err != nil {
			log.Printf("Failed to get state for %s: %v", exampleEntity, err)
		} else {
			fmt.Printf("\n集成 '%s' 的设备详细信息 - %s:\n", exampleIntegration, exampleEntity)
			fmt.Printf("  状态: %s\n", state.State)
			fmt.Printf("  Context ID: %s\n", state.Context.ID)
			fmt.Printf("  最后更改: %s\n", state.LastChanged.Format("2006-01-02 15:04:05"))
			fmt.Printf("  最后更新: %s\n", state.LastUpdated.Format("2006-01-02 15:04:05"))

			if len(state.Attributes) > 0 {
				fmt.Println("  属性:")

				// 获取并排序属性键
				var attrKeys []string
				for k := range state.Attributes {
					attrKeys = append(attrKeys, k)
				}
				sort.Strings(attrKeys)

				for _, key := range attrKeys {
					fmt.Printf("    %s: %v\n", key, state.Attributes[key])
				}
			}
		}
	}

	fmt.Println("\n程序执行完毕。按Ctrl+C退出...")

	// 等待中断信号以优雅地关闭
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("正在关闭...")
}

// 从entity_id推断集成名称
func getIntegrationNameFromEntity(state hago.State) string {
	// 首先从属性中尝试获取集成信息
	if platform, ok := state.Attributes["platform"].(string); ok && platform != "" {
		return platform
	}

	// 从实体ID尝试推断
	parts := strings.SplitN(state.EntityID, ".", 2)
	if len(parts) < 2 {
		return "unknown"
	}

	domain := parts[0]
	entityName := parts[1]

	// 如果entityName包含特定前缀，可能表示特定集成
	if strings.HasPrefix(entityName, "xiaomi_") {
		return "Xiaomi"
	} else if strings.HasPrefix(entityName, "tuya_") {
		return "Tuya"
	} else if strings.HasPrefix(entityName, "deye_") {
		return "Deye"
	} else if strings.HasPrefix(entityName, "zhimi_") {
		return "Zhimi"
	} else if strings.HasPrefix(entityName, "opple_") {
		return "Opple"
	} else if strings.HasPrefix(entityName, "fawad_") {
		return "Fawad"
	} else if strings.HasPrefix(entityName, "090615_") {
		return "PTXZN"
	} else if strings.HasPrefix(entityName, "miaomiaoc_") {
		return "Mijia"
	}

	// 尝试从实体名称找到品牌或制造商信息
	lowercaseEntityName := strings.ToLower(entityName)
	if strings.Contains(lowercaseEntityName, "xiaomi") {
		return "Xiaomi"
	} else if strings.Contains(lowercaseEntityName, "mi_") || strings.Contains(lowercaseEntityName, "_mi") {
		return "Xiaomi"
	} else if strings.Contains(lowercaseEntityName, "tuya") {
		return "Tuya"
	} else if strings.Contains(lowercaseEntityName, "mqtt") {
		return "MQTT"
	} else if strings.Contains(lowercaseEntityName, "google") {
		return "Google"
	}

	// 如果有一个友好的名称，尝试从中提取信息
	if friendlyName, ok := state.Attributes["friendly_name"].(string); ok && friendlyName != "" {
		lowercaseFriendlyName := strings.ToLower(friendlyName)
		if strings.Contains(lowercaseFriendlyName, "xiaomi") {
			return "Xiaomi"
		} else if strings.Contains(lowercaseFriendlyName, "tuya") {
			return "Tuya"
		} else if strings.Contains(lowercaseFriendlyName, "deye") {
			return "Deye"
		} else if strings.Contains(lowercaseFriendlyName, "小米") || strings.Contains(lowercaseFriendlyName, "小爱") {
			return "Xiaomi"
		}
	}

	// 根据domain推断集成
	switch domain {
	case "light", "switch", "binary_sensor", "sensor", "button", "number", "climate":
		// 这些是通用类型，使用vendor推断
		if vendor, ok := state.Attributes["vendor"].(string); ok && vendor != "" {
			return vendor
		}
		if manufacturer, ok := state.Attributes["manufacturer"].(string); ok && manufacturer != "" {
			return manufacturer
		}
		// 没有厂商信息，使用domain作为后备
		return "Home Assistant " + domain
	case "automation", "script", "scene":
		return "Home Assistant Automation"
	case "weather":
		return "Weather"
	case "sun":
		return "Sun"
	case "person":
		return "Person"
	case "zone":
		return "Zone"
	case "media_player":
		return "Media"
	case "camera":
		return "Camera"
	case "tts":
		return "Text-to-Speech"
	case "update":
		return "Updates"
	default:
		// 使用domain作为后备
		return "Home Assistant " + domain
	}
}

// 显示集成下的设备信息
func displayIntegrationDevices(integration string, states []hago.State) {
	fmt.Printf("\n## %s (%d个设备)\n", integration, len(states))

	// 先按设备类型进一步分类
	domainMap := make(map[string][]hago.State)
	for _, state := range states {
		parts := strings.SplitN(state.EntityID, ".", 2)
		if len(parts) == 2 {
			domain := parts[0]
			domainMap[domain] = append(domainMap[domain], state)
		}
	}

	// 获取所有域并按字母顺序排序
	var domains []string
	for domain := range domainMap {
		domains = append(domains, domain)
	}
	sort.Strings(domains)

	// 显示每个类型的设备
	for _, domain := range domains {
		fmt.Printf("\n  🔶 %s 类型设备 (%d个):\n", domain, len(domainMap[domain]))

		// 对设备按名称排序
		sort.Slice(domainMap[domain], func(i, j int) bool {
			return domainMap[domain][i].EntityID < domainMap[domain][j].EntityID
		})

		for _, state := range domainMap[domain] {
			// 获取友好名称
			friendlyName := ""
			if name, ok := state.Attributes["friendly_name"].(string); ok {
				friendlyName = name
			}

			// 获取设备状态和单位
			stateValue := state.State
			unit := ""
			if u, ok := state.Attributes["unit_of_measurement"].(string); ok {
				unit = u
			}

			// 显示基本信息
			if friendlyName != "" {
				fmt.Printf("\n    🔹 %s (%s)\n", state.EntityID, friendlyName)
			} else {
				fmt.Printf("\n    🔹 %s\n", state.EntityID)
			}

			if unit != "" {
				fmt.Printf("      状态: %s %s\n", stateValue, unit)
			} else {
				fmt.Printf("      状态: %s\n", stateValue)
			}

			fmt.Printf("      最后更新: %s\n", state.LastUpdated.Format("2006-01-02 15:04:05"))

			// 显示主要属性（过滤掉一些不太重要的）
			importantAttrs := []string{"device_class", "state_class", "icon", "supported_features"}
			hasDisplayedAttrs := false

			for _, key := range importantAttrs {
				if value, ok := state.Attributes[key]; ok {
					if !hasDisplayedAttrs {
						fmt.Println("      主要属性:")
						hasDisplayedAttrs = true
					}
					fmt.Printf("        %s: %v\n", key, value)
				}
			}
		}
	}
}
