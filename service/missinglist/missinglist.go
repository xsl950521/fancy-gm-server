package missinglist

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

// MissingListItem 漏发名单数据项
type MissingListItem struct {
	Day        int    `json:"day"`
	PSid       int64  `json:"psid"`
	Score      int64  `json:"score"`
	ServerID   string `json:"server_id"`
	Type       string `json:"type"`
	Rank       int    `json:"rank"`
	On         int64  `json:"on"`
	Hid        int    `json:"hid"`
	ClubID     *int   `json:"clubid,omitempty"`     // 可选字段，公会榜才有
	ClubScore  *int64 `json:"clubscore,omitempty"`  // 可选字段，公会榜才有
	Iid        string `json:"iid,omitempty"`         // 从请求参数中获取
	DateTime   string `json:"datetime,omitempty"`    // 转换后的时间
	IsClubRank bool   `json:"is_club_rank"`         // 是否为公会榜
}

// APIResponse API响应结构
type APIResponse struct {
	Data    []MissingListItem `json:"data"`
	Code    int               `json:"code"`
	Message string            `json:"message"`
}

// FetchMissingList 拉取漏发名单数据
func FetchMissingList(baseURL string, timestamp string, sign string, iid string) ([]MissingListItem, error) {
	// 构建请求URL
	url := fmt.Sprintf("%s?op=feast_missing_list&timestamp=%s&__sign=%s&iid=%s",
		baseURL, timestamp, sign, iid)

	// 发送HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析JSON
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("API返回错误: %s", apiResp.Message)
	}

	// 处理数据：添加iid、转换时间、标记公会榜
	items := make([]MissingListItem, 0, len(apiResp.Data))
	for _, item := range apiResp.Data {
		item.Iid = iid
		item.DateTime = convertTimestamp(item.On)
		item.IsClubRank = item.ClubID != nil

		items = append(items, item)
	}

	return items, nil
}

// convertTimestamp 转换时间戳（on + 946656000 转换为日期时间格式）
func convertTimestamp(on int64) string {
	// on + 946656000 得到秒级时间戳
	timestamp := on + 946656000
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05")
}

// FilterItems 根据条件筛选数据
func FilterItems(items []MissingListItem, filters FilterOptions) []MissingListItem {
	filtered := make([]MissingListItem, 0)

	for _, item := range items {
		// 筛选hid
		if filters.Hid > 0 && item.Hid != filters.Hid {
			continue
		}

		// 筛选iid
		if filters.Iid != "" && item.Iid != filters.Iid {
			continue
		}

		// 筛选on（时间戳）
		if filters.On > 0 && item.On != filters.On {
			continue
		}

		// 筛选公会榜/个人榜
		if filters.IsClubRank != nil {
			if *filters.IsClubRank != item.IsClubRank {
				continue
			}
		}

		// 筛选总榜/日榜
		if filters.RankType != "" {
			if filters.RankType == "daily" && item.Type != "daily" {
				continue
			}
			if filters.RankType == "total" && item.Type != "total" {
				continue
			}
		}

		filtered = append(filtered, item)
	}

	return filtered
}

// FilterOptions 筛选选项
type FilterOptions struct {
	Hid        int     `json:"hid"`
	Iid        string  `json:"iid"`
	On         int64   `json:"on"`
	IsClubRank *bool   `json:"is_club_rank"` // nil表示不过滤，true表示只显示公会榜，false表示只显示个人榜
	RankType   string  `json:"rank_type"`     // 总榜/日榜筛选：""表示不过滤，"daily"表示日榜，"total"表示总榜
}

// GroupItems 按条件分组数据
func GroupItems(items []MissingListItem) map[string][]MissingListItem {
	groups := make(map[string][]MissingListItem)

	for _, item := range items {
		// 生成分组key：hid_iid_on_type
		rankType := "个人榜"
		if item.IsClubRank {
			rankType = "公会榜"
		}
		key := fmt.Sprintf("%d_%s_%d_%s", item.Hid, item.Iid, item.On, rankType)

		groups[key] = append(groups[key], item)
	}

	// 对每组数据按rank排序
	for key := range groups {
		sort.Slice(groups[key], func(i, j int) bool {
			return groups[key][i].Rank < groups[key][j].Rank
		})
	}

	return groups
}
