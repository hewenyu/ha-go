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

	// 1. 按照"集成-设备-实体"的顺序组织数据

	// 首先，创建一个集成 -> 设备 -> 实体的映射
	integrationDeviceMap := make(map[string]map[string][]hago.State)

	// 按照集成和设备ID对实体进行分组
	for _, state := range states {
		// 获取集成名称
		integrationName := getIntegrationNameFromEntity(state)

		// 获取设备ID（从实体ID或属性中推断）
		deviceID := getDeviceIDFromEntity(state)

		// 确保集成映射存在
		if _, ok := integrationDeviceMap[integrationName]; !ok {
			integrationDeviceMap[integrationName] = make(map[string][]hago.State)
		}

		// 将实体添加到对应的设备下
		integrationDeviceMap[integrationName][deviceID] = append(integrationDeviceMap[integrationName][deviceID], state)
	}

	// 获取所有集成名称并排序
	var integrationNames []string
	for name := range integrationDeviceMap {
		integrationNames = append(integrationNames, name)
	}
	sort.Strings(integrationNames)

	// 创建一个统计摘要
	fmt.Println("\n====== Home Assistant 集成、设备与实体统计摘要 ======")

	// 计算总设备数
	totalDevices := 0
	for _, deviceMap := range integrationDeviceMap {
		totalDevices += len(deviceMap)
	}

	fmt.Printf("共找到 %d 个实体，分布在 %d 个设备上，来自 %d 个集成\n",
		len(states), totalDevices, len(integrationNames))

	// 显示每个集成的设备和实体数量
	fmt.Println("\n集成统计:")
	for _, name := range integrationNames {
		deviceCount := len(integrationDeviceMap[name])

		// 计算该集成下的实体总数
		entityCount := 0
		for _, entities := range integrationDeviceMap[name] {
			entityCount += len(entities)
		}

		fmt.Printf("  %-25s: %3d个设备, %3d个实体\n", name, deviceCount, entityCount)
	}

	// 询问用户是否要查看详细信息
	fmt.Println("\n按回车键查看详细信息，或按Ctrl+C退出...")
	// 模拟等待用户输入（实际使用时可以取消注释）
	// fmt.Scanln()

	// 2. 以"集成-设备-实体"的层次结构输出详细信息
	fmt.Println("\n====== 按集成-设备-实体层次显示所有设备 ======")

	// 首先展示重要的集成
	importantIntegrations := []string{"xiaomi", "tuya", "mqtt", "deye", "zhimi", "opple"}
	processedIntegrations := make(map[string]bool)

	// 首先显示重要的集成
	for _, name := range importantIntegrations {
		for _, integration := range integrationNames {
			if strings.Contains(strings.ToLower(integration), strings.ToLower(name)) && !processedIntegrations[integration] {
				displayIntegrationDetails(integration, integrationDeviceMap[integration])
				processedIntegrations[integration] = true
			}
		}
	}

	// 然后显示其余的集成
	for _, integration := range integrationNames {
		if !processedIntegrations[integration] {
			displayIntegrationDetails(integration, integrationDeviceMap[integration])
		}
	}

	fmt.Println("\n程序执行完毕。按Ctrl+C退出...")

	// 等待中断信号以优雅地关闭
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("正在关闭...")
}

// 从实体推断设备ID
func getDeviceIDFromEntity(state hago.State) string {
	// 首先检查属性中是否有设备ID
	if deviceID, ok := state.Attributes["device_id"].(string); ok && deviceID != "" {
		return deviceID
	}

	// 如果没有直接的设备ID，则从实体ID中推断
	parts := strings.SplitN(state.EntityID, ".", 2)
	if len(parts) < 2 {
		return "未知设备"
	}

	domain := parts[0]
	entityID := parts[1]

	// 对于某些集成，可以通过特定前缀或模式识别设备
	// 例如：xiaomi_cn_540583230_eaffh1_state_playing_p_7_1
	// 可以提取出 xiaomi_cn_540583230_eaffh1 作为设备ID

	// 提取可能的设备标识符
	deviceIdentifier := ""

	// 先检查友好名称中的设备名称
	if friendlyName, ok := state.Attributes["friendly_name"].(string); ok && friendlyName != "" {
		// 如果友好名称中包含设备名称与属性名称的分隔，提取设备部分
		nameParts := strings.Split(friendlyName, "  ")
		if len(nameParts) > 1 {
			deviceIdentifier = nameParts[0]
			return deviceIdentifier
		}
	}

	// 常见的设备ID模式
	patterns := []struct {
		prefix string
		parts  int
	}{
		{"xiaomi_cn_", 4},    // xiaomi_cn_<id>_<model>
		{"zhimi_cn_", 4},     // zhimi_cn_<id>_<model>
		{"opple_cn_", 4},     // opple_cn_<id>_<model>
		{"deye_", 1},         // deye_<id>
		{"miaomiaoc_cn_", 5}, // miaomiaoc_cn_<type>_<id>_<model>
		{"090615_cn_", 4},    // 090615_cn_<id>_<model>
		{"fawad_cn_", 4},     // fawad_cn_<id>_<model>
	}

	for _, pattern := range patterns {
		if strings.HasPrefix(entityID, pattern.prefix) {
			idParts := strings.SplitN(entityID, "_", pattern.parts+1)
			if len(idParts) > pattern.parts {
				// 构建设备ID前缀
				devicePrefix := strings.Join(idParts[:pattern.parts], "_")

				// 如果entityID中包含设备类型或其他信息，则继续细分
				if extraParts := strings.SplitN(idParts[pattern.parts], "_p_", 2); len(extraParts) > 1 {
					deviceIdentifier = devicePrefix + "_" + extraParts[0]
				} else {
					deviceIdentifier = devicePrefix
				}

				return deviceIdentifier
			}
		}
	}

	// 对于特殊设备，使用特定识别方法
	if domain == "climate" && strings.HasPrefix(entityID, "090615_cn_proxy") {
		parts := strings.Split(entityID, "_")
		if len(parts) >= 5 {
			// 提取代理后面的数字部分作为设备标识符
			return "空调" + parts[4]
		}
	}

	// 如果无法识别特定模式，使用最简单的方法：使用整个实体ID或其一部分
	// 优先提取实体ID中可能的设备部分
	if deviceParts := strings.SplitN(entityID, "_p_", 2); len(deviceParts) > 1 {
		return domain + "." + deviceParts[0]
	}

	// 最后的后备方案：使用实体域和前缀作为设备ID
	// 将同类型且前缀相似的实体分到一起
	for _, prefix := range []string{"xiaomi", "tuya", "deye", "zhimi", "opple", "fawad", "090615"} {
		if strings.Contains(entityID, prefix) {
			return domain + "." + prefix
		}
	}

	// 如果都没有匹配，使用域名作为设备ID
	return domain
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

// 显示集成详细信息，按设备和实体层次
func displayIntegrationDetails(integration string, deviceMap map[string][]hago.State) {
	fmt.Printf("\n## 集成: %s (%d个设备)\n", integration, len(deviceMap))

	// 获取所有设备ID并排序
	var deviceIDs []string
	for id := range deviceMap {
		deviceIDs = append(deviceIDs, id)
	}
	sort.Strings(deviceIDs)

	// 显示每个设备及其实体
	for _, deviceID := range deviceIDs {
		entities := deviceMap[deviceID]

		// 提取设备友好名称
		deviceName := getDeviceFriendlyName(entities)

		if deviceName != "" && deviceName != deviceID {
			fmt.Printf("\n  🔶 设备: %s (%s) - %d个实体\n", deviceID, deviceName, len(entities))
		} else {
			fmt.Printf("\n  🔶 设备: %s - %d个实体\n", deviceID, len(entities))
		}

		// 将实体按域名分类
		domainMap := make(map[string][]hago.State)
		for _, entity := range entities {
			parts := strings.SplitN(entity.EntityID, ".", 2)
			if len(parts) == 2 {
				domain := parts[0]
				domainMap[domain] = append(domainMap[domain], entity)
			}
		}

		// 获取所有域名并排序
		var domains []string
		for domain := range domainMap {
			domains = append(domains, domain)
		}
		sort.Strings(domains)

		// 优先显示重要的域
		importantDomains := []string{"light", "switch", "climate", "sensor", "binary_sensor", "media_player"}
		processedDomains := make(map[string]bool)

		// 首先显示重要的域
		for _, domain := range importantDomains {
			if domainEntities, ok := domainMap[domain]; ok {
				displayEntityGroup(domain, domainEntities)
				processedDomains[domain] = true
			}
		}

		// 然后显示其余的域
		for _, domain := range domains {
			if !processedDomains[domain] {
				displayEntityGroup(domain, domainMap[domain])
			}
		}
	}
}

// 从一组实体中提取设备友好名称
func getDeviceFriendlyName(entities []hago.State) string {
	// 首先尝试从实体的friendly_name属性中提取
	for _, entity := range entities {
		if friendlyName, ok := entity.Attributes["friendly_name"].(string); ok && friendlyName != "" {
			// 如果friendly_name包含分隔符，取第一部分作为设备名称
			parts := strings.Split(friendlyName, "  ")
			if len(parts) > 0 {
				return parts[0]
			}
			return friendlyName
		}
	}

	// 如果没有找到合适的friendly_name，返回空字符串
	return ""
}

// 显示同一域的实体组
func displayEntityGroup(domain string, entities []hago.State) {
	fmt.Printf("\n    🔷 %s类型实体 (%d个):\n", domain, len(entities))

	// 对实体按ID排序
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].EntityID < entities[j].EntityID
	})

	// 显示每个实体
	for _, entity := range entities {
		// 获取实体友好名称
		friendlyName := ""
		if name, ok := entity.Attributes["friendly_name"].(string); ok {
			// 尝试从friendly_name中提取实体功能部分
			parts := strings.Split(name, "  ")
			if len(parts) > 1 {
				friendlyName = parts[1]
			} else {
				friendlyName = name
			}
		}

		// 获取状态和单位
		stateValue := entity.State
		unit := ""
		if u, ok := entity.Attributes["unit_of_measurement"].(string); ok {
			unit = u
		}

		// 显示基本信息
		if friendlyName != "" {
			fmt.Printf("\n      🔹 %s (%s)\n", entity.EntityID, friendlyName)
		} else {
			fmt.Printf("\n      🔹 %s\n", entity.EntityID)
		}

		if unit != "" {
			fmt.Printf("        状态: %s %s\n", stateValue, unit)
		} else {
			fmt.Printf("        状态: %s\n", stateValue)
		}

		// 显示主要属性
		importantAttrs := []string{"device_class", "state_class", "icon", "supported_features"}
		hasDisplayedAttrs := false

		for _, key := range importantAttrs {
			if value, ok := entity.Attributes[key]; ok {
				if !hasDisplayedAttrs {
					fmt.Println("        主要属性:")
					hasDisplayedAttrs = true
				}
				fmt.Printf("          %s: %v\n", key, value)
			}
		}
	}
}
