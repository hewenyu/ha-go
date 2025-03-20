package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	hago "github.com/hewenyu/ha-go"
)

// 集成信息结构
type IntegrationInfo struct {
	Name         string
	Domain       string
	ServiceCount int
	Services     map[string][]string // 服务名称 -> 可用参数列表
	EntityCount  int
	Entities     []hago.State
}

func main() {
	// Replace with your Home Assistant URL and API token
	baseURL := "http://192.168.199.130:8123"
	apiToken := os.Getenv("HATOKEN")
	if apiToken == "" {
		fmt.Println("请设置环境变量HATOKEN为您的Home Assistant长期访问令牌")
		fmt.Println("例如: export HATOKEN=eyJhbGc...（您的长期访问令牌）")
		os.Exit(1)
	}

	// 创建Home Assistant客户端
	client, err := hago.NewClient(baseURL, apiToken)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 创建API实例
	api := hago.NewAPI(client)

	// 检查API是否可用
	if err := api.CheckAPI(); err != nil {
		log.Fatalf("API不可用: %v", err)
	}
	fmt.Println("Home Assistant API 连接成功!")

	// 获取配置信息
	config, err := api.GetConfig()
	if err != nil {
		log.Fatalf("获取配置失败: %v", err)
	}
	fmt.Printf("Home Assistant 版本: %v\n", config["version"])

	// 获取所有实体状态
	states, err := api.GetStates()
	if err != nil {
		log.Fatalf("获取实体状态失败: %v", err)
	}
	fmt.Printf("找到 %d 个实体\n", len(states))

	// 获取可用的服务
	services, err := api.GetServices()
	if err != nil {
		log.Fatalf("获取服务列表失败: %v", err)
	}

	// 收集集成信息
	fmt.Println("\n正在收集集成信息...")
	integrations := collectIntegrationInfo(states, services)

	// 显示集成统计信息
	fmt.Printf("\n====== Home Assistant 集成应用统计 ======\n")
	fmt.Printf("找到 %d 个集成应用\n", len(integrations))

	// 按集成域名排序
	var integrationNames []string
	for name := range integrations {
		integrationNames = append(integrationNames, name)
	}
	sort.Strings(integrationNames)

	// 显示集成信息
	fmt.Println("\n集成应用列表:")
	for i, name := range integrationNames {
		info := integrations[name]
		fmt.Printf("%3d. %-30s: %3d个服务, %3d个实体\n", i+1, name, info.ServiceCount, info.EntityCount)
	}

	// 交互式操作菜单
	interactiveMenu(api, integrations, integrationNames)
}

// 收集集成信息
func collectIntegrationInfo(states []hago.State, services map[string]map[string]interface{}) map[string]IntegrationInfo {
	integrations := make(map[string]IntegrationInfo)

	// 根据服务域名创建集成列表
	for domain, domainServices := range services {
		serviceNames := make(map[string][]string)
		for serviceName, serviceInfo := range domainServices {
			var params []string
			// 尝试提取服务参数
			if fields, ok := serviceInfo.(map[string]interface{}); ok {
				// 首先尝试fields直接包含的字段参数
				if fieldsData, ok := fields["fields"].(map[string]interface{}); ok {
					for field := range fieldsData {
						params = append(params, field)
					}
				}

				// 如果没有fields字段，尝试target和fields_schema结构
				if target, ok := fields["target"].(map[string]interface{}); ok {
					if entity, ok := target["entity"].(map[string]interface{}); ok {
						if _, ok := entity["domain"].(string); ok {
							params = append(params, "entity_id")
						}
					}
				}

				// 检查fields_schema字段
				if schema, ok := fields["fields_schema"].(map[string]interface{}); ok {
					for field := range schema {
						// 避免参数重复
						isDuplicate := false
						for _, p := range params {
							if p == field {
								isDuplicate = true
								break
							}
						}
						if !isDuplicate {
							params = append(params, field)
						}
					}
				}

				// 检查field参数
				if fieldParams, ok := fields["field_params"].(map[string]interface{}); ok {
					for field := range fieldParams {
						isDuplicate := false
						for _, p := range params {
							if p == field {
								isDuplicate = true
								break
							}
						}
						if !isDuplicate {
							params = append(params, field)
						}
					}
				}
			}

			// 如果还是没有找到参数，添加一些常见的
			if len(params) == 0 && (strings.Contains(serviceName, "turn_on") ||
				strings.Contains(serviceName, "turn_off") ||
				strings.Contains(serviceName, "toggle")) {
				params = append(params, "entity_id")
			}

			serviceNames[serviceName] = params
		}

		integrations[domain] = IntegrationInfo{
			Name:         domain,
			Domain:       domain,
			ServiceCount: len(domainServices),
			Services:     serviceNames,
			EntityCount:  0,
			Entities:     []hago.State{},
		}
	}

	// 按照实体的域名归类到对应的集成
	for _, state := range states {
		parts := strings.SplitN(state.EntityID, ".", 2)
		if len(parts) == 2 {
			domain := parts[0]
			info, ok := integrations[domain]
			if !ok {
				// 如果这个域名没有对应的服务，创建一个新的集成条目
				integrations[domain] = IntegrationInfo{
					Name:         domain,
					Domain:       domain,
					ServiceCount: 0,
					Services:     make(map[string][]string),
					EntityCount:  1,
					Entities:     []hago.State{state},
				}
			} else {
				info.EntityCount++
				info.Entities = append(info.Entities, state)
				integrations[domain] = info
			}
		}
	}

	return integrations
}

// 交互式菜单
func interactiveMenu(api *hago.API, integrations map[string]IntegrationInfo, integrationNames []string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n====== Home Assistant 集成应用管理 ======")
		fmt.Println("1. 查看集成详细信息")
		fmt.Println("2. 调用集成服务")
		fmt.Println("3. 查询实体状态")
		fmt.Println("4. 退出")
		fmt.Print("\n请选择操作 (1-4): ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			viewIntegrationDetails(reader, integrations, integrationNames)
		case "2":
			callIntegrationService(reader, api, integrations, integrationNames)
		case "3":
			queryEntityState(reader, api, integrations)
		case "4":
			fmt.Println("程序退出")
			return
		default:
			fmt.Println("无效选择，请重试")
		}
	}
}

// 查看集成详细信息
func viewIntegrationDetails(reader *bufio.Reader, integrations map[string]IntegrationInfo, integrationNames []string) {
	fmt.Println("\n====== 查看集成详细信息 ======")

	// 显示集成列表
	for i, name := range integrationNames {
		fmt.Printf("%3d. %s\n", i+1, name)
	}

	fmt.Print("\n请选择集成编号 (1-" + strconv.Itoa(len(integrationNames)) + "): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(integrationNames) {
		fmt.Println("无效选择")
		return
	}

	integrationName := integrationNames[idx-1]
	info := integrations[integrationName]

	fmt.Printf("\n集成: %s\n", integrationName)
	fmt.Printf("服务数量: %d\n", info.ServiceCount)

	if info.ServiceCount > 0 {
		fmt.Println("可用服务:")
		i := 1
		for service, params := range info.Services {
			fmt.Printf("  %d. %s.%s\n", i, integrationName, service)
			if len(params) > 0 {
				fmt.Printf("     参数: %s\n", strings.Join(params, ", "))
			}
			i++
		}
	}

	fmt.Printf("\n实体数量: %d\n", info.EntityCount)
	if info.EntityCount > 0 {
		fmt.Println("实体示例:")

		// 最多显示10个实体
		maxShow := 10
		if info.EntityCount < maxShow {
			maxShow = info.EntityCount
		}

		for i := 0; i < maxShow && i < len(info.Entities); i++ {
			entity := info.Entities[i]
			friendlyName := entity.EntityID
			if name, ok := entity.Attributes["friendly_name"].(string); ok && name != "" {
				friendlyName = name
			}
			fmt.Printf("  %d. %s (状态: %s)\n", i+1, entity.EntityID, entity.State)
			if friendlyName != entity.EntityID {
				fmt.Printf("     名称: %s\n", friendlyName)
			}
		}

		if info.EntityCount > maxShow {
			fmt.Printf("  (仅显示前%d个, 共%d个)\n", maxShow, info.EntityCount)
		}
	}
}

// 调用集成服务
func callIntegrationService(reader *bufio.Reader, api *hago.API, integrations map[string]IntegrationInfo, integrationNames []string) {
	fmt.Println("\n====== 调用集成服务 ======")

	// 仅显示有服务的集成
	var availableIntegrations []string
	var indexMap []string // 保存原始索引映射

	for _, name := range integrationNames {
		if integrations[name].ServiceCount > 0 {
			availableIntegrations = append(availableIntegrations, name)
			indexMap = append(indexMap, name)
		}
	}

	if len(availableIntegrations) == 0 {
		fmt.Println("没有可用的服务")
		return
	}

	// 显示可用集成
	for i, name := range availableIntegrations {
		fmt.Printf("%3d. %s (%d个服务)\n", i+1, name, integrations[name].ServiceCount)
	}

	fmt.Print("\n请选择集成编号: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(availableIntegrations) {
		fmt.Println("无效选择")
		return
	}

	integrationName := availableIntegrations[idx-1]
	info := integrations[integrationName]

	// 显示服务列表
	fmt.Printf("\n集成 '%s' 的可用服务:\n", integrationName)
	var serviceNames []string
	for service := range info.Services {
		serviceNames = append(serviceNames, service)
	}
	sort.Strings(serviceNames)

	for i, service := range serviceNames {
		fmt.Printf("%3d. %s.%s\n", i+1, integrationName, service)
		if params := info.Services[service]; len(params) > 0 {
			fmt.Printf("     参数: %s\n", strings.Join(params, ", "))
		}
	}

	fmt.Print("\n请选择服务编号: ")
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)

	serviceIdx, err := strconv.Atoi(input)
	if err != nil || serviceIdx < 1 || serviceIdx > len(serviceNames) {
		fmt.Println("无效选择")
		return
	}

	serviceName := serviceNames[serviceIdx-1]
	params := info.Services[serviceName]

	// 收集服务参数
	serviceData := make(map[string]interface{})

	if len(params) > 0 {
		fmt.Printf("\n服务 '%s.%s' 需要以下参数:\n", integrationName, serviceName)
		for _, param := range params {
			fmt.Printf("请输入 '%s' 的值 (留空跳过): ", param)
			input, _ = reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input != "" {
				// 尝试将输入值转换为适当的类型
				if input == "true" || input == "false" {
					serviceData[param] = (input == "true")
				} else if val, err := strconv.Atoi(input); err == nil {
					serviceData[param] = val
				} else if val, err := strconv.ParseFloat(input, 64); err == nil {
					serviceData[param] = val
				} else {
					serviceData[param] = input
				}
			}
		}
	}

	// 调用服务
	fmt.Printf("\n正在调用服务: %s.%s\n", integrationName, serviceName)
	fmt.Printf("参数: %v\n", serviceData)

	err = api.CallService(integrationName, serviceName, serviceData)
	if err != nil {
		fmt.Printf("调用服务失败: %v\n", err)
	} else {
		fmt.Println("服务调用成功!")
	}
}

// 查询实体状态
func queryEntityState(reader *bufio.Reader, api *hago.API, integrations map[string]IntegrationInfo) {
	fmt.Println("\n====== 查询实体状态 ======")
	fmt.Print("请输入实体ID (例如 light.living_room): ")

	input, _ := reader.ReadString('\n')
	entityID := strings.TrimSpace(input)

	if entityID == "" {
		fmt.Println("实体ID不能为空")
		return
	}

	state, err := api.GetState(entityID)
	if err != nil {
		fmt.Printf("获取状态失败: %v\n", err)
		return
	}

	fmt.Printf("\n实体: %s\n", state.EntityID)
	fmt.Printf("状态: %s\n", state.State)
	fmt.Printf("最后更改: %s\n", state.LastChanged.Format(time.RFC3339))
	fmt.Printf("最后更新: %s\n", state.LastUpdated.Format(time.RFC3339))

	if len(state.Attributes) > 0 {
		fmt.Println("属性:")

		// 按字母顺序排序属性键
		var attrKeys []string
		for k := range state.Attributes {
			attrKeys = append(attrKeys, k)
		}
		sort.Strings(attrKeys)

		// 打印属性
		for _, key := range attrKeys {
			fmt.Printf("  %s: %v\n", key, state.Attributes[key])
		}
	}

	// 提供控制选项
	parts := strings.SplitN(entityID, ".", 2)
	if len(parts) == 2 {
		domain := parts[0]
		if info, ok := integrations[domain]; ok && info.ServiceCount > 0 {
			fmt.Printf("\n此实体可能支持以下服务:\n")
			for service := range info.Services {
				fmt.Printf("  %s.%s\n", domain, service)
			}

			fmt.Print("\n是否要调用服务? (y/n): ")
			input, _ = reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if strings.ToLower(input) == "y" {
				// 使用相同的域选择一个服务
				fmt.Printf("\n请输入要调用的服务名称 (例如 %s.turn_on): ", domain)
				input, _ = reader.ReadString('\n')
				input = strings.TrimSpace(input)

				serviceParts := strings.SplitN(input, ".", 2)
				if len(serviceParts) != 2 {
					fmt.Println("无效的服务名称格式，应为 domain.service")
					return
				}

				domain, service := serviceParts[0], serviceParts[1]

				// 准备服务数据
				serviceData := map[string]interface{}{
					"entity_id": entityID,
				}

				// 调用服务
				fmt.Printf("调用服务: %s.%s 用于实体 %s\n", domain, service, entityID)
				err = api.CallService(domain, service, serviceData)
				if err != nil {
					fmt.Printf("调用服务失败: %v\n", err)
				} else {
					fmt.Println("服务调用成功!")

					// 再次获取状态以显示更改
					time.Sleep(1 * time.Second) // 给Home Assistant一点时间来处理
					newState, err := api.GetState(entityID)
					if err == nil && newState.State != state.State {
						fmt.Printf("实体状态已从 '%s' 变为 '%s'\n", state.State, newState.State)
					}
				}
			}
		}
	}
}
