package analytics

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Statistics 统计数据
type Statistics struct {
	TotalRecords      int                    `json:"totalRecords"`
	ColumnStats       map[string]ColumnStat  `json:"columnStats"`
	Distribution      map[string]interface{} `json:"distribution"`
	TopRecords        []map[string]interface{} `json:"topRecords"`
	TimeSeries        []TimeSeriesPoint      `json:"timeSeries,omitempty"`
	IsRankingData     bool                    `json:"isRankingData"`     // 是否为排行榜数据
	RankingAnalysis   *RankingAnalysis        `json:"rankingAnalysis,omitempty"` // 排行榜专用分析
}

// RankingAnalysis 排行榜专用分析数据
type RankingAnalysis struct {
	ByGroup        map[string]GroupStats    `json:"byGroup"`        // 按组别统计
	ByDay          map[string]DayStats     `json:"byDay,omitempty"` // 按日期统计（仅日榜）
	RankDistribution []RankBucket          `json:"rankDistribution"` // 排名分布
	ScoreDistribution map[string]interface{} `json:"scoreDistribution"` // 分数分布
	TopByGroup     map[string][]TopPlayer  `json:"topByGroup"`     // 各组别Top玩家
	TopByGroupDay  map[string]map[string][]TopPlayer `json:"topByGroupDay,omitempty"` // 按组别+日期分组的Top玩家
}

// GroupStats 组别统计
type GroupStats struct {
	TotalPlayers   int     `json:"totalPlayers"`
	AvgRank        float64 `json:"avgRank"`
	AvgScore       float64 `json:"avgScore"`
	MaxScore       float64 `json:"maxScore"`
	MinScore       float64 `json:"minScore"`
	TopRank        int     `json:"topRank"`
}

// DayStats 日期统计
type DayStats struct {
	TotalPlayers int     `json:"totalPlayers"`
	AvgScore      float64 `json:"avgScore"`
	MaxScore      float64 `json:"maxScore"`
}

// RankBucket 排名区间
type RankBucket struct {
	Range  string `json:"range"`  // 如 "1-10", "11-50"
	Count  int    `json:"count"`
	Label  string `json:"label"`
}

// TopPlayer Top玩家信息
type TopPlayer struct {
	Member string  `json:"member"`
	Rank   int     `json:"rank"`
	Score  float64 `json:"score"`
	Group  string  `json:"group,omitempty"`
	Day    string  `json:"day,omitempty"`
}

// ColumnStat 列统计信息
type ColumnStat struct {
	Type          string      `json:"type"`          // "number", "string", "date"
	Min           interface{} `json:"min,omitempty"`
	Max           interface{} `json:"max,omitempty"`
	Avg           interface{} `json:"avg,omitempty"`
	Sum           interface{} `json:"sum,omitempty"`
	Count         int         `json:"count"`
	UniqueValues  int         `json:"uniqueValues,omitempty"`
	NullCount     int         `json:"nullCount"`
}

// TimeSeriesPoint 时间序列数据点
type TimeSeriesPoint struct {
	Time  string      `json:"time"`
	Value interface{} `json:"value"`
	Label string      `json:"label,omitempty"`
}

// AnalyzeExcelFile 分析Excel文件
func AnalyzeExcelFile(excelPath string, sheetName string) (*Statistics, error) {
	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	// 如果没有指定工作表，使用第一个
	if sheetName == "" {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, fmt.Errorf("no sheets found")
		}
		sheetName = sheets[0]
	}

	// 读取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet: %w", err)
	}

	if len(rows) == 0 {
		return &Statistics{
			TotalRecords: 0,
			ColumnStats:  make(map[string]ColumnStat),
		}, nil
	}

	// 第一行是表头
	columns := rows[0]
	dataRows := rows[1:]

	stats := &Statistics{
		TotalRecords: len(dataRows),
		ColumnStats:  make(map[string]ColumnStat),
		Distribution: make(map[string]interface{}),
		TopRecords:   make([]map[string]interface{}, 0),
	}

	// 初始化列统计
	columnData := make(map[string][]interface{})
	columnTypes := make(map[string]string)

	for _, col := range columns {
		columnData[col] = make([]interface{}, 0)
		columnTypes[col] = "unknown"
	}

	// 收集数据
	for _, row := range dataRows {
		for i, col := range columns {
			var value interface{} = ""
			if i < len(row) {
				value = strings.TrimSpace(row[i])
			}
			columnData[col] = append(columnData[col], value)
		}
	}

	// 分析每列
	for _, col := range columns {
		data := columnData[col]
		colStat := analyzeColumn(col, data)
		stats.ColumnStats[col] = colStat
	}

	// 生成分布数据（针对数值列）
	for col, stat := range stats.ColumnStats {
		if stat.Type == "number" {
			stats.Distribution[col] = generateDistribution(columnData[col], stat)
		}
	}

	// 生成Top记录（针对排名相关列）
	stats.TopRecords = generateTopRecords(columns, dataRows, stats.ColumnStats)

	// 生成时间序列（如果有时间列）
	stats.TimeSeries = generateTimeSeries(columns, dataRows, stats.ColumnStats)

	// 检测是否为排行榜数据并生成专门分析
	stats.IsRankingData = isRankingData(columns)
	if stats.IsRankingData {
		stats.RankingAnalysis = analyzeRankingData(columns, dataRows, stats.ColumnStats)
	}

	return stats, nil
}

// isRankingData 检测是否为排行榜数据
func isRankingData(columns []string) bool {
	hasGroup := false
	hasRank := false
	hasScore := false
	
	for _, col := range columns {
		colLower := strings.ToLower(col)
		if strings.Contains(colLower, "组别") || strings.Contains(colLower, "group") {
			hasGroup = true
		}
		if strings.Contains(colLower, "rank") || strings.Contains(colLower, "排名") {
			hasRank = true
		}
		if strings.Contains(colLower, "score") || strings.Contains(colLower, "分数") || strings.Contains(colLower, "积分") {
			hasScore = true
		}
	}
	
	return hasGroup && hasRank && hasScore
}

// analyzeRankingData 分析排行榜数据
func analyzeRankingData(columns []string, dataRows [][]string, columnStats map[string]ColumnStat) *RankingAnalysis {
	analysis := &RankingAnalysis{
		ByGroup:           make(map[string]GroupStats),
		ByDay:             make(map[string]DayStats),
		RankDistribution:  make([]RankBucket, 0),
		ScoreDistribution: make(map[string]interface{}),
		TopByGroup:        make(map[string][]TopPlayer),
		TopByGroupDay:     make(map[string]map[string][]TopPlayer),
	}

	// 找到列索引
	groupIdx := -1
	rankIdx := -1
	scoreIdx := -1
	dayIdx := -1
	memberIdx := -1

	for i, col := range columns {
		colLower := strings.ToLower(col)
		if strings.Contains(colLower, "组别") || strings.Contains(colLower, "group") {
			groupIdx = i
		}
		if (strings.Contains(colLower, "rank") || strings.Contains(colLower, "排名")) && rankIdx == -1 {
			rankIdx = i
		}
		if (strings.Contains(colLower, "processed") && strings.Contains(colLower, "score")) || 
		   (strings.Contains(colLower, "处理后") && strings.Contains(colLower, "分数")) {
			scoreIdx = i
		}
		if strings.Contains(colLower, "day") || strings.Contains(colLower, "日期") {
			dayIdx = i
		}
		if strings.Contains(colLower, "member") || strings.Contains(colLower, "成员") {
			memberIdx = i
		}
	}

	if groupIdx < 0 || rankIdx < 0 || scoreIdx < 0 {
		return analysis
	}

	// 按组别统计
	groupData := make(map[string][]struct {
		rank  int
		score float64
		day   string
		member string
	})

	ranks := make([]int, 0)
	scores := make([]float64, 0)

	for _, row := range dataRows {
		if groupIdx >= len(row) || rankIdx >= len(row) || scoreIdx >= len(row) {
			continue
		}

		group := strings.TrimSpace(row[groupIdx])
		rankStr := strings.TrimSpace(row[rankIdx])
		scoreStr := strings.TrimSpace(row[scoreIdx])

		if group == "" || rankStr == "" || scoreStr == "" {
			continue
		}

		rank, err1 := strconv.Atoi(rankStr)
		score, err2 := strconv.ParseFloat(scoreStr, 64)

		if err1 != nil || err2 != nil {
			continue
		}

		day := ""
		if dayIdx >= 0 && dayIdx < len(row) {
			day = strings.TrimSpace(row[dayIdx])
		}

		member := ""
		if memberIdx >= 0 && memberIdx < len(row) {
			member = strings.TrimSpace(row[memberIdx])
		}

		groupData[group] = append(groupData[group], struct {
			rank  int
			score float64
			day   string
			member string
		}{rank, score, day, member})

		ranks = append(ranks, rank)
		scores = append(scores, score)
	}

	// 计算组别统计
	for group, data := range groupData {
		totalPlayers := len(data)
		if totalPlayers == 0 {
			continue
		}

		sumRank := 0
		sumScore := 0.0
		maxScore := data[0].score
		minScore := data[0].score
		topRank := data[0].rank

		for _, d := range data {
			sumRank += d.rank
			sumScore += d.score
			if d.score > maxScore {
				maxScore = d.score
			}
			if d.score < minScore {
				minScore = d.score
			}
			if d.rank < topRank {
				topRank = d.rank
			}
		}

		analysis.ByGroup[group] = GroupStats{
			TotalPlayers: totalPlayers,
			AvgRank:      float64(sumRank) / float64(totalPlayers),
			AvgScore:     sumScore / float64(totalPlayers),
			MaxScore:     maxScore,
			MinScore:     minScore,
			TopRank:      topRank,
		}

		// 按日期统计（如果有日期列）
		if dayIdx >= 0 {
			dayData := make(map[string][]float64)
			for _, d := range data {
				if d.day != "" {
					dayData[d.day] = append(dayData[d.day], d.score)
				}
			}

			for day, scores := range dayData {
				sum := 0.0
				max := scores[0]
				for _, s := range scores {
					sum += s
					if s > max {
						max = s
					}
				}
				analysis.ByDay[day] = DayStats{
					TotalPlayers: len(scores),
					AvgScore:     sum / float64(len(scores)),
					MaxScore:     max,
				}
			}
		}

		// Top 10 玩家（按组别）
		sortedData := make([]struct {
			rank  int
			score float64
			day   string
			member string
		}, len(data))
		copy(sortedData, data)
		sort.Slice(sortedData, func(i, j int) bool {
			return sortedData[i].rank < sortedData[j].rank
		})

		topCount := 10
		if len(sortedData) < topCount {
			topCount = len(sortedData)
		}

		topPlayers := make([]TopPlayer, topCount)
		for i := 0; i < topCount; i++ {
			topPlayers[i] = TopPlayer{
				Member: sortedData[i].member,
				Rank:   sortedData[i].rank,
				Score:  sortedData[i].score,
				Group:  group,
				Day:    sortedData[i].day,
			}
		}
		analysis.TopByGroup[group] = topPlayers

		// 按组别+日期分组Top玩家（如果有日期列）
		if dayIdx >= 0 {
			groupDayData := make(map[string][]struct {
				rank  int
				score float64
				member string
			})

			for _, d := range data {
				if d.day != "" {
					groupDayData[d.day] = append(groupDayData[d.day], struct {
						rank  int
						score float64
						member string
					}{d.rank, d.score, d.member})
				}
			}

			if len(groupDayData) > 0 {
				analysis.TopByGroupDay[group] = make(map[string][]TopPlayer)
				for day, dayData := range groupDayData {
					sort.Slice(dayData, func(i, j int) bool {
						return dayData[i].rank < dayData[j].rank
					})

					dayTopCount := 10
					if len(dayData) < dayTopCount {
						dayTopCount = len(dayData)
					}

					dayTopPlayers := make([]TopPlayer, dayTopCount)
					for i := 0; i < dayTopCount; i++ {
						dayTopPlayers[i] = TopPlayer{
							Member: dayData[i].member,
							Rank:   dayData[i].rank,
							Score:  dayData[i].score,
							Group:  group,
							Day:    day,
						}
					}
					analysis.TopByGroupDay[group][day] = dayTopPlayers
				}
			}
		}
	}

	// 排名分布
	if len(ranks) > 0 {
		sort.Ints(ranks)
		maxRank := ranks[len(ranks)-1]
		
		buckets := []struct {
			label string
			min   int
			max   int
		}{
			{"1-10", 1, 10},
			{"11-50", 11, 50},
			{"51-100", 51, 100},
			{"101-500", 101, 500},
			{"501-1000", 501, 1000},
			{"1000+", 1001, maxRank + 1},
		}

		for _, bucket := range buckets {
			count := 0
			for _, rank := range ranks {
				if rank >= bucket.min && rank < bucket.max {
					count++
				}
			}
			if count > 0 {
				analysis.RankDistribution = append(analysis.RankDistribution, RankBucket{
					Range: bucket.label,
					Count: count,
					Label: bucket.label,
				})
			}
		}
	}

	// 分数分布（增加粒度，从10个增加到25个）
	if len(scores) > 0 {
		sort.Float64s(scores)
		minScore := scores[0]
		maxScore := scores[len(scores)-1]
		
		bucketCount := 25 // 增加到25个bucket，提高粒度
		bucketSize := (maxScore - minScore) / float64(bucketCount)
		if bucketSize == 0 {
			bucketSize = 1
		}

		bucketLabels := make([]string, bucketCount)
		bucketData := make([]int, bucketCount)

		for i := 0; i < bucketCount; i++ {
			start := minScore + float64(i)*bucketSize
			end := minScore + float64(i+1)*bucketSize
			if i == bucketCount-1 {
				end = maxScore + 1
			}
			// 根据分数范围调整显示精度
			if maxScore-minScore > 1000 {
				bucketLabels[i] = fmt.Sprintf("%.0f-%.0f", start, end)
			} else if maxScore-minScore > 100 {
				bucketLabels[i] = fmt.Sprintf("%.1f-%.1f", start, end)
			} else {
				bucketLabels[i] = fmt.Sprintf("%.2f-%.2f", start, end)
			}

			for _, score := range scores {
				if score >= start && (score < end || (i == bucketCount-1 && score <= end)) {
					bucketData[i]++
				}
			}
		}

		analysis.ScoreDistribution = map[string]interface{}{
			"labels": bucketLabels,
			"data":   bucketData,
		}
	}

	return analysis
}

// analyzeColumn 分析单列数据
func analyzeColumn(colName string, data []interface{}) ColumnStat {
	stat := ColumnStat{
		Type:         "string",
		Count:        len(data),
		NullCount:    0,
		UniqueValues: 0,
	}

	// 检测列类型
	isNumeric := true
	isTime := false
	numbers := make([]float64, 0)
	uniqueMap := make(map[string]bool)
	nonEmptyCount := 0

	for _, val := range data {
		strVal := fmt.Sprintf("%v", val)
		if strVal == "" {
			stat.NullCount++
			continue
		}
		nonEmptyCount++

		// 检查唯一值
		uniqueMap[strVal] = true

		// 尝试解析为数字
		if isNumeric {
			num, err := strconv.ParseFloat(strVal, 64)
			if err != nil {
				isNumeric = false
			} else {
				numbers = append(numbers, num)
			}
		}

		// 检查是否为时间戳（10位或13位数字）
		if !isTime && len(strVal) >= 10 {
			if num, err := strconv.ParseInt(strVal, 10, 64); err == nil {
				if (num > 1000000000 && num < 9999999999) || (num > 1000000000000 && num < 9999999999999) {
					isTime = true
				}
			}
		}
	}

	stat.UniqueValues = len(uniqueMap)

	if isTime {
		stat.Type = "date"
	} else if isNumeric && len(numbers) > 0 {
		stat.Type = "number"
		sort.Float64s(numbers)
		stat.Min = numbers[0]
		stat.Max = numbers[len(numbers)-1]

		// 计算平均值
		sum := 0.0
		for _, n := range numbers {
			sum += n
		}
		stat.Avg = sum / float64(len(numbers))
		stat.Sum = sum
	} else {
		stat.Type = "string"
	}

	return stat
}

// generateDistribution 生成分布数据（直方图）
func generateDistribution(data []interface{}, stat ColumnStat) map[string]interface{} {
	if stat.Type != "number" {
		return nil
	}

	numbers := make([]float64, 0)
	for _, val := range data {
		if strVal := fmt.Sprintf("%v", val); strVal != "" {
			if num, err := strconv.ParseFloat(strVal, 64); err == nil {
				numbers = append(numbers, num)
			}
		}
	}

	if len(numbers) == 0 {
		return nil
	}

	// 计算分桶（10个桶）
	min := stat.Min.(float64)
	max := stat.Max.(float64)
	bucketSize := (max - min) / 10
	if bucketSize == 0 {
		bucketSize = 1
	}

	buckets := make([]int, 10)
	labels := make([]string, 10)

	for i := 0; i < 10; i++ {
		start := min + float64(i)*bucketSize
		end := min + float64(i+1)*bucketSize
		if i == 9 {
			end = max + 1 // 最后一个桶包含最大值
		}
		labels[i] = fmt.Sprintf("%.2f-%.2f", start, end)

		for _, num := range numbers {
			if num >= start && (num < end || (i == 9 && num <= end)) {
				buckets[i]++
			}
		}
	}

	return map[string]interface{}{
		"labels": labels,
		"data":   buckets,
	}
}

// generateTopRecords 生成Top记录
func generateTopRecords(columns []string, rows [][]string, columnStats map[string]ColumnStat) []map[string]interface{} {
	// 查找排名或分数相关的列
	var rankCol, scoreCol string
	for col, stat := range columnStats {
		colLower := strings.ToLower(col)
		if (strings.Contains(colLower, "rank") || strings.Contains(colLower, "排名")) && stat.Type == "number" {
			rankCol = col
		}
		if (strings.Contains(colLower, "score") || strings.Contains(colLower, "积分") || strings.Contains(colLower, "分数")) && stat.Type == "number" {
			scoreCol = col
		}
	}

	if rankCol == "" && scoreCol == "" {
		return nil
	}

	// 获取列索引
	rankIdx := -1
	scoreIdx := -1
	for i, col := range columns {
		if col == rankCol {
			rankIdx = i
		}
		if col == scoreCol {
			scoreIdx = i
		}
	}

	// 创建记录列表
	type record struct {
		row   []string
		value float64
	}

	records := make([]record, 0)
	for _, row := range rows {
		var value float64
		if scoreIdx >= 0 && scoreIdx < len(row) {
			if num, err := strconv.ParseFloat(strings.TrimSpace(row[scoreIdx]), 64); err == nil {
				value = num
			}
		} else if rankIdx >= 0 && rankIdx < len(row) {
			if num, err := strconv.ParseFloat(strings.TrimSpace(row[rankIdx]), 64); err == nil {
				value = -num // 排名越小越好，所以取负值
			}
		} else {
			continue
		}

		records = append(records, record{row: row, value: value})
	}

	// 排序（按分数降序或排名升序）
	sort.Slice(records, func(i, j int) bool {
		if scoreIdx >= 0 {
			return records[i].value > records[j].value
		}
		return records[i].value < records[j].value
	})

	// 取前10条
	topCount := 10
	if len(records) < topCount {
		topCount = len(records)
	}

	topRecords := make([]map[string]interface{}, topCount)
	for i := 0; i < topCount; i++ {
		recordMap := make(map[string]interface{})
		for j, col := range columns {
			if j < len(records[i].row) {
				recordMap[col] = records[i].row[j]
			} else {
				recordMap[col] = ""
			}
		}
		topRecords[i] = recordMap
	}

	return topRecords
}

// generateTimeSeries 生成时间序列数据
func generateTimeSeries(columns []string, rows [][]string, columnStats map[string]ColumnStat) []TimeSeriesPoint {
	// 查找时间相关的列
	var timeCol, valueCol string
	for col, stat := range columnStats {
		colLower := strings.ToLower(col)
		if stat.Type == "date" || strings.Contains(colLower, "time") || strings.Contains(colLower, "时间") {
			timeCol = col
			break
		}
	}

	// 查找数值列作为值
	for col, stat := range columnStats {
		if stat.Type == "number" && (strings.Contains(strings.ToLower(col), "score") || strings.Contains(strings.ToLower(col), "积分") || strings.Contains(strings.ToLower(col), "分数")) {
			valueCol = col
			break
		}
	}

	if timeCol == "" || valueCol == "" {
		return nil
	}

	// 获取列索引
	timeIdx := -1
	valueIdx := -1
	for i, col := range columns {
		if col == timeCol {
			timeIdx = i
		}
		if col == valueCol {
			valueIdx = i
		}
	}

	if timeIdx < 0 || valueIdx < 0 {
		return nil
	}

	// 收集时间序列数据
	timeSeriesMap := make(map[string]float64)
	timeSeriesCount := make(map[string]int)

	for _, row := range rows {
		if timeIdx >= len(row) || valueIdx >= len(row) {
			continue
		}

		timeStr := strings.TrimSpace(row[timeIdx])
		valueStr := strings.TrimSpace(row[valueIdx])

		if timeStr == "" || valueStr == "" {
			continue
		}

		// 解析时间戳
		var timestamp int64
		if ts, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
			timestamp = ts
			// 如果是13位时间戳，转换为10位
			if timestamp > 9999999999 {
				timestamp = timestamp / 1000
			}
		} else {
			continue
		}

		// 解析值
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			continue
		}

		// 按天聚合（简化处理，使用时间戳的日期部分）
		dateKey := fmt.Sprintf("%d", timestamp/86400*86400) // 按天聚合

		timeSeriesMap[dateKey] += value
		timeSeriesCount[dateKey]++
	}

	// 转换为时间序列点
	timeSeries := make([]TimeSeriesPoint, 0, len(timeSeriesMap))
	for timeKey, totalValue := range timeSeriesMap {
		count := timeSeriesCount[timeKey]
		avgValue := totalValue / float64(count)

		// 解析时间戳为可读格式
		timestamp, _ := strconv.ParseInt(timeKey, 10, 64)
		timeLabel := formatTimestamp(timestamp)

		timeSeries = append(timeSeries, TimeSeriesPoint{
			Time:  timeKey,
			Value: math.Round(avgValue*100) / 100, // 保留2位小数
			Label: timeLabel,
		})
	}

	// 按时间排序
	sort.Slice(timeSeries, func(i, j int) bool {
		ti, _ := strconv.ParseInt(timeSeries[i].Time, 10, 64)
		tj, _ := strconv.ParseInt(timeSeries[j].Time, 10, 64)
		return ti < tj
	})

	return timeSeries
}

// formatTimestamp 格式化时间戳
func formatTimestamp(ts int64) string {
	// 简化处理，返回时间戳字符串
	// 实际应用中可以使用 time 包格式化
	return fmt.Sprintf("%d", ts)
}
