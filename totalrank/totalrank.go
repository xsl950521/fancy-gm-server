package totalrank

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

// TotalRankKeyInfo 用于存储解析后的总榜键信息
type TotalRankKeyInfo struct {
	ActTime  string // 活动开启时间
	GroupId  string // 分组标记
	RankType string // 排行榜类型 (total_rank/total_rank_club)
}

// TotalRankEntry 用于存储总榜分数条目（无 Day 字段）
type TotalRankEntry struct {
	GroupId        string
	Member         string
	OriginalScore  float64
	ProcessedScore float64
	Rank           int
}

// ParseTotalRankKey 解析总榜键信息
// 总榜键格式: hd_bydr_820800000_2918406_hd_bydr_total_rank 或 hd_bydr_820800000_2918406_hd_bydr_total_rank_club
func ParseTotalRankKey(key string) (TotalRankKeyInfo, bool) {
	parts := strings.Split(key, "_")
	info := TotalRankKeyInfo{}

	// 检查是否包含 total_rank
	if !strings.Contains(key, "total_rank") {
		return info, false
	}

	// 总榜键通常至少有 7 个部分
	if len(parts) >= 7 {
		info.ActTime = parts[2] // 活动开启时间
		info.GroupId = parts[3] // 分组标记

		// 判断类型
		if strings.Contains(key, "total_rank_club") {
			info.RankType = "total_rank_club"
		} else if strings.Contains(key, "total_rank") {
			info.RankType = "total_rank"
		}
	}

	return info, info.GroupId != "" && info.RankType != ""
}

// ReadTotalRankData 从文件读取并解析总榜数据
// 返回按 groupId 和 rankType 分组的数据: [groupId][rankType][]TotalRankEntry
func ReadTotalRankData(filePath string) (map[string]map[string][]TotalRankEntry, error) {
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

	// 存储数据，按 groupId 和 rankType 分组
	// [groupId][rankType][]TotalRankEntry
	rankData := make(map[string]map[string][]TotalRankEntry)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 匹配Key行
		if keyMatch := keyRegex.FindStringSubmatch(line); keyMatch != nil {
			key := keyMatch[1]
			// 解析键信息
			keyInfo, isValid := ParseTotalRankKey(key)
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
				entry := TotalRankEntry{
					GroupId:       keyInfo.GroupId,
					Member:        match[1],
					OriginalScore: score,
				}

				// 确保 groupId 和 rankType 的 map 存在
				if rankData[keyInfo.GroupId] == nil {
					rankData[keyInfo.GroupId] = make(map[string][]TotalRankEntry)
				}
				rankData[keyInfo.GroupId][keyInfo.RankType] = append(rankData[keyInfo.GroupId][keyInfo.RankType], entry)
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

// ProcessTotalRankData 处理总榜数据：计算排名和处理分数
func ProcessTotalRankData(rankData map[string]map[string][]TotalRankEntry) {
	// 处理每个组别
	for _, rankTypeMap := range rankData {
		// 处理每个排行榜类型
		for _, entries := range rankTypeMap {
			// 按原始分数降序排序（分数越高排名越高）
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].OriginalScore > entries[j].OriginalScore
			})

			// 计算排名（处理相同分数的情况）
			currentRank := 0
			lastScore := math.Inf(-1) // 初始值为负无穷
			for i := range entries {
				if entries[i].OriginalScore != lastScore {
					// 分数不同时，更新当前排名
					currentRank = i + 1
					lastScore = entries[i].OriginalScore
				}
				// 分配排名（相同分数有相同排名）
				entries[i].Rank = currentRank

				// 处理分数值
				modValue := math.Mod(entries[i].OriginalScore, 100000000)
				entries[i].ProcessedScore = (entries[i].OriginalScore - modValue) / 100000000
			}
		}
	}
}

// WriteTotalRankToExcel 将总榜数据写入Excel文件
// rankData 按 groupId 和 rankType 分组: [groupId][rankType][]TotalRankEntry
// 如果工作表已存在，将追加数据而不是覆盖
func WriteTotalRankToExcel(f *excelize.File, sheetName string, rankData map[string]map[string][]TotalRankEntry) {
	// 检查工作表是否存在
	sheetExists := worksheetExists(f, sheetName)
	var startRow int
	var styleStartRow int

	if !sheetExists {
		// 创建工作表
		f.NewSheet(sheetName)

		// 设置表头（无 Day 列）
		f.SetCellValue(sheetName, "A1", "组别")
		f.SetCellValue(sheetName, "B1", "Rank")
		f.SetCellValue(sheetName, "C1", "Member")
		f.SetCellValue(sheetName, "D1", "Processed Score")
		f.SetCellValue(sheetName, "E1", "Original Score")

		// 设置列宽
		f.SetColWidth(sheetName, "A", "A", 15) // 组别
		f.SetColWidth(sheetName, "B", "B", 8)  // Rank
		f.SetColWidth(sheetName, "C", "C", 20) // Member
		f.SetColWidth(sheetName, "D", "D", 15) // 处理后的分数
		f.SetColWidth(sheetName, "E", "E", 25) // 原始分数

		startRow = 2
		styleStartRow = 2
	} else {
		// 工作表已存在，找到最后一行
		lastRow := findLastRow(f, sheetName, []string{"A", "B", "C", "D", "E"})
		if lastRow == 0 {
			// 空工作表，从第2行开始（假设有表头）
			startRow = 2
		} else {
			// 从最后一行之后开始
			startRow = lastRow + 1
		}
		styleStartRow = startRow
	}

	rowNum := startRow

	// 按组别排序
	groupIds := make([]string, 0, len(rankData))
	for groupId := range rankData {
		groupIds = append(groupIds, groupId)
	}
	sort.Strings(groupIds)

	// 处理每个组别
	for _, groupId := range groupIds {
		rankTypeMap := rankData[groupId]

		// 按排行榜类型排序（先处理 total_rank，再处理 total_rank_club）
		rankTypes := make([]string, 0, len(rankTypeMap))
		for rankType := range rankTypeMap {
			rankTypes = append(rankTypes, rankType)
		}
		sort.Strings(rankTypes)

		// 处理每个排行榜类型
		for _, rankType := range rankTypes {
			entries := rankTypeMap[rankType]

			// 写入数据
			for _, entry := range entries {
				f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), entry.GroupId)
				f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), entry.Rank)
				f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), entry.Member)
				f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), entry.ProcessedScore)
				f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), entry.OriginalScore)
				rowNum++
			}
		}
	}

	// 设置排名列样式（居中）- 仅对新追加的数据应用样式
	rankStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	if rowNum > styleStartRow {
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", styleStartRow), fmt.Sprintf("B%d", rowNum-1), rankStyle)
	}

	// 设置处理后的分数列为整数格式
	scoreStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 1, // 整数格式
	})
	if rowNum > styleStartRow {
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", styleStartRow), fmt.Sprintf("D%d", rowNum-1), scoreStyle)
	}

	// 设置原始分数为科学计数法显示
	sciStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 11, // 科学计数法格式
	})
	if rowNum > styleStartRow {
		f.SetCellStyle(sheetName, fmt.Sprintf("E%d", styleStartRow), fmt.Sprintf("E%d", rowNum-1), sciStyle)
	}
}

// ProcessTotalRankFile 处理总榜文件的完整流程：读取、处理、写入Excel
// 返回处理后的数据和错误
func ProcessTotalRankFile(filePath string, excelFile *excelize.File, personalSheetName, clubSheetName string) error {
	// 读取数据
	rankData, err := ReadTotalRankData(filePath)
	if err != nil {
		return fmt.Errorf("failed to read total rank data: %w", err)
	}

	// 处理数据（计算排名和处理分数）
	ProcessTotalRankData(rankData)

	// 分离个人榜和公会榜数据
	personalRank := make(map[string]map[string][]TotalRankEntry)
	clubRank := make(map[string]map[string][]TotalRankEntry)

	for groupId, rankTypeMap := range rankData {
		for rankType, entries := range rankTypeMap {
			if rankType == "total_rank" {
				if personalRank[groupId] == nil {
					personalRank[groupId] = make(map[string][]TotalRankEntry)
				}
				personalRank[groupId][rankType] = entries
			} else if rankType == "total_rank_club" {
				if clubRank[groupId] == nil {
					clubRank[groupId] = make(map[string][]TotalRankEntry)
				}
				clubRank[groupId][rankType] = entries
			}
		}
	}

	// 写入Excel
	if len(personalRank) > 0 {
		WriteTotalRankToExcel(excelFile, personalSheetName, personalRank)
	}

	if len(clubRank) > 0 {
		WriteTotalRankToExcel(excelFile, clubSheetName, clubRank)
	}

	return nil
}

