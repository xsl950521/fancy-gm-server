package rewardrecord

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"
)

// maxScannerBufferSize 设置 bufio.Scanner 的最大缓冲区大小为 1MB
// 用于处理包含超长行的文件（如包含大量 Redis hash 数据的文件）
// 默认的 64KB 限制会导致 "token too long" 错误
const maxScannerBufferSize = 1048576 // 1MB = 1048576 bytes

// worksheetExists checks if a worksheet exists in the Excel file.
func worksheetExists(f *excelize.File, sheetName string) bool {
	sheetList := f.GetSheetList()
	for _, name := range sheetList {
		if name == sheetName {
			return true
		}
	}
	return false
}

// findLastRow finds the last row with data in a worksheet.
// Returns 1 if only headers exist, or 0 if worksheet is empty.
// Uses a more efficient approach: check from row 2 upwards with a reasonable limit.
func findLastRow(f *excelize.File, sheetName string, checkColumns []string) int {
	// Check if headers exist (row 1)
	hasHeader := false
	for _, col := range checkColumns {
		cellName := fmt.Sprintf("%s1", col)
		value, err := f.GetCellValue(sheetName, cellName)
		if err == nil && strings.TrimSpace(value) != "" {
			hasHeader = true
			break
		}
	}
	
	if !hasHeader {
		return 0 // Empty worksheet
	}
	
	// Start from row 2 and check upwards to find last row with data
	// Use a reasonable limit (100000 rows) to avoid infinite loops
	maxRows := 100000
	lastRowWithData := 1 // Default to row 1 (headers)
	
	for row := 2; row <= maxRows; row++ {
		hasData := false
		for _, col := range checkColumns {
			cellName := fmt.Sprintf("%s%d", col, row)
			value, err := f.GetCellValue(sheetName, cellName)
			if err == nil && strings.TrimSpace(value) != "" {
				hasData = true
				break
			}
		}
		if hasData {
			lastRowWithData = row
		} else if lastRowWithData < row-100 {
			// If we've gone 100 rows without finding data, assume we've reached the end
			// This optimization helps with large empty sections
			break
		}
	}
	
	return lastRowWithData
}

// RewardRecordKeyInfo 用于存储解析后的奖励记录键信息
type RewardRecordKeyInfo struct {
	ActTime string // 活动开启时间
	GroupId string // 分组标记
}

// ParsedRankingKey 用于存储解析后的排行榜key信息
type ParsedRankingKey struct {
	OriginalKey string // 原始完整的key
	RankType    string // 排行榜类型: "rank" (日榜) 或 "total_rank" (总榜)
	Day         string // 天数 (日榜才有，总榜为空)
	IsClub      bool   // 是否为公会榜
	IsTotalRank bool   // 是否为总榜
}

// RewardRecordEntry 用于存储奖励记录条目
type RewardRecordEntry struct {
	GroupId     string             // 分组标记
	UserID      string             // 用户ID
	RankingKeys []string           // 用户获得的排行榜key列表（原始）
	ParsedKeys  []ParsedRankingKey // 解析后的排行榜key信息列表
}

// ParseRewardRecordKey 解析奖励记录键信息
// 键格式: hd_bydr_820800000_2918404_hd_bydr_reward_record
func ParseRewardRecordKey(key string) (RewardRecordKeyInfo, bool) {
	parts := strings.Split(key, "_")
	info := RewardRecordKeyInfo{}

	// 检查是否包含 reward_record
	if !strings.Contains(key, "reward_record") {
		return info, false
	}

	// 奖励记录键通常至少有 7 个部分
	if len(parts) >= 7 {
		info.ActTime = parts[2] // 活动开启时间
		info.GroupId = parts[3] // 分组标记
	}

	return info, info.GroupId != ""
}

// ParseRewardRecordValue 解析奖励记录值（Lua table格式）
// 格式: '{[1]="key1",[2]="key2",...}'
func ParseRewardRecordValue(value string) []string {
	var keys []string
	// 匹配 [数字]="key" 格式
	re := regexp.MustCompile(`\[(\d+)\]="([^"]+)"`)
	matches := re.FindAllStringSubmatch(value, -1)

	for _, match := range matches {
		if len(match) == 3 {
			// 提取key值
			keys = append(keys, match[2])
		}
	}

	return keys
}

// ParseRankingKey 解析单个排行榜key，提取结构化信息
// 支持的格式:
// - hd_bydr_820800000_2918404_hd_bydr_rank_1 (日榜个人第1天)
// - hd_bydr_820800000_2918404_hd_bydr_rank_club_1 (日榜公会第1天)
// - hd_bydr_820800000_2918404_hd_bydr_total_rank (总榜个人)
// - hd_bydr_820800000_2918404_hd_bydr_total_rank_club (总榜公会)
func ParseRankingKey(key string) ParsedRankingKey {
	parsed := ParsedRankingKey{
		OriginalKey: key,
		RankType:    "",
		Day:         "",
		IsClub:      false,
		IsTotalRank: false,
	}

	// 检查是否为总榜
	if strings.Contains(key, "total_rank") {
		parsed.IsTotalRank = true
		parsed.RankType = "total_rank"

		// 检查是否为公会榜
		if strings.Contains(key, "total_rank_club") {
			parsed.IsClub = true
		}
	} else if strings.Contains(key, "rank") {
		// 日榜
		parsed.IsTotalRank = false
		parsed.RankType = "rank"

		// 检查是否为公会榜
		if strings.Contains(key, "rank_club") {
			parsed.IsClub = true
		}

		// 提取天数：key的最后一部分应该是天数
		parts := strings.Split(key, "_")
		if len(parts) > 0 {
			// 尝试从最后一部分提取天数
			lastPart := parts[len(parts)-1]
			// 如果最后一部分是数字，则是天数
			if matched, _ := regexp.MatchString(`^\d+$`, lastPart); matched {
				parsed.Day = lastPart
			}
		}
	}

	return parsed
}

// ReadRewardRecordData 从文件读取并解析奖励记录数据
// 返回按 groupId 分组的数据: [groupId][]RewardRecordEntry
func ReadRewardRecordData(filePath string) (map[string][]RewardRecordEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 准备正则表达式
	// Key: hd_bydr_820800000_2918404_hd_bydr_reward_record, Type: hash, Value: {'1308317241': '{[1]="...",[2]="..."}', ...}
	keyRegex := regexp.MustCompile(`^Key: (.+), Type: hash, Value: (.+)$`)
	// 匹配 hash 值: 'userid': '{...}'
	// 由于值中包含嵌套的引号和括号，使用更复杂的匹配
	// 匹配模式: '数字': '{...}' 其中 {...} 可能包含引号
	hashValueRegex := regexp.MustCompile(`'(\d+)':\s*'({(?:[^']|'[^'}])*})'`)

	scanner := bufio.NewScanner(file)
	// 设置缓冲区大小以支持超长行（最多 1MB）
	scanner.Buffer(nil, maxScannerBufferSize)

	// 存储数据，按 groupId 分组
	// [groupId][]RewardRecordEntry
	recordData := make(map[string][]RewardRecordEntry)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 匹配Key行
		if keyMatch := keyRegex.FindStringSubmatch(line); keyMatch != nil {
			key := keyMatch[1]
			value := keyMatch[2]

			// 解析键信息
			keyInfo, isValid := ParseRewardRecordKey(key)
			if !isValid {
				continue
			}

			// 解析hash值（用户ID -> 排行榜key列表）
			hashMatches := hashValueRegex.FindAllStringSubmatch(value, -1)
			for _, match := range hashMatches {
				if len(match) != 3 {
					continue
				}

				userID := match[1]
				luaTableValue := match[2]

				// 解析Lua table获取排行榜key列表
				rankingKeys := ParseRewardRecordValue(luaTableValue)

				// 解析每个ranking key
				parsedKeys := make([]ParsedRankingKey, 0, len(rankingKeys))
				for _, key := range rankingKeys {
					parsedKeys = append(parsedKeys, ParseRankingKey(key))
				}

				// 创建条目
				entry := RewardRecordEntry{
					GroupId:     keyInfo.GroupId,
					UserID:      userID,
					RankingKeys: rankingKeys,
					ParsedKeys:  parsedKeys,
				}

				recordData[keyInfo.GroupId] = append(recordData[keyInfo.GroupId], entry)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		// 检查是否是 "token too long" 错误，提供更清晰的错误消息
		if strings.Contains(err.Error(), "token too long") {
			return nil, fmt.Errorf("error reading file: line too long (exceeds %d bytes). %w", maxScannerBufferSize, err)
		}
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return recordData, nil
}

// WriteRewardRecordToExcel 将奖励记录数据写入Excel文件
// 使用长格式：每个ranking key占一行，三列：组别、UserID、Key
// 数据按层级组织：组别 → Server ID (UserID) → Key
// 排序顺序：先按组别排序，组内按UserID排序，每个UserID的keys也排序
// recordData 按 groupId 分组: [groupId][]RewardRecordEntry
// 如果工作表已存在，将追加数据而不是覆盖
func WriteRewardRecordToExcel(f *excelize.File, sheetName string, recordData map[string][]RewardRecordEntry) {
	// 检查工作表是否存在
	sheetExists := worksheetExists(f, sheetName)
	var startRow int

	if !sheetExists {
		// 创建工作表
		f.NewSheet(sheetName)

		// 设置表头：三列
		f.SetCellValue(sheetName, "A1", "组别")
		f.SetCellValue(sheetName, "B1", "UserID")
		f.SetCellValue(sheetName, "C1", "Key")

		// 设置列宽
		f.SetColWidth(sheetName, "A", "A", 15) // 组别
		f.SetColWidth(sheetName, "B", "B", 20) // UserID
		f.SetColWidth(sheetName, "C", "C", 80) // Key

		startRow = 2
	} else {
		// 工作表已存在，找到最后一行
		lastRow := findLastRow(f, sheetName, []string{"A", "B", "C"})
		if lastRow == 0 {
			// 空工作表，从第2行开始（假设有表头）
			startRow = 2
		} else {
			// 从最后一行之后开始
			startRow = lastRow + 1
		}
	}

	rowNum := startRow

	// 按组别排序，确保组别按字母顺序排列
	groupIds := make([]string, 0, len(recordData))
	for groupId := range recordData {
		groupIds = append(groupIds, groupId)
	}
	sort.Strings(groupIds)

	// 处理每个组别
	for _, groupId := range groupIds {
		entries := recordData[groupId]

		// 按UserID排序，确保同一组别内的Server ID按顺序排列
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].UserID < entries[j].UserID
		})

		// 为每个用户的每个key创建一行
		for _, entry := range entries {
			// 对keys进行排序，确保每个Server ID的keys按顺序排列
			keys := make([]string, len(entry.RankingKeys))
			copy(keys, entry.RankingKeys)
			sort.Strings(keys)

			// 为每个key创建一行，显示层级关系：组别 → Server ID → Key
			for _, key := range keys {
				f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), entry.GroupId)
				f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), entry.UserID)
				f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), key)
				rowNum++
			}
		}
	}
}

// ProcessRewardRecordFile 处理奖励记录文件的完整流程：读取、写入Excel
func ProcessRewardRecordFile(filePath string, excelFile *excelize.File, sheetName string) error {
	// 读取数据
	recordData, err := ReadRewardRecordData(filePath)
	if err != nil {
		return fmt.Errorf("failed to read reward record data: %w", err)
	}

	// 写入Excel
	if len(recordData) > 0 {
		WriteRewardRecordToExcel(excelFile, sheetName, recordData)
	}

	return nil
}
