// Package maillog provides functionality to parse mail log files
// from text files (.txt) containing comma-separated values with mail-related data
// and export them to Excel format.
package maillog

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// MailLogEntry 用于存储解析后的邮件日志条目
// 字段顺序：sid, gid, uid, roleid, rolename, isgm, caty, title, content, sendtime, optime, items, reason, op
type MailLogEntry struct {
	SID      int64  // 服务器ID
	GID      int    // 组ID
	UID      string // 用户ID (MAC地址格式)
	RoleID   int64  // 角色ID
	RoleName string // 角色名称 (URL编码，需要解码)
	IsGM     int    // 是否为GM (0或1)
	Caty     string // 类别
	Title    string // 标题
	Content  string // 内容
	Rank     int    // 排名 (从content字段解析，格式：^^排名^^积分)
	Score    int64  // 积分 (从content字段解析，格式：^^排名^^积分)
	SendTime int64  // 发送时间 (时间戳)
	OpTime   int64  // 操作时间 (时间戳)
	Items    string // 物品 (Lua表格式，作为字符串存储)
	Reason   string // 原因
	Op       int    // 操作类型
}

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
			break
		}
	}

	return lastRowWithData
}

// decodeURLField 解码URL编码的字段，如果解码失败则返回原始值
func decodeURLField(encoded string) string {
	decoded, err := url.QueryUnescape(encoded)
	if err != nil {
		// 如果解码失败，返回原始值
		return encoded
	}
	return decoded
}

// parseContentField 从content字段中解析排名和积分
// content字段格式：...^^排名^^积分
// 例如：@@尊敬的守护者你好...^^42^^168000
// 返回排名和积分，如果解析失败则返回0
func parseContentField(content string) (rank int, score int64) {
	// 使用正则表达式匹配 ^^排名^^积分 格式
	// 匹配模式：^^数字^^数字
	rankScoreRegex := regexp.MustCompile(`\^\^(\d+)\^\^(\d+)`)
	matches := rankScoreRegex.FindStringSubmatch(content)

	if len(matches) == 3 {
		// matches[0] 是完整匹配，matches[1] 是排名，matches[2] 是积分
		if r, err := strconv.Atoi(matches[1]); err == nil {
			rank = r
		}
		if s, err := strconv.ParseInt(matches[2], 10, 64); err == nil {
			score = s
		}
	}

	return rank, score
}

// parseMailLogLine 解析单行邮件日志数据
// 由于items字段包含逗号（如 {[605]=35,[313200]=60,[550]=40}），不能简单使用CSV解析
// 需要手动解析，识别items字段（以{开始，以}结束）
func parseMailLogLine(line string) ([]string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	var fields []string
	var currentField strings.Builder
	inItemsField := false
	braceCount := 0
	runes := []rune(line)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == '{' {
			// 开始items字段
			if !inItemsField {
				inItemsField = true
				braceCount = 1
			} else {
				braceCount++
			}
			currentField.WriteRune(r)
		} else if r == '}' && inItemsField {
			// 结束items字段
			currentField.WriteRune(r)
			braceCount--
			if braceCount == 0 {
				// items字段结束
				fields = append(fields, currentField.String())
				currentField.Reset()
				inItemsField = false
				// 跳过items字段后的逗号（如果存在）
				if i+1 < len(runes) && runes[i+1] == ',' {
					i++ // 跳过逗号
				}
			}
		} else if r == ',' && !inItemsField {
			// 普通字段分隔符
			fields = append(fields, currentField.String())
			currentField.Reset()
		} else {
			// 普通字符
			currentField.WriteRune(r)
		}
	}

	// 添加最后一个字段
	if currentField.Len() > 0 {
		fields = append(fields, currentField.String())
	}

	return fields, nil
}

// ReadMailLogData 从txt文件读取并解析邮件日志数据
// 文件格式：每行包含12或14个逗号分隔的字段
// 12个字段：sid, gid, uid, roleid, rolename, isgm, caty, title, content, items, reason, op (缺少sendtime和optime)
// 14个字段：sid, gid, uid, roleid, rolename, isgm, caty, title, content, sendtime, optime, items, reason, op
// items字段包含逗号，格式如：{[605]=35,[313200]=60,[550]=40}
// 返回解析后的邮件日志条目列表
func ReadMailLogData(filePath string) ([]MailLogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// 设置缓冲区大小以支持超长行（最多 1MB）
	const maxScannerBufferSize = 1048576 // 1MB
	scanner.Buffer(nil, maxScannerBufferSize)

	var entries []MailLogEntry
	lineNum := 0

	// 逐行读取记录
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// 跳过空行
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 手动解析行（处理items字段中的逗号）
		record, err := parseMailLogLine(line)
		if err != nil {
			fmt.Printf("Warning: Line %d, failed to parse: %v. Skipping.\n", lineNum, err)
			continue
		}

		// 验证字段数量（支持12或14个字段）
		fieldCount := len(record)
		if fieldCount != 12 && fieldCount != 14 {
			fmt.Printf("Warning: Line %d has %d fields, expected 12 or 14. Skipping.\n", lineNum, fieldCount)
			continue
		}

		// 判断是否有 sendtime 和 optime 字段
		hasTimeFields := fieldCount == 14

		// 解析数字字段
		sid, err := strconv.ParseInt(strings.TrimSpace(record[0]), 10, 64)
		if err != nil {
			fmt.Printf("Warning: Line %d, invalid sid '%s': %v. Using 0.\n", lineNum, record[0], err)
			sid = 0
		}

		gid, err := strconv.Atoi(strings.TrimSpace(record[1]))
		if err != nil {
			fmt.Printf("Warning: Line %d, invalid gid '%s': %v. Using 0.\n", lineNum, record[1], err)
			gid = 0
		}

		roleid, err := strconv.ParseInt(strings.TrimSpace(record[3]), 10, 64)
		if err != nil {
			fmt.Printf("Warning: Line %d, invalid roleid '%s': %v. Using 0.\n", lineNum, record[3], err)
			roleid = 0
		}

		isgm, err := strconv.Atoi(strings.TrimSpace(record[5]))
		if err != nil {
			fmt.Printf("Warning: Line %d, invalid isgm '%s': %v. Using 0.\n", lineNum, record[5], err)
			isgm = 0
		}

		// 处理 sendtime 和 optime（如果存在）
		var sendtime, optime int64
		if hasTimeFields {
			sendtime, err = strconv.ParseInt(strings.TrimSpace(record[9]), 10, 64)
			if err != nil {
				fmt.Printf("Warning: Line %d, invalid sendtime '%s': %v. Using 0.\n", lineNum, record[9], err)
				sendtime = 0
			}

			optime, err = strconv.ParseInt(strings.TrimSpace(record[10]), 10, 64)
			if err != nil {
				fmt.Printf("Warning: Line %d, invalid optime '%s': %v. Using 0.\n", lineNum, record[10], err)
				optime = 0
			}
		} else {
			// 如果没有时间字段，使用默认值0
			sendtime = 0
			optime = 0
		}

		// 根据字段数量确定 op 字段的位置
		opIndex := fieldCount - 1 // 最后一个字段是 op
		op, err := strconv.Atoi(strings.TrimSpace(record[opIndex]))
		if err != nil {
			fmt.Printf("Warning: Line %d, invalid op '%s': %v. Using 0.\n", lineNum, record[opIndex], err)
			op = 0
		}

		// 处理字符串字段
		uid := strings.TrimSpace(record[2])
		rolename := decodeURLField(strings.TrimSpace(record[4]))
		caty := strings.TrimSpace(record[6])
		title := decodeURLField(strings.TrimSpace(record[7]))
		content := decodeURLField(strings.TrimSpace(record[8]))

		// 从content字段解析排名和积分
		rank, score := parseContentField(content)

		// items 和 reason 的位置取决于是否有时间字段
		var items, reason string
		if hasTimeFields {
			// 14个字段：items在索引11，reason在索引12
			items = strings.TrimSpace(record[11])
			reason = strings.TrimSpace(record[12])
		} else {
			// 12个字段：items在索引9，reason在索引10
			items = strings.TrimSpace(record[9])
			reason = strings.TrimSpace(record[10])
		}

		// 创建条目
		entry := MailLogEntry{
			SID:      sid,
			GID:      gid,
			UID:      uid,
			RoleID:   roleid,
			RoleName: rolename,
			IsGM:     isgm,
			Caty:     caty,
			Title:    title,
			Content:  content,
			Rank:     rank,
			Score:    score,
			SendTime: sendtime,
			OpTime:   optime,
			Items:    items,
			Reason:   reason,
			Op:       op,
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// WriteMailLogToExcel 将邮件日志数据写入Excel文件
// 如果工作表已存在，将追加数据而不是覆盖
func WriteMailLogToExcel(f *excelize.File, sheetName string, entries []MailLogEntry) {
	// 检查工作表是否存在
	sheetExists := worksheetExists(f, sheetName)
	var startRow int

	if !sheetExists {
		// 创建工作表
		f.NewSheet(sheetName)

		// 设置表头：16列（添加了Rank和Score）
		headers := []string{"SID", "GID", "UID", "RoleID", "RoleName", "IsGM", "Caty", "Title", "Content", "Rank", "Score", "SendTime", "OpTime", "Items", "Reason", "Op"}
		for i, header := range headers {
			cellName := fmt.Sprintf("%c1", 'A'+i)
			f.SetCellValue(sheetName, cellName, header)
		}

		// 设置列宽
		f.SetColWidth(sheetName, "A", "A", 15) // SID
		f.SetColWidth(sheetName, "B", "B", 10) // GID
		f.SetColWidth(sheetName, "C", "C", 20) // UID
		f.SetColWidth(sheetName, "D", "D", 15) // RoleID
		f.SetColWidth(sheetName, "E", "E", 30) // RoleName
		f.SetColWidth(sheetName, "F", "F", 8)  // IsGM
		f.SetColWidth(sheetName, "G", "G", 15) // Caty
		f.SetColWidth(sheetName, "H", "H", 30) // Title
		f.SetColWidth(sheetName, "I", "I", 50) // Content
		f.SetColWidth(sheetName, "J", "J", 10) // Rank
		f.SetColWidth(sheetName, "K", "K", 15) // Score
		f.SetColWidth(sheetName, "L", "L", 15) // SendTime
		f.SetColWidth(sheetName, "M", "M", 15) // OpTime
		f.SetColWidth(sheetName, "N", "N", 50) // Items
		f.SetColWidth(sheetName, "O", "O", 20) // Reason
		f.SetColWidth(sheetName, "P", "P", 8)  // Op

		startRow = 2
	} else {
		// 工作表已存在，找到最后一行
		checkColumns := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P"}
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
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), entry.SID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), entry.GID)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), entry.UID)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), entry.RoleID)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), entry.RoleName)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), entry.IsGM)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), entry.Caty)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), entry.Title)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowNum), entry.Content)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowNum), entry.Rank)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowNum), entry.Score)
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", rowNum), entry.SendTime)
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", rowNum), entry.OpTime)
		f.SetCellValue(sheetName, fmt.Sprintf("N%d", rowNum), entry.Items)
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", rowNum), entry.Reason)
		f.SetCellValue(sheetName, fmt.Sprintf("P%d", rowNum), entry.Op)
		rowNum++
	}
}

// ProcessMailLogFile 处理邮件日志文件的完整流程：读取、写入Excel
func ProcessMailLogFile(filePath string, excelFile *excelize.File, sheetName string) error {
	// 读取数据
	entries, err := ReadMailLogData(filePath)
	if err != nil {
		return fmt.Errorf("failed to read mail log data: %w", err)
	}

	// 写入Excel
	WriteMailLogToExcel(excelFile, sheetName, entries)

	return nil
}
