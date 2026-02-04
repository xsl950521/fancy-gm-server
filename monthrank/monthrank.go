// Package monthrank provides functionality to parse month rank data files
// from text files (.txt) containing Redis zset data and export them to Excel format.
package monthrank

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// maxScannerBufferSize 设置 bufio.Scanner 的最大缓冲区大小为 1MB
// 用于处理包含超长行的文件（如包含大量 Redis zset 数据的文件）
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

// MonthRankKeyInfo 用于存储解析后的月榜键信息
// Key 格式: hd_month_rank_rewards_820540800_2601132302_hd_month_stage_rank_1
type MonthRankKeyInfo struct {
	ActTime  string // 活动开启时间 (parts[5]: 2601132302)
	GroupId  string // 分组标记 (parts[4]: 820540800)
	Stage    string // 阶段 (parts[10]: 1)
	RankType string // 排行榜类型 (month_stage_rank)
}

// MonthRankEntry 用于存储月榜分数条目
type MonthRankEntry struct {
	GroupId        string
	Member         string
	OriginalScore  float64
	ProcessedScore float64
	Rank           int
	Stage          string
}

// ParseMonthRankKey 解析月榜键信息
// 月榜键格式: hd_month_rank_rewards_820540800_2601132302_hd_month_stage_rank_1
// 按 _ 分割: ["hd", "month", "rank", "rewards", "820540800", "2601132302", "hd", "month", "stage", "rank", "1"]
func ParseMonthRankKey(key string) (MonthRankKeyInfo, bool) {
	parts := strings.Split(key, "_")
	info := MonthRankKeyInfo{}

	// 检查是否包含 month_rank_rewards
	if !strings.Contains(key, "month_rank_rewards") {
		return info, false
	}

	// 月榜键通常至少有 11 个部分
	if len(parts) >= 11 {
		// parts[4] = "820540800" (GroupId)
		// parts[5] = "2601132302" (ActTime)
		// parts[10] = "1" (Stage)
		info.GroupId = parts[4]
		info.ActTime = parts[5]
		info.Stage = parts[10]
		info.RankType = "month_stage_rank"
	}

	return info, info.GroupId != "" && info.Stage != ""
}

// ReadMonthRankData 从文件读取并解析月榜数据
// 返回按 groupId 和 stage 分组的数据: [groupId][stage][]MonthRankEntry
func ReadMonthRankData(filePath string) (map[string]map[string][]MonthRankEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 准备正则表达式
	keyRegex := regexp.MustCompile(`^Key: (.+), Type: zset, Value: (\[.*)$`)
	valueRegex := regexp.MustCompile(`\('([^']+)',\s*([\d.]+)\)`)

	scanner := bufio.NewScanner(file)
	// 设置缓冲区大小以支持超长行（最多 1MB）
	scanner.Buffer(nil, maxScannerBufferSize)

	// 存储数据，按 groupId 和 stage 分组
	// [groupId][stage][]MonthRankEntry
	rankData := make(map[string]map[string][]MonthRankEntry)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 匹配Key行
		if keyMatch := keyRegex.FindStringSubmatch(line); keyMatch != nil {
			key := keyMatch[1]
			// 解析键信息
			keyInfo, isValid := ParseMonthRankKey(key)
			if !isValid {
				continue
			}

			// 解析Value数据
			valueMatches := valueRegex.FindAllStringSubmatch(keyMatch[2], -1)
			for _, match := range valueMatches {
				if len(match) != 3 {
					continue
				}

				score, err := strconv.ParseFloat(match[2], 64)
				if err != nil {
					fmt.Printf("Error parsing score: %v\n", err)
					continue
				}

				// 创建条目
				entry := MonthRankEntry{
					GroupId:       keyInfo.GroupId,
					Member:        match[1],
					OriginalScore: score,
					Stage:         keyInfo.Stage,
				}

				// 确保 groupId 和 stage 的 map 存在
				if rankData[keyInfo.GroupId] == nil {
					rankData[keyInfo.GroupId] = make(map[string][]MonthRankEntry)
				}
				rankData[keyInfo.GroupId][keyInfo.Stage] = append(rankData[keyInfo.GroupId][keyInfo.Stage], entry)
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

	return rankData, nil
}

// WriteMonthRankToExcel 将月榜数据写入Excel文件
// 如果工作表已存在，将追加数据而不是覆盖
func WriteMonthRankToExcel(f *excelize.File, sheetName string, entries []MonthRankEntry) {
	// 检查工作表是否存在
	sheetExists := worksheetExists(f, sheetName)
	var startRow int

	if !sheetExists {
		// 创建工作表
		f.NewSheet(sheetName)

		// 设置表头
		headers := []string{"GroupId", "Stage", "Rank", "Member", "OriginalScore", "ProcessedScore"}
		for i, header := range headers {
			cellName := fmt.Sprintf("%c1", 'A'+i)
			f.SetCellValue(sheetName, cellName, header)
		}

		// 设置列宽
		f.SetColWidth(sheetName, "A", "A", 15) // GroupId
		f.SetColWidth(sheetName, "B", "B", 10) // Stage
		f.SetColWidth(sheetName, "C", "C", 10) // Rank
		f.SetColWidth(sheetName, "D", "D", 20) // Member
		f.SetColWidth(sheetName, "E", "E", 15) // OriginalScore
		f.SetColWidth(sheetName, "F", "F", 15) // ProcessedScore

		startRow = 2
	} else {
		// 工作表已存在，找到最后一行
		checkColumns := []string{"A", "B", "C", "D", "E", "F"}
		lastRow := findLastRow(f, sheetName, checkColumns)
		if lastRow == 0 {
			// 空工作表，从第2行开始（假设有表头）
			startRow = 2
		} else {
			// 从最后一行之后开始
			startRow = lastRow + 1
		}
	}

	// 写入数据行
	rowNum := startRow
	for _, entry := range entries {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), entry.GroupId)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), entry.Stage)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), entry.Rank)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), entry.Member)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), entry.OriginalScore)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), entry.ProcessedScore)
		rowNum++
	}
}

// ProcessMonthRankFile 处理单个月榜文件，读取数据并写入Excel
// 返回处理后的数据和错误
func ProcessMonthRankFile(filePath string, excelFile *excelize.File, sheetName string) error {
	// 读取数据
	rankData, err := ReadMonthRankData(filePath)
	if err != nil {
		return fmt.Errorf("failed to read month rank data: %w", err)
	}

	// 处理每个 groupId 和 stage 的数据
	for groupId, stageData := range rankData {
		for stage, entries := range stageData {
			// 按分数降序排序
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].OriginalScore > entries[j].OriginalScore
			})

			// 计算排名和处理后的分数
			for i := range entries {
				entries[i].Rank = i + 1
				// 处理后的分数：取对数并乘以1000，然后四舍五入
				if entries[i].OriginalScore > 0 {
					entries[i].ProcessedScore = math.Round(math.Log(entries[i].OriginalScore+1) * 1000)
				} else {
					entries[i].ProcessedScore = 0
				}
			}

			// 生成工作表名称：MonthRank_GroupId_Stage
			sheetNameWithGroup := fmt.Sprintf("%s_%s_%s", sheetName, groupId, stage)
			if len(sheetNameWithGroup) > 31 {
				// Excel 工作表名称限制为 31 个字符
				sheetNameWithGroup = sheetNameWithGroup[:31]
			}

			// 写入Excel
			WriteMonthRankToExcel(excelFile, sheetNameWithGroup, entries)
		}
	}

	return nil
}
