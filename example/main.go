package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

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

	// 创建一个按域(domain)分类的实体映射
	domainMap := make(map[string][]hago.State)

	// 将实体按域分类
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

	// 创建一个统计摘要
	fmt.Println("\n====== Home Assistant 设备统计摘要 ======")
	fmt.Printf("共找到 %d 个设备，分布在 %d 种类型中\n", len(states), len(domains))

	// 显示每种类型的设备数量
	fmt.Println("\n设备类型统计:")
	for _, domain := range domains {
		count := len(domainMap[domain])
		fmt.Printf("  %-20s: %3d个\n", domain, count)
	}

	// 询问用户是否要查看详细信息
	fmt.Println("\n按回车键查看设备详细信息，或按Ctrl+C退出...")
	// 模拟等待用户输入（实际使用时可以取消注释）
	// fmt.Scanln()

	// 输出每个域的实体信息
	fmt.Println("\n====== 按类型列出所有设备详细信息 ======")

	// 只显示用户可能最关心的几种设备类型
	importantDomains := []string{"light", "switch", "sensor", "climate", "media_player", "camera"}

	// 首先显示重要的设备类型
	for _, domain := range importantDomains {
		if entityStates, ok := domainMap[domain]; ok {
			printDomainDevices(domain, entityStates)
			// 从domains删除已显示的domain
			for i, d := range domains {
				if d == domain {
					domains = append(domains[:i], domains[i+1:]...)
					break
				}
			}
		}
	}

	// 然后显示其余的设备类型
	for _, domain := range domains {
		printDomainDevices(domain, domainMap[domain])
	}

	// 提供获取特定实体的详细信息的示例
	fmt.Println("\n====== 获取特定设备详细信息示例 ======")

	// 尝试获取一个典型的实体，如果存在
	var exampleEntity string
	if len(domainMap["light"]) > 0 {
		exampleEntity = domainMap["light"][0].EntityID
	} else if len(domainMap["switch"]) > 0 {
		exampleEntity = domainMap["switch"][0].EntityID
	} else if len(states) > 0 {
		exampleEntity = states[0].EntityID
	}

	if exampleEntity != "" {
		state, err := api.GetState(exampleEntity)
		if err != nil {
			log.Printf("Failed to get state for %s: %v", exampleEntity, err)
		} else {
			fmt.Printf("\n详细信息 - %s:\n", exampleEntity)
			fmt.Printf("  状态: %s\n", state.State)
			fmt.Printf("  Context ID: %s\n", state.Context.ID)
			fmt.Printf("  最后更改: %s\n", state.LastChanged.Format(time.RFC3339))
			fmt.Printf("  最后更新: %s\n", state.LastUpdated.Format(time.RFC3339))

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

	// 不启用WebSocket示例，因为在运行中遇到连接问题
	fmt.Println("\n====== WebSocket示例已禁用 ======")
	fmt.Println("WebSocket连接在当前环境中可能无法正常工作。")
	fmt.Println("要使用WebSocket功能，请确保:")
	fmt.Println("1. Home Assistant实例可以通过WebSocket访问")
	fmt.Println("2. 使用正确的URL格式(ws://或wss://)")
	fmt.Println("3. 使用有效的API令牌")

	fmt.Println("\n程序执行完毕。按Ctrl+C退出...")

	// 等待中断信号以优雅地关闭
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("正在关闭...")
}

// 打印域内的设备信息
func printDomainDevices(domain string, states []hago.State) {
	fmt.Printf("\n## %s (%d个设备)\n", domain, len(states))

	// 对设备按名称排序
	sort.Slice(states, func(i, j int) bool {
		return states[i].EntityID < states[j].EntityID
	})

	for _, state := range states {
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
			fmt.Printf("\n  🔹 %s (%s)\n", state.EntityID, friendlyName)
		} else {
			fmt.Printf("\n  🔹 %s\n", state.EntityID)
		}

		if unit != "" {
			fmt.Printf("    状态: %s %s\n", stateValue, unit)
		} else {
			fmt.Printf("    状态: %s\n", stateValue)
		}

		fmt.Printf("    最后更新: %s\n", state.LastUpdated.Format("2006-01-02 15:04:05"))

		// 显示主要属性（过滤掉一些不太重要的）
		importantAttrs := []string{"device_class", "state_class", "icon", "supported_features"}
		hasDisplayedAttrs := false

		for _, key := range importantAttrs {
			if value, ok := state.Attributes[key]; ok {
				if !hasDisplayedAttrs {
					fmt.Println("    主要属性:")
					hasDisplayedAttrs = true
				}
				fmt.Printf("      %s: %v\n", key, value)
			}
		}

		// 如果需要查看所有属性，可以取消下面的注释
		/*
			if len(state.Attributes) > 0 {
				fmt.Println("    所有属性:")

				var attrKeys []string
				for k := range state.Attributes {
					attrKeys = append(attrKeys, k)
				}
				sort.Strings(attrKeys)

				for _, key := range attrKeys {
					fmt.Printf("      %s: %v\n", key, state.Attributes[key])
				}
			}
		*/
	}
}
