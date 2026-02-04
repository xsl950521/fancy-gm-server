// Package clubpid provides functionality to parse club pid data files
// from text files (.txt) containing Redis string data and export them to Excel format.
package clubpid

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// maxScannerBufferSize 设置 bufio.Scanner 的最大缓冲区大小为 1MB
// 用于处理包含超长行的文件
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

// ClubPidKeyInfo 用于存储解析后的 club pid 键信息
// Key 格式: hd_bydr_822700800_29038201_hd_club_pid_8642639_195780154
// 解析格式: hd_bydr_acttime_hid_hd_club_pid_clubid_psid
type ClubPidKeyInfo struct {
	ActTime int64  // 活动开启时间 (parts[2])
	Hid     int64  // 分组标记 (parts[3])
	ClubId  int64  // 俱乐部ID (parts[7])
	PSid    int64  // 玩家ID (parts[8])
}

// ClubPidEntry 用于存储 club pid 条目
type ClubPidEntry struct {
	ActTime int64 // 活动开启时间
	Hid     int64 // 分组标记
	ClubId  int64 // 俱乐部ID
	PSid    int64 // 玩家ID
	Score   int64 // 分数
}

// ParseClubPidKey 解析 club pid 键信息
// Key 格式: hd_bydr_822700800_29038201_hd_club_pid_8642639_195780154
// 按 _ 分割: ["hd", "bydr", "822700800", "29038201", "hd", "club", "pid", "8642639", "195780154"]
func ParseClubPidKey(key string) (ClubPidKeyInfo, bool) {
	parts := strings.Split(key, "_")
	info := ClubPidKeyInfo{}

	// 检查是否包含 club_pid
	if !strings.Contains(key, "club_pid") {
		return info, false
	}

	// club pid 键通常至少有 9 个部分
	if len(parts) >= 9 {
		// parts[2] = "822700800" (ActTime)
		// parts[3] = "29038201" (Hid)
		// parts[7] = "8642639" (ClubId)
		// parts[8] = "195780154" (PSid)

		actTime, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return info, false
		}

		hid, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			return info, false
		}

		clubId, err := strconv.ParseInt(parts[7], 10, 64)
		if err != nil {
			return info, false
		}

		psid, err := strconv.ParseInt(parts[8], 10, 64)
		if err != nil {
			return info, false
		}

		info.ActTime = actTime
		info.Hid = hid
		info.ClubId = clubId
		info.PSid = psid

		return info, true
	}

	return info, false
}

// ReadClubPidData 从文件读取并解析 club pid 数据
// 文件格式：每行包含 Key, Type, Value
// Key: hd_bydr_822700800_29038201_hd_club_pid_8642639_195780154, Type: string, Value: 68200
func ReadClubPidData(filePath string) ([]ClubPidEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 准备正则表达式
	// 匹配格式: Key: <key>, Type: string, Value: <value>
	keyValueRegex := regexp.MustCompile(`^Key: (.+), Type: string, Value: (\d+)$`)

	scanner := bufio.NewScanner(file)
	// 设置缓冲区大小以支持超长行（最多 1MB）
	scanner.Buffer(nil, maxScannerBufferSize)

	var entries []ClubPidEntry
	lineNum := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		// 跳过空行
		if line == "" {
			continue
		}

		// 匹配 Key 和 Value
		matches := keyValueRegex.FindStringSubmatch(line)
		if len(matches) != 3 {
			fmt.Printf("Warning: Line %d, failed to parse format. Skipping.\n", lineNum)
			continue
		}

		key := matches[1]
		valueStr := matches[2]

		// 解析键信息
		keyInfo, isValid := ParseClubPidKey(key)
		if !isValid {
			fmt.Printf("Warning: Line %d, invalid key format '%s'. Skipping.\n", lineNum, key)
			continue
		}

		// 解析分数值
		score, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			fmt.Printf("Warning: Line %d, invalid score '%s': %v. Skipping.\n", lineNum, valueStr, err)
			continue
		}

		// 创建条目
		entry := ClubPidEntry{
			ActTime: keyInfo.ActTime,
			Hid:     keyInfo.Hid,
			ClubId:  keyInfo.ClubId,
			PSid:    keyInfo.PSid,
			Score:   score,
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		// 检查是否是 "token too long" 错误，提供更清晰的错误消息
		if strings.Contains(err.Error(), "token too long") {
			return nil, fmt.Errorf("error reading file: line too long (exceeds %d bytes). %w", maxScannerBufferSize, err)
		}
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return entries, nil
}

// WriteClubPidToExcel 将 club pid 数据写入Excel文件
// 如果工作表已存在，将追加数据而不是覆盖
func WriteClubPidToExcel(f *excelize.File, sheetName string, entries []ClubPidEntry) {
	// 检查工作表是否存在
	sheetExists := worksheetExists(f, sheetName)
	var startRow int

	if !sheetExists {
		// 创建工作表
		f.NewSheet(sheetName)

		// 设置表头
		headers := []string{"ActTime", "Hid", "ClubId", "PSid", "Score"}
		for i, header := range headers {
			cellName := fmt.Sprintf("%c1", 'A'+i)
			f.SetCellValue(sheetName, cellName, header)
		}

		// 设置列宽
		f.SetColWidth(sheetName, "A", "A", 15) // ActTime
		f.SetColWidth(sheetName, "B", "B", 15) // Hid
		f.SetColWidth(sheetName, "C", "C", 15) // ClubId
		f.SetColWidth(sheetName, "D", "D", 20) // PSid
		f.SetColWidth(sheetName, "E", "E", 15) // Score

		startRow = 2
	} else {
		// 工作表已存在，找到最后一行
		checkColumns := []string{"A", "B", "C", "D", "E"}
		lastRow := findLastRow(f, sheetName, checkColumns)
		if lastRow == 0 {
			// 空工作表，从第2行开始（假设有表头）
			startRow = 2
		} else {
			// 从最后一行之后开始
			startRow = lastRow + 1
		}
	}

	// 写入数据行（所有字段都是数值类型）
	rowNum := startRow
	for _, entry := range entries {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), entry.ActTime)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), entry.Hid)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), entry.ClubId)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), entry.PSid)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), entry.Score)
		rowNum++
	}
}

// ProcessClubPidFile 处理 club pid 文件的完整流程：读取、写入Excel
func ProcessClubPidFile(filePath string, excelFile *excelize.File, sheetName string) error {
	// 读取数据
	entries, err := ReadClubPidData(filePath)
	if err != nil {
		return fmt.Errorf("failed to read club pid data: %w", err)
	}

	// 写入Excel
	WriteClubPidToExcel(excelFile, sheetName, entries)

	return nil
}
