package missinglist

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
)

// LuaConfig 解析后的Lua配置
type LuaConfig struct {
	GameType         string                 `json:"game_type"`          // bydr, dsc, cqsj, yxds
	Titles           map[string]string      `json:"titles"`             // 榜单类型 -> 标题
	Contents         map[string]string      `json:"contents"`           // 榜单类型 -> 内容
	Rewards          map[string]interface{} `json:"rewards"`            // 榜单类型 -> 奖励配置
	RewardsWithNames map[string]interface{} `json:"rewards_with_names"` // 榜单类型 -> 奖励配置（包含道具名称）
	SuperLimits      map[string]SuperLimit  `json:"super_limits"`       // 超级奖限制
}

// SuperLimit 超级奖限制条件
type SuperLimit struct {
	MaxRank  int `json:"max_rank"`
	MinScore int `json:"min_score"`
}

// ParseLuaConfig 解析Lua配置文件
func ParseLuaConfig(luaContent string, gameType string) (*LuaConfig, error) {
	config := &LuaConfig{
		GameType:         gameType,
		Titles:           make(map[string]string),
		Contents:         make(map[string]string),
		Rewards:          make(map[string]interface{}),
		RewardsWithNames: make(map[string]interface{}),
		SuperLimits:      make(map[string]SuperLimit),
	}

	// 根据游戏类型选择不同的解析策略
	var err error
	switch gameType {
	case "bydr":
		config, err = parseBydrConfig(luaContent, config)
	case "dsc":
		config, err = parseDscConfig(luaContent, config)
	case "cqsj":
		config, err = parseCqsjConfig(luaContent, config)
	case "yxds":
		config, err = parseYxdsConfig(luaContent, config)
	case "daqiqiu":
		config, err = parseCqsjConfig(luaContent, config)
	default:
		return nil, fmt.Errorf("不支持的游戏类型: %s", gameType)
	}

	if err != nil {
		return nil, err
	}

	// 转换奖励配置，添加道具名称
	convertRewardsWithNames(config)

	return config, nil
}

// convertRewardsWithNames 将奖励配置中的道具ID转换为道具名称
func convertRewardsWithNames(config *LuaConfig) {
	// 加载道具映射
	itemMap, err := LoadItemMap()
	if err != nil {
		// 如果加载失败，RewardsWithNames使用原始Rewards
		config.RewardsWithNames = config.Rewards
		return
	}

	// 遍历所有榜单类型的奖励配置
	for rankType, reward := range config.Rewards {
		if rewardMap, ok := reward.(map[string]map[string]int); ok {
			// 创建包含名称的奖励配置
			rewardWithNames := make(map[string]map[string]interface{})

			for rankLimit, items := range rewardMap {
				itemsWithNames := make(map[string]interface{})
				for itemId, count := range items {
					itemName := itemMap[itemId]
					if itemName == "" {
						itemName = itemId
					}
					// 存储格式：{"道具ID": {"id": "道具ID", "name": "道具名称", "count": 数量}}
					itemsWithNames[itemId] = map[string]interface{}{
						"id":    itemId,
						"name":  itemName,
						"count": count,
					}
				}
				rewardWithNames[rankLimit] = itemsWithNames
			}

			config.RewardsWithNames[rankType] = rewardWithNames
		} else {
			// 如果不是预期的格式，直接复制
			config.RewardsWithNames[rankType] = reward
		}
	}
}

// parseBydrConfig 解析bydr类型的Lua配置
func parseBydrConfig(luaContent string, config *LuaConfig) (*LuaConfig, error) {
	// 解析邮件标题
	titlePatterns := map[string]string{
		"daily_person":       `mailtitle_daily_zd\s*=\s*['"]([^'"]+)['"]`,
		"total_person":       `mailtitle_total_zd\s*=\s*['"]([^'"]+)['"]`,
		"daily_person_super": `mailtitle_daily_super\s*=\s*['"]([^'"]+)['"]`,
		"total_person_super": `mailtitle_total_super\s*=\s*['"]([^'"]+)['"]`,
		"daily_club":         `mailtitle_daily_gh\s*=\s*['"]([^'"]+)['"]`,
		"total_club":         `mailtitle_total_gh\s*=\s*['"]([^'"]+)['"]`,
	}

	for rankType, pattern := range titlePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(luaContent)
		if len(matches) > 1 {
			config.Titles[rankType] = matches[1]
		}
	}

	// 解析邮件内容
	contentPatterns := map[string]string{
		"daily_person":       `mailcontent_daily_zd\s*=\s*['"]([^'"]+)['"]`,
		"total_person":       `mailcontent_total_zd\s*=\s*['"]([^'"]+)['"]`,
		"daily_person_super": `mailcontent_daily_super\s*=\s*['"]([^'"]+)['"]`,
		"total_person_super": `mailcontent_total_super\s*=\s*['"]([^'"]+)['"]`,
		"daily_club":         `mailcontent_daily_gh\s*=\s*['"]([^'"]+)['"]`,
		"total_club":         `mailcontent_total_gh\s*=\s*['"]([^'"]+)['"]`,
	}

	for rankType, pattern := range contentPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(luaContent)
		if len(matches) > 1 {
			config.Contents[rankType] = matches[1]
		}
	}

	// 解析奖励配置（使用多行匹配，因为奖励表可能跨多行）
	rewardPatterns := map[string]string{
		"daily_person":       `daily_reward_zd\s*=\s*\{`,
		"total_person":       `total_reward_zd\s*=\s*\{`,
		"daily_person_super": `day_super_reward_zd\s*=\s*\{`,
		"total_person_super": `total_super_reward_zd\s*=\s*\{`,
		"daily_club":         `daily_reward_gh\s*=\s*\{`,
		"total_club":         `total_reward_gh\s*=\s*\{`,
	}

	for rankType, startPattern := range rewardPatterns {
		tableContent, ok := extractTableByPattern(luaContent, startPattern)
		if !ok {
			// 调试：记录未找到的奖励配置
			if rankType == "daily_person_super" || rankType == "total_person_super" {
				log.Printf("警告: 未找到 %s 的奖励配置，模式: %s", rankType, startPattern)
			}
			continue
		}
		rewardConfig, err := parseRewardTable(tableContent)
		if err != nil {
			// 调试：记录解析错误
			if rankType == "daily_person_super" || rankType == "total_person_super" {
				log.Printf("错误: 解析 %s 奖励配置失败: %v", rankType, err)
			}
			continue
		}
		if len(rewardConfig) > 0 {
			config.Rewards[rankType] = rewardConfig
			// 调试：记录成功解析的超级榜配置
			if rankType == "daily_person_super" || rankType == "total_person_super" {
				log.Printf("成功: 解析 %s 奖励配置，包含 %d 个排名", rankType, len(rewardConfig))
			}
		} else {
			// 调试：记录空配置
			if rankType == "daily_person_super" || rankType == "total_person_super" {
				log.Printf("警告: %s 奖励配置为空", rankType)
			}
		}
	}

	// 解析超级奖限制（排名/积分）
	superNumDay := extractNumber(luaContent, `super_num_day\s*=\s*(\d+)`)
	superLimitDay := extractNumber(luaContent, `super_limit_day\s*=\s*(\d+)`)
	if superNumDay > 0 || superLimitDay > 0 {
		config.SuperLimits["daily_person_super"] = SuperLimit{
			MaxRank:  superNumDay,
			MinScore: superLimitDay,
		}
	}

	superNumTotal := extractNumber(luaContent, `super_num_total\s*=\s*(\d+)`)
	superLimitTotal := extractNumber(luaContent, `super_limit_total\s*=\s*(\d+)`)
	if superNumTotal > 0 || superLimitTotal > 0 {
		config.SuperLimits["total_person_super"] = SuperLimit{
			MaxRank:  superNumTotal,
			MinScore: superLimitTotal,
		}
	}

	return config, nil
}

// parseDscConfig 解析dsc类型的Lua配置
func parseDscConfig(luaContent string, config *LuaConfig) (*LuaConfig, error) {
	// DSC的配置格式可能与bydr类似，但字段名可能不同
	// 这里先使用类似的解析逻辑，后续可以根据实际DSC配置文件调整
	return parseBydrConfig(luaContent, config)
}

// parseCqsjConfig 解析cqsj类型的Lua配置
func parseCqsjConfig(luaContent string, config *LuaConfig) (*LuaConfig, error) {
	// cqsj/daqiqiu 奖励表位于 page[1] 中
	pageTable, ok := extractTableByPattern(luaContent, `page\s*=\s*\{`)
	if !ok {
		pageTable = luaContent
	}

	// 取 page[1] 的配置块
	page1Content := extractBracketTable(pageTable, 1)
	if page1Content == "" {
		page1Content = pageTable
	}

	// 标题（page[1].title 或顶层 title）
	title := extractString(page1Content, `title\s*=\s*['"]([^'"]+)['"]`)
	if title == "" {
		title = extractString(luaContent, `title\s*=\s*['"]([^'"]+)['"]`)
	}
	if title != "" {
		config.Titles["daily_person"] = title + "个人日榜奖励"
		config.Titles["total_person"] = title + "个人总榜奖励"
		config.Titles["total_person_super"] = title + "超级奖奖励"
	}

	// 内容默认文案
	if config.Contents["daily_person"] == "" {
		config.Contents["daily_person"] = "尊敬的守护者你好，您在活动个人榜昨日的日榜排名中最终获得了第%s名，最终共计获得%s积分，以下为您获得的奖励，请领取！"
	}
	if config.Contents["total_person"] == "" {
		config.Contents["total_person"] = "尊敬的守护者你好，您在活动个人榜的总榜排名中最终获得了第%s名，最终共计获得%s积分，以下为您获得的奖励，请领取！"
	}
	if config.Contents["total_person_super"] == "" {
		config.Contents["total_person_super"] = "尊敬的守护者你好，您在活动中获得了超级奖第%s名，以下为您获得的奖励，请领取！"
	}

	rewardPatterns := map[string]string{
		"daily_person":       `daily_reward\s*=\s*\{`,
		"total_person":       `total_reward\s*=\s*\{`,
		"total_person_super": `super_reward\s*=\s*\{`,
	}
	for rankType, startPattern := range rewardPatterns {
		tableContent, ok := extractTableByPattern(page1Content, startPattern)
		if !ok {
			continue
		}
		rewardConfig, err := parseRewardTable(tableContent)
		if err == nil && len(rewardConfig) > 0 {
			config.Rewards[rankType] = rewardConfig
		}
	}

	// 超级奖限制（page[1] 内）
	superNum := extractNumber(page1Content, `super_num\s*=\s*(\d+)`)
	superLimit := extractNumber(page1Content, `super_limit\s*=\s*(\d+)`)
	if superNum > 0 || superLimit > 0 {
		config.SuperLimits["total_person_super"] = SuperLimit{
			MaxRank:  superNum,
			MinScore: superLimit,
		}
	}

	return config, nil
}

// extractString 提取字符串
func extractString(luaContent string, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(luaContent)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractBracketTable 提取形如 [index] = { ... } 的块
func extractBracketTable(luaContent string, index int) string {
	pattern := fmt.Sprintf(`\[%d\]\s*=\s*\{`, index)
	tableContent, ok := extractTableByPattern(luaContent, pattern)
	if ok {
		return tableContent
	}
	return ""
}

// extractTableByPattern 根据起始pattern提取Lua表内容
func extractTableByPattern(luaContent string, startPattern string) (string, bool) {
	startRe := regexp.MustCompile(startPattern)
	startMatch := startRe.FindStringIndex(luaContent)
	if startMatch == nil {
		return "", false
	}

	startPos := startMatch[1]
	// startPos points right after the opening "{", so depth starts at 1
	depth := 1
	endPos := -1
	for i := startPos; i < len(luaContent); i++ {
		if luaContent[i] == '{' {
			depth++
		} else if luaContent[i] == '}' {
			depth--
			if depth == 0 {
				endPos = i
				break
			}
		}
	}

	if endPos > startPos {
		return luaContent[startPos:endPos], true
	}
	return "", false
}

// extractNumber 提取单个数字
func extractNumber(luaContent string, pattern string) int {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(luaContent)
	if len(matches) > 1 {
		if v, err := strconv.Atoi(matches[1]); err == nil {
			return v
		}
	}
	return 0
}

// parseYxdsConfig 解析yxds类型的Lua配置
func parseYxdsConfig(luaContent string, config *LuaConfig) (*LuaConfig, error) {
	// YXDS使用独立字段：mail_title / mail_daily_content / mail_total_content / mail_super_content
	title := ""
	titleMatch := regexp.MustCompile(`mail_title\s*=\s*['"]([^'"]+)['"]`).FindStringSubmatch(luaContent)
	if len(titleMatch) > 1 {
		title = titleMatch[1]
	}

	if title != "" {
		config.Titles["daily_person"] = title + "日榜奖励"
		config.Titles["total_person"] = title + "总榜奖励"
		config.Titles["total_person_super"] = title + "超级奖奖励"
	}

	contentPatterns := map[string]string{
		"daily_person":       `mail_daily_content\s*=\s*['"]([^'"]+)['"]`,
		"total_person":       `mail_total_content\s*=\s*['"]([^'"]+)['"]`,
		"total_person_super": `mail_super_content\s*=\s*['"]([^'"]+)['"]`,
	}
	for rankType, pattern := range contentPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(luaContent)
		if len(matches) > 1 {
			config.Contents[rankType] = matches[1]
		}
	}

	rewardPatterns := map[string]string{
		"daily_person":       `daily_reward\s*=\s*\{`,
		"total_person":       `total_reward\s*=\s*\{`,
		"total_person_super": `super_reward\s*=\s*\{`,
	}
	for rankType, startPattern := range rewardPatterns {
		tableContent, ok := extractTableByPattern(luaContent, startPattern)
		if !ok {
			continue
		}
		rewardConfig, err := parseRewardTable(tableContent)
		if err == nil && len(rewardConfig) > 0 {
			config.Rewards[rankType] = rewardConfig
		}
	}

	superNum := extractNumber(luaContent, `super_num\s*=\s*(\d+)`)
	superLimit := extractNumber(luaContent, `super_limit\s*=\s*(\d+)`)
	if superNum > 0 || superLimit > 0 {
		config.SuperLimits["total_person_super"] = SuperLimit{
			MaxRank:  superNum,
			MinScore: superLimit,
		}
	}

	return config, nil
}

// parseRewardTable 解析奖励表，格式如: [1] = { [313200] = 200, [180301] = 280 }, [2] = { [313200] = 120, [180301] = 180 }
func parseRewardTable(tableContent string) (map[string]map[string]int, error) {
	rewardConfig := make(map[string]map[string]int)

	// 匹配每个排名配置: [排名] = { [道具ID] = 数量, ... }
	// 使用非贪婪匹配，并处理多行情况
	rankPattern := regexp.MustCompile(`(?s)\[(\d+)\]\s*=\s*\{([^}]+)\}`)
	matches := rankPattern.FindAllStringSubmatch(tableContent, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		rank := match[1]
		itemsContent := match[2]

		// 解析道具列表: [道具ID] = 数量
		// 支持逗号分隔的多个道具
		itemPattern := regexp.MustCompile(`\[(\d+)\]\s*=\s*(\d+)`)
		itemMatches := itemPattern.FindAllStringSubmatch(itemsContent, -1)

		items := make(map[string]int)
		for _, itemMatch := range itemMatches {
			if len(itemMatch) >= 3 {
				itemID := itemMatch[1]
				count, err := strconv.Atoi(itemMatch[2])
				if err == nil {
					items[itemID] = count
				}
			}
		}

		if len(items) > 0 {
			rewardConfig[rank] = items
		}
	}

	return rewardConfig, nil
}

// ToJSON 将配置转换为JSON格式
func (c *LuaConfig) ToJSON() (string, error) {
	jsonData, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// GetRewardConfigForRankType 获取指定榜单类型的奖励配置（用于补发）
func (c *LuaConfig) GetRewardConfigForRankType(rankType string) map[string]map[string]int {
	if reward, ok := c.Rewards[rankType]; ok {
		if rewardMap, ok := reward.(map[string]map[string]int); ok {
			return rewardMap
		}
	}
	return nil
}

// GetTitleForRankType 获取指定榜单类型的标题
func (c *LuaConfig) GetTitleForRankType(rankType string) string {
	if title, ok := c.Titles[rankType]; ok {
		return title
	}
	return ""
}

// GetContentForRankType 获取指定榜单类型的内容
func (c *LuaConfig) GetContentForRankType(rankType string) string {
	if content, ok := c.Contents[rankType]; ok {
		return content
	}
	return ""
}
