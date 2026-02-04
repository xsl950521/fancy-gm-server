package missinglist

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xuri/excelize/v2"
)

var (
	// httpTransport 全局HTTP传输层，使用连接池和Keep-Alive（共享连接池）
	httpTransportOnce sync.Once
	httpTransport     *http.Transport
)

// getHTTPClient 获取优化的HTTP客户端（连接池、Keep-Alive）
func getHTTPClient(timeoutSec int) *http.Client {
	// 初始化共享的Transport（连接池）
	httpTransportOnce.Do(func() {
		httpTransport = &http.Transport{
			MaxIdleConns:        200,              // 最大空闲连接数
			MaxIdleConnsPerHost: 100,               // 每个主机的最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时
			DisableKeepAlives:   false,            // 启用Keep-Alive
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,  // 连接超时（减少到3秒）
				KeepAlive: 30 * time.Second,  // Keep-Alive时间
			}).DialContext,
			TLSHandshakeTimeout:   3 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ForceAttemptHTTP2:     false, // 禁用HTTP/2，使用HTTP/1.1 Keep-Alive
		}
	})
	
	// 为每个请求创建新的Client，但共享Transport（连接池）
	timeout := time.Duration(timeoutSec) * time.Second
	if timeout < 1*time.Second {
		timeout = 5 * time.Second
	}
	
	return &http.Client{
		Transport: httpTransport,
		Timeout:   timeout,
	}
}

// getHTTPClientNoReuse 获取不使用连接复用的HTTP客户端（用于连续请求）
func getHTTPClientNoReuse(timeoutSec int) *http.Client {
	// 创建一个独立的Transport，禁用Keep-Alive
	noReuseTransport := &http.Transport{
		MaxIdleConns:        0,   // 不保留空闲连接
		MaxIdleConnsPerHost: 0,   // 每个主机不保留空闲连接
		DisableKeepAlives:   true, // 禁用Keep-Alive，每次请求都使用新连接
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: -1, // 禁用Keep-Alive
		}).DialContext,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     false,
	}
	
	timeout := time.Duration(timeoutSec) * time.Second
	if timeout < 1*time.Second {
		timeout = 5 * time.Second
	}
	
	return &http.Client{
		Transport: noReuseTransport,
		Timeout:   timeout,
	}
}

// MailData 邮件数据
type MailData struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	RoleIDList []int64  `json:"roleid_list"`
	PropList   []Prop   `json:"prop_list"`
	Group      int      `json:"group"`
}

// Prop 道具信息
type Prop struct {
	ID    int `json:"id"`
	Count int `json:"count"`
}

// ResendRequest 补发请求
type ResendRequest struct {
	URL      string   `json:"url"`
	MailData MailData `json:"mail_data"`
}

// ResendResponse 补发响应
type ResendResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SendResendRequest 发送补发请求
func SendResendRequest(url string, mailData MailData) (*ResendResponse, error) {
	// 构建请求体
	requestBody, err := json.Marshal(mailData)
	if err != nil {
		return nil, fmt.Errorf("序列化邮件数据失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var resendResp ResendResponse
	if err := json.Unmarshal(body, &resendResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &resendResp, fmt.Errorf("HTTP错误: %d, 消息: %s", resp.StatusCode, resendResp.Message)
	}

	return &resendResp, nil
}

// PlayerInfo 玩家信息
type PlayerInfo struct {
	PSID      int64  `json:"psid"`       // 玩家ID
	Rank      int    `json:"rank"`       // 排名
	Score     int64  `json:"score"`      // 积分（可选）
	ClubID    *int   `json:"clubid,omitempty"` // 公会ID（可选）
	Username  string `json:"username,omitempty"` // 玩家名称（可选）
	UserID    string `json:"userid,omitempty"`   // 玩家UID（可选）
}

// ResendConfig 补发配置
type ResendConfig struct {
	Title       string            `json:"title"`        // 邮件标题
	Content     string            `json:"content"`     // 邮件内容
	Group       int               `json:"group"`        // 组别
	RewardConfig map[string]map[string]int `json:"reward_config"` // 奖励配置：{"排名上限": {"道具ID": 数量}}
}

// ParseExcelFile 解析Excel文件（Base64编码）
func ParseExcelFile(base64Data string) ([]PlayerInfo, error) {
	// 解码Base64
	fileData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("Base64解码失败: %v", err)
	}

	// 创建临时文件或使用内存
	f, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("打开Excel文件失败: %v", err)
	}
	defer f.Close()

	// 获取第一个工作表
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("Excel文件中没有工作表")
	}
	sheetName := sheetList[0]

	// 读取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %v", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel文件数据不足（至少需要表头和数据行）")
	}

	// 第一行是表头，查找PSID、Rank、Score列
	headers := rows[0]
	psidCol := -1
	rankCol := -1
	scoreCol := -1

	for i, header := range headers {
		headerLower := strings.ToLower(strings.TrimSpace(header))
		if headerLower == "psid" || headerLower == "player_id" || headerLower == "玩家id" {
			psidCol = i
		} else if headerLower == "rank" || headerLower == "排名" {
			rankCol = i
		} else if headerLower == "score" || headerLower == "积分" {
			scoreCol = i
		}
	}

	if psidCol == -1 {
		return nil, fmt.Errorf("未找到PSID列（请确保表头包含PSID、player_id或玩家ID）")
	}
	if rankCol == -1 {
		return nil, fmt.Errorf("未找到Rank列（请确保表头包含Rank或排名）")
	}

	// 解析数据行
	players := make([]PlayerInfo, 0)
	for _, row := range rows[1:] {
		if len(row) <= psidCol || len(row) <= rankCol {
			continue
		}

		// 解析PSID
		psidStr := strings.TrimSpace(row[psidCol])
		psid, err := strconv.ParseInt(psidStr, 10, 64)
		if err != nil {
			continue // 跳过无效行
		}

		// 解析Rank
		rankStr := strings.TrimSpace(row[rankCol])
		rank, err := strconv.Atoi(rankStr)
		if err != nil {
			continue // 跳过无效行
		}

		// 解析Score（可选）
		var score int64 = 0
		if scoreCol != -1 && len(row) > scoreCol {
			scoreStr := strings.TrimSpace(row[scoreCol])
			if scoreStr != "" {
				if s, err := strconv.ParseInt(scoreStr, 10, 64); err == nil {
					score = s
				}
			}
		}

		players = append(players, PlayerInfo{
			PSID:  psid,
			Rank:  rank,
			Score: score,
		})
	}

	return players, nil
}

// MatchReward 根据排名匹配奖励配置
func MatchReward(rank int, rewardConfig map[string]map[string]int) map[string]int {
	if len(rewardConfig) == 0 {
		return nil
	}

	// 将排名上限转换为整数并排序
	rankLimits := make([]int, 0, len(rewardConfig))
	for k := range rewardConfig {
		limit, err := strconv.Atoi(k)
		if err == nil {
			rankLimits = append(rankLimits, limit)
		}
	}

	if len(rankLimits) == 0 {
		return nil
	}

	sort.Ints(rankLimits)

	// 找到第一个排名上限 >= 玩家排名的配置
	for _, limit := range rankLimits {
		if rank <= limit {
			return rewardConfig[strconv.Itoa(limit)]
		}
	}

	// 如果没有匹配的，返回最后一个配置（最大排名上限）
	maxLimit := rankLimits[len(rankLimits)-1]
	return rewardConfig[strconv.Itoa(maxLimit)]
}

// ProcessResend 处理补发请求
func ProcessResend(url string, players []PlayerInfo, config ResendConfig) ([]ResendResult, error) {
	results := make([]ResendResult, 0, len(players))

	for _, player := range players {
		// 匹配奖励配置
		reward := MatchReward(player.Rank, config.RewardConfig)
		if reward == nil {
			results = append(results, ResendResult{
				PSID:    player.PSID,
				Rank:    player.Rank,
				Success: false,
				Message: "未找到匹配的奖励配置",
			})
			continue
		}

		// 构建道具列表
		propList := make([]Prop, 0, len(reward))
		for propIDStr, count := range reward {
			propID, err := strconv.Atoi(propIDStr)
			if err != nil {
				continue
			}
			propList = append(propList, Prop{
				ID:    propID,
				Count: count,
			})
		}

		// 格式化邮件标题和内容（替换%s占位符）
		title := strings.ReplaceAll(config.Title, "%s", strconv.Itoa(player.Rank))
		title = strings.ReplaceAll(title, "%s", strconv.FormatInt(player.Score, 10))
		content := strings.ReplaceAll(config.Content, "%s", strconv.Itoa(player.Rank))
		content = strings.ReplaceAll(content, "%s", strconv.FormatInt(player.Score, 10))

		// 构建邮件数据
		mailData := MailData{
			Title:      title,
			Content:    content,
			RoleIDList: []int64{player.PSID},
			PropList:   propList,
			Group:      config.Group,
		}

		// 发送补发请求
		resp, err := SendResendRequest(url, mailData)
		if err != nil {
			results = append(results, ResendResult{
				PSID:    player.PSID,
				Rank:    player.Rank,
				Success: false,
				Message: err.Error(),
			})
			continue
		}

		results = append(results, ResendResult{
			PSID:    player.PSID,
			Rank:    player.Rank,
			Success: resp.Code == 0,
			Message: resp.Message,
		})
	}

	return results, nil
}

// ResendResult 补发结果
type ResendResult struct {
	PSID    int64  `json:"psid"`
	Rank    int    `json:"rank"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ProcessResendV2 处理补发请求（新版本，使用x-www-form-urlencoded格式）
func ProcessResendV2(apiURL string, rankType string, group string, timestamp string, sign string, title string, content string, rewardConfig map[string]map[string]int, players []PlayerInfo, rankWorkers int, httpTimeoutSec int) ([]ResendResult, error) {
	// 如果group是"all"，需要特殊处理（所有玩家一起发送，roleid_list为"lv0-lv9"）
	if group == "all" {
		// 按排名分组，每个排名发送一次
		playersByRank := make(map[int][]PlayerInfo)
		for _, player := range players {
			playersByRank[player.Rank] = append(playersByRank[player.Rank], player)
		}

		ranks := make([]int, 0, len(playersByRank))
		for rank := range playersByRank {
			ranks = append(ranks, rank)
		}
		sort.Ints(ranks)

		return processRanksConcurrently(apiURL, rankType, group, timestamp, sign, title, content, rewardConfig, ranks, playersByRank, rankWorkers, httpTimeoutSec)
	}

	// 其他group类型：按排名分组玩家，相同排名的玩家一起发送
	playersByRank := make(map[int][]PlayerInfo)
	for _, player := range players {
		playersByRank[player.Rank] = append(playersByRank[player.Rank], player)
	}

	// 对排名排序
	ranks := make([]int, 0, len(playersByRank))
	for rank := range playersByRank {
		ranks = append(ranks, rank)
	}
	sort.Ints(ranks)

	return processRanksConcurrently(apiURL, rankType, group, timestamp, sign, title, content, rewardConfig, ranks, playersByRank, rankWorkers, httpTimeoutSec)
}

func processRanksConcurrently(apiURL string, rankType string, group string, timestamp string, sign string, title string, content string, rewardConfig map[string]map[string]int, ranks []int, playersByRank map[int][]PlayerInfo, rankWorkers int, httpTimeoutSec int) ([]ResendResult, error) {
	results := make([]ResendResult, 0, len(ranks))
	if len(ranks) == 0 {
		return results, nil
	}

	workerCount := rankWorkers
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > len(ranks) {
		workerCount = len(ranks)
	}

	rankCh := make(chan int)
	resultCh := make(chan []ResendResult, len(ranks))
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
			go func() {
				defer wg.Done()
				for rank := range rankCh {
					rankPlayers := playersByRank[rank]
					resultCh <- processRankResend(apiURL, rankType, group, timestamp, sign, title, content, rewardConfig, rank, rankPlayers, httpTimeoutSec)
				}
			}()
	}

	go func() {
		for _, rank := range ranks {
			rankCh <- rank
		}
		close(rankCh)
		wg.Wait()
		close(resultCh)
	}()

	for res := range resultCh {
		results = append(results, res...)
	}
	return results, nil
}

// processRankResend 处理单个排名的补发
func processRankResend(apiURL string, rankType string, group string, timestamp string, sign string, title string, content string, rewardConfig map[string]map[string]int, rank int, rankPlayers []PlayerInfo, httpTimeoutSec int) []ResendResult {
	results := make([]ResendResult, 0, len(rankPlayers))
		
	// 匹配奖励配置
	reward := MatchReward(rank, rewardConfig)
	if reward == nil {
		for _, player := range rankPlayers {
			results = append(results, ResendResult{
				PSID:    player.PSID,
				Rank:    player.Rank,
				Success: false,
				Message: "未找到匹配的奖励配置",
			})
		}
		return results
	}

	// 构建道具列表字符串（格式：itemid|itemnum,itemid|itemnum）
	propListParts := make([]string, 0, len(reward))
	for propIDStr, count := range reward {
		propListParts = append(propListParts, fmt.Sprintf("%s|%d", propIDStr, count))
	}
	propList := strings.Join(propListParts, ",")

	// 格式化邮件标题和内容（替换%s占位符）
	formattedTitle := strings.ReplaceAll(title, "%s", strconv.Itoa(rank))
	formattedTitle = strings.ReplaceAll(formattedTitle, "%s", strconv.FormatInt(rankPlayers[0].Score, 10))
	formattedContent := strings.ReplaceAll(content, "%s", strconv.Itoa(rank))
	formattedContent = strings.ReplaceAll(formattedContent, "%s", strconv.FormatInt(rankPlayers[0].Score, 10))

	// 构建roleid_list（根据group类型不同）
	var roleIDListStr string
	if group == "all" {
		// 等级范围（lv0-lv9），单个值
		roleIDListStr = "lv0-lv9"
	} else {
		roleIDList := make([]string, 0, len(rankPlayers))
		for _, player := range rankPlayers {
			switch group {
			case "users":
				// 玩家ID
				roleIDList = append(roleIDList, strconv.FormatInt(player.PSID, 10))
			case "usernames":
				// 玩家名称
				if player.Username != "" {
					roleIDList = append(roleIDList, player.Username)
				} else {
					roleIDList = append(roleIDList, strconv.FormatInt(player.PSID, 10))
				}
			case "userids":
				// 玩家UID
				if player.UserID != "" {
					roleIDList = append(roleIDList, player.UserID)
				} else {
					roleIDList = append(roleIDList, strconv.FormatInt(player.PSID, 10))
				}
			case "member":
				// 公会ID
				if player.ClubID != nil {
					roleIDList = append(roleIDList, strconv.Itoa(*player.ClubID))
				} else {
					// 如果没有公会ID，使用玩家ID
					roleIDList = append(roleIDList, strconv.FormatInt(player.PSID, 10))
				}
			default:
				roleIDList = append(roleIDList, strconv.FormatInt(player.PSID, 10))
			}
		}
		roleIDListStr = strings.Join(roleIDList, ",")
	}

	// 构建表单数据
	formData := url.Values{}
	formData.Set("op", "send_mail")
	formData.Set("group", group)
	formData.Set("roleid_list", roleIDListStr)
	formData.Set("title", formattedTitle)
	formData.Set("content", formattedContent)
	formData.Set("prop_list", propList)
	if timestamp != "" {
		formData.Set("timestamp", timestamp)
	}
	if sign != "" {
		formData.Set("__sign", sign)
	}

	// 发送POST请求
	resp, err := doPostForm(apiURL, formData, httpTimeoutSec)
	if err != nil {
		for _, player := range rankPlayers {
			results = append(results, ResendResult{
				PSID:    player.PSID,
				Rank:    player.Rank,
				Success: false,
				Message: "请求失败: " + err.Error(),
			})
		}
		return results
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		for _, player := range rankPlayers {
			results = append(results, ResendResult{
				PSID:    player.PSID,
				Rank:    player.Rank,
				Success: false,
				Message: "读取响应失败: " + err.Error(),
			})
		}
		return results
	}

	// 解析响应
	var apiResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		// 如果解析失败，尝试作为文本处理
		apiResp.Message = string(body)
	}

	success := resp.StatusCode == http.StatusOK && apiResp.Code == 0
	message := apiResp.Message
	if message == "" {
		if success {
			message = "补发成功"
		} else {
			message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
	}

	// 为每个玩家记录结果
	for _, player := range rankPlayers {
		results = append(results, ResendResult{
			PSID:    player.PSID,
			Rank:    player.Rank,
			Success: success,
			Message: message,
		})
	}

	return results
}

// MailEntry 批量邮件条目
type MailEntry struct {
	PSID     int64             `json:"psid"`
	Title    string            `json:"title"`
	Content  string            `json:"content"`
	PropList string            `json:"prop_list,omitempty"` // 格式：itemid|itemnum,itemid|itemnum
	Items    map[string]int    `json:"items,omitempty"`      // 格式：{"itemid": count}
}

// ProcessResendV3 处理补发请求（批量发送版本，使用send_mail_multi接口）
func ProcessResendV3(apiURL string, rankType string, timestamp string, sign string, title string, content string, rewardConfig map[string]map[string]int, players []PlayerInfo, batchSize int, batchWorkers int, httpTimeoutSec int) ([]ResendResult, error) {
	results := make([]ResendResult, 0, len(players))
	
	// 构建批量邮件数据数组 - 预分配容量，减少内存分配
	mailEntries := make([]MailEntry, 0, len(players))
	validPlayers := make([]PlayerInfo, 0, len(players))
	
	// 预计算字符串，减少重复分配
	rankStr := make(map[int]string)
	scoreStr := make(map[int64]string)
	
	for _, player := range players {
		// 匹配奖励配置
		reward := MatchReward(player.Rank, rewardConfig)
		if reward == nil {
			results = append(results, ResendResult{
				PSID:    player.PSID,
				Rank:    player.Rank,
				Success: false,
				Message: "未找到匹配的奖励配置",
			})
			continue
		}
		
		// 构建道具列表字符串（格式：itemid|itemnum,itemid|itemnum）
		// 使用strings.Builder提高性能
		var propListBuilder strings.Builder
		propListBuilder.Grow(len(reward) * 16) // 预分配空间
		first := true
		for propIDStr, count := range reward {
			if !first {
				propListBuilder.WriteString(",")
			}
			propListBuilder.WriteString(propIDStr)
			propListBuilder.WriteString("|")
			propListBuilder.WriteString(strconv.Itoa(count))
			first = false
		}
		propList := propListBuilder.String()
		
		// 格式化邮件标题和内容（替换%s占位符）
		// 使用缓存减少字符串转换
		rStr, ok := rankStr[player.Rank]
		if !ok {
			rStr = strconv.Itoa(player.Rank)
			rankStr[player.Rank] = rStr
		}
		sStr, ok := scoreStr[player.Score]
		if !ok {
			sStr = strconv.FormatInt(player.Score, 10)
			scoreStr[player.Score] = sStr
		}
		
		formattedTitle := strings.ReplaceAll(strings.ReplaceAll(title, "%s", rStr), "%s", sStr)
		formattedContent := strings.ReplaceAll(strings.ReplaceAll(content, "%s", rStr), "%s", sStr)
		
		// 添加到邮件条目列表
		mailEntries = append(mailEntries, MailEntry{
			PSID:     player.PSID,
			Title:    formattedTitle,
			Content:  formattedContent,
			PropList: propList,
		})
		validPlayers = append(validPlayers, player)
	}
	
	if len(mailEntries) == 0 {
		return results, fmt.Errorf("没有有效的邮件数据")
	}

	// 分批发送 - 根据请求体大小动态调整批量大小
	maxBatchSize := batchSize
	if maxBatchSize < 1 {
		maxBatchSize = 1000 // 默认1000
	}
	
	// 限制单个请求的最大JSON大小（2MB），避免请求体过大导致连接被关闭
	const maxRequestSizeBytes = 2 * 1024 * 1024 // 2MB
	
	// 如果数据量小于批次大小，检查请求体大小
	if len(mailEntries) <= maxBatchSize {
		// 估算JSON大小（粗略估算：每个条目约200-500字节）
		estimatedSize := len(mailEntries) * 500
		if estimatedSize > maxRequestSizeBytes {
			// 如果估算大小超过限制，减小批量大小
			maxBatchSize = maxRequestSizeBytes / 500
			if maxBatchSize < 100 {
				maxBatchSize = 100 // 最小批量大小
			}
		}
		
		// 如果调整后的批量大小仍然可以一次性发送
		if len(mailEntries) <= maxBatchSize {
			return sendMailMultiBatch(apiURL, timestamp, sign, mailEntries, validPlayers, httpTimeoutSec)
		}
	}
	
	// 需要分批发送
	batchResults, err := sendMailMultiBatched(apiURL, timestamp, sign, mailEntries, validPlayers, maxBatchSize, batchWorkers, httpTimeoutSec)
	results = append(results, batchResults...)
	if err != nil && len(results) == 0 {
		return results, err
	}
	return results, nil
}

func sendMailMultiBatch(apiURL string, timestamp string, sign string, entries []MailEntry, players []PlayerInfo, httpTimeoutSec int) ([]ResendResult, error) {
	// 直接禁用连接复用，每次请求都使用新连接，避免连接状态异常
	return sendMailMultiBatchWithReuse(apiURL, timestamp, sign, entries, players, httpTimeoutSec, false)
}

func sendMailMultiBatchWithReuse(apiURL string, timestamp string, sign string, entries []MailEntry, players []PlayerInfo, httpTimeoutSec int, allowReuse bool) ([]ResendResult, error) {
	results := make([]ResendResult, 0, len(players))

	// 将邮件数据数组序列化为JSON
	dataJSON, err := json.Marshal(entries)
	if err != nil {
		return results, fmt.Errorf("序列化邮件数据失败: %v", err)
	}
	
	// 检查请求体大小，如果超过2MB，返回错误（应该在上层已经处理，这里作为安全检查）
	const maxRequestSizeBytes = 2 * 1024 * 1024 // 2MB
	if len(dataJSON) > maxRequestSizeBytes {
		return results, fmt.Errorf("请求体过大 (%d bytes，超过 %d bytes 限制)，请减小批量大小", len(dataJSON), maxRequestSizeBytes)
	}
	
	// 根据请求体大小动态调整超时时间
	// 小请求（<100KB）：使用配置的超时时间
	// 中等请求（100KB-500KB）：超时时间 * 2
	// 大请求（>500KB）：超时时间 * 3，但不超过60秒
	adjustedTimeout := httpTimeoutSec
	if len(dataJSON) > 500*1024 {
		adjustedTimeout = httpTimeoutSec * 3
		if adjustedTimeout > 60 {
			adjustedTimeout = 60
		}
	} else if len(dataJSON) > 100*1024 {
		adjustedTimeout = httpTimeoutSec * 2
		if adjustedTimeout > 30 {
			adjustedTimeout = 30
		}
	}

	// 构建表单数据
	formData := url.Values{}
	formData.Set("op", "send_mail_multi")
	formData.Set("data", string(dataJSON))
	if timestamp != "" {
		formData.Set("timestamp", timestamp)
	}
	if sign != "" {
		formData.Set("__sign", sign)
	}

	// 发送POST请求（使用调整后的超时时间，根据allowReuse决定是否使用连接复用）
	resp, err := doPostFormWithReuse(apiURL, formData, adjustedTimeout, allowReuse)
	if err != nil {
		for _, player := range players {
			results = append(results, ResendResult{
				PSID:    player.PSID,
				Rank:    player.Rank,
				Success: false,
				Message: "请求失败: " + err.Error(),
			})
		}
		return results, err
	}
	// 读取响应（必须完全读取，确保连接可以正确复用）
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close() // 读取失败时关闭
		for _, player := range players {
			results = append(results, ResendResult{
				PSID:    player.PSID,
				Rank:    player.Rank,
				Success: false,
				Message: "读取响应失败: " + err.Error(),
			})
		}
		return results, err
	}
	resp.Body.Close() // 确保响应体完全关闭，释放连接

	// 解析响应
	var apiResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		apiResp.Message = string(body)
	}

	success := resp.StatusCode == http.StatusOK && apiResp.Code == 0
	message := apiResp.Message
	if message == "" {
		if success {
			message = "批量补发成功"
		} else {
			message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
	}

	for _, player := range players {
		results = append(results, ResendResult{
			PSID:    player.PSID,
			Rank:    player.Rank,
			Success: success,
			Message: message,
		})
	}

	if success {
		return results, nil
	}
	return results, fmt.Errorf("批量补发失败: %s", message)
}

func sendMailMultiBatched(apiURL string, timestamp string, sign string, entries []MailEntry, players []PlayerInfo, maxBatchSize int, batchWorkers int, httpTimeoutSec int) ([]ResendResult, error) {
	// 直接禁用连接复用，每次请求都使用新连接，避免连接状态异常
	return sendMailMultiBatchedWithReuse(apiURL, timestamp, sign, entries, players, maxBatchSize, batchWorkers, httpTimeoutSec, false)
}

func sendMailMultiBatchedWithReuse(apiURL string, timestamp string, sign string, entries []MailEntry, players []PlayerInfo, maxBatchSize int, batchWorkers int, httpTimeoutSec int, allowReuse bool) ([]ResendResult, error) {
	results := make([]ResendResult, 0, len(players))
	var firstErr error

	// 如果超过最大批次，先按最大批次拆分
	if len(entries) > maxBatchSize {
		workerCount := batchWorkers
		if workerCount < 1 {
			workerCount = 1
		}
		sema := make(chan struct{}, workerCount)
		type batchResult struct {
			results []ResendResult
			err     error
		}
		resultCh := make(chan batchResult)
		var wg sync.WaitGroup

		for i := 0; i < len(entries); i += maxBatchSize {
			end := i + maxBatchSize
			if end > len(entries) {
				end = len(entries)
			}
			batchEntries := entries[i:end]
			batchPlayers := players[i:end]

			wg.Add(1)
			sema <- struct{}{}
			go func(e []MailEntry, p []PlayerInfo) {
				defer wg.Done()
				defer func() { <-sema }()
				res, err := sendMailMultiBatched(apiURL, timestamp, sign, e, p, maxBatchSize, batchWorkers, httpTimeoutSec)
				resultCh <- batchResult{results: res, err: err}
			}(batchEntries, batchPlayers)
		}

		go func() {
			wg.Wait()
			close(resultCh)
		}()

		for br := range resultCh {
			results = append(results, br.results...)
			if br.err != nil && firstErr == nil {
				firstErr = br.err
			}
		}
		return results, firstErr
	}

	// 尝试发送当前批次
	batchResults, err := sendMailMultiBatch(apiURL, timestamp, sign, entries, players, httpTimeoutSec)
	results = append(results, batchResults...)
	if err == nil {
		return results, nil
	}

	// 批次发送失败，检查是否是请求体过大或连接被关闭的错误
	// 如果是，进一步拆分批次
	isSizeError := strings.Contains(err.Error(), "请求体过大") || 
	               strings.Contains(err.Error(), "forcibly closed") ||
	               strings.Contains(err.Error(), "connection reset")
	
	if isSizeError && len(entries) > 100 {
		// 请求体过大或连接被关闭，减小批量大小重试
		newBatchSize := len(entries) / 2
		if newBatchSize < 100 {
			newBatchSize = 100
		}
		// 递归调用，使用更小的批量大小
		return sendMailMultiBatched(apiURL, timestamp, sign, entries, players, newBatchSize, batchWorkers, httpTimeoutSec)
	}

	// 其他错误或批次已经很小，进一步拆分（避免重复重试完整批次）
	if len(entries) <= 1 {
		return results, err
	}
	mid := len(entries) / 2
	leftResults, leftErr := sendMailMultiBatched(apiURL, timestamp, sign, entries[:mid], players[:mid], maxBatchSize, batchWorkers, httpTimeoutSec)
	rightResults, rightErr := sendMailMultiBatched(apiURL, timestamp, sign, entries[mid:], players[mid:], maxBatchSize, batchWorkers, httpTimeoutSec)
	results = append(results, leftResults...)
	results = append(results, rightResults...)

	if leftErr != nil {
		firstErr = leftErr
	}
	if rightErr != nil && firstErr == nil {
		firstErr = rightErr
	}
	return results, firstErr
}

func doPostForm(apiURL string, formData url.Values, httpTimeoutSec int) (*http.Response, error) {
	// 直接禁用连接复用，每次请求都使用新连接，避免连接状态异常
	return doPostFormWithReuse(apiURL, formData, httpTimeoutSec, false)
}

// doPostFormWithReuse 发送POST请求，可选择是否使用连接复用
func doPostFormWithReuse(apiURL string, formData url.Values, httpTimeoutSec int, allowReuse bool) (*http.Response, error) {
	var client *http.Client
	if allowReuse {
		client = getHTTPClient(httpTimeoutSec)
	} else {
		// 不使用连接复用，每次请求都使用新连接
		client = getHTTPClientNoReuse(httpTimeoutSec)
	}
	
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if allowReuse {
		req.Header.Set("Connection", "keep-alive") // 明确使用Keep-Alive
		req.Close = false // 确保不关闭连接，允许复用
	} else {
		req.Header.Set("Connection", "close") // 明确关闭连接
		req.Close = true // 强制关闭连接，不使用连接池
	}
	
	resp, err := client.Do(req)
	if err != nil {
		// 如果是连接错误（连接被关闭、连接重置等），可能是连接池中的连接已关闭
		// 如果允许复用，尝试重试一次，使用新连接
		if allowReuse {
			errStr := err.Error()
			if strings.Contains(errStr, "forcibly closed") || 
			   strings.Contains(errStr, "connection reset") ||
			   strings.Contains(errStr, "broken pipe") ||
			   strings.Contains(errStr, "EOF") {
				// 连接错误，重试一次，不使用连接复用
				time.Sleep(200 * time.Millisecond) // 短暂延迟，让连接完全关闭
				return doPostFormWithReuse(apiURL, formData, httpTimeoutSec, false)
			}
		}
		return nil, err
	}
	
	return resp, nil
}

func getEnvInt(name string, defaultValue int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}
	return val
}
