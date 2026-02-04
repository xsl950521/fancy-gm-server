package handlers

import (
	"net/http"
	"redis_data/config"
	"redis_data/service/missinglist"

	"github.com/gin-gonic/gin"
)

// MissingListHandler 漏发名单处理器
type MissingListHandler struct {
	webConfig *config.WebConfig
}

// NewMissingListHandler 创建漏发名单处理器
func NewMissingListHandler(webConfig *config.WebConfig) *MissingListHandler {
	return &MissingListHandler{
		webConfig: webConfig,
	}
}

// FetchRequest 拉取请求参数
type FetchRequest struct {
	BaseURL   string `json:"base_url" binding:"required"`
	Timestamp string `json:"timestamp" binding:"required"`
	Sign      string `json:"sign" binding:"required"`
	Iid       string `json:"iid" binding:"required"`
}

// FilterRequest 筛选请求参数
type FilterRequest struct {
	Items     []missinglist.MissingListItem `json:"items" binding:"required"`
	Hid       int                            `json:"hid"`
	Iid       string                         `json:"iid"`
	On        int64                          `json:"on"`
	IsClubRank *bool                         `json:"is_club_rank"`
	RankType   string                         `json:"rank_type"` // 总榜/日榜筛选
}

// FetchMissingList 拉取漏发名单
func (h *MissingListHandler) FetchMissingList(c *gin.Context) {
	var req FetchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}

	// 拉取数据
	items, err := missinglist.FetchMissingList(req.BaseURL, req.Timestamp, req.Sign, req.Iid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "拉取数据失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"count":   len(items),
	})
}

// FilterMissingList 筛选漏发名单
func (h *MissingListHandler) FilterMissingList(c *gin.Context) {
	var req FilterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}

	// 应用筛选
	filters := missinglist.FilterOptions{
		Hid:        req.Hid,
		Iid:        req.Iid,
		On:         req.On,
		IsClubRank: req.IsClubRank,
		RankType:   req.RankType,
	}

	filtered := missinglist.FilterItems(req.Items, filters)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    filtered,
		"count":   len(filtered),
	})
}

// GroupMissingList 分组漏发名单
func (h *MissingListHandler) GroupMissingList(c *gin.Context) {
	var req struct {
		Items []missinglist.MissingListItem `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}

	// 分组数据
	groups := missinglist.GroupItems(req.Items)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    groups,
		"count":   len(groups),
	})
}

// ResendRequest 补发请求参数
type ResendRequest struct {
	URL       string            `json:"url" binding:"required"`
	MailData  missinglist.MailData `json:"mail_data" binding:"required"`
}

// ParseExcelRequest Excel解析请求
type ParseExcelRequest struct {
	FileData string `json:"file_data" binding:"required"` // Base64编码的Excel文件
}

// ParseExcel 解析Excel文件，提取玩家ID和排名
func (h *MissingListHandler) ParseExcel(c *gin.Context) {
	var req ParseExcelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}

	// 解析Excel文件
	players, err := missinglist.ParseExcelFile(req.FileData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "解析Excel失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    players,
		"count":   len(players),
	})
}

// ResendMailRequest 补发邮件请求（新格式）
type ResendMailRequest struct {
	URL         string                      `json:"url" binding:"required"`
	RankType    string                      `json:"rank_type" binding:"required"` // 榜单类型：daily_person, total_person, daily_club, total_club
	Group       string                      `json:"group"`                          // 可选，批量发送时不需要
	Timestamp   string                      `json:"timestamp"`
	Sign        string                      `json:"sign"`
	Title       string                      `json:"title" binding:"required"`
	Content     string                      `json:"content" binding:"required"`
	RewardConfig map[string]map[string]int  `json:"reward_config" binding:"required"`
	Players     []missinglist.PlayerInfo    `json:"players" binding:"required"`
	UseBatch    bool                        `json:"use_batch"`                    // 是否使用批量发送接口
}

// ResendMail 补发邮件
func (h *MissingListHandler) ResendMail(c *gin.Context) {
	var req ResendMailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误: " + err.Error(),
		})
		return
	}

	// 处理补发请求
	var results []missinglist.ResendResult
	var err error
	
	if req.UseBatch {
		// 使用批量发送接口（send_mail_multi）
		results, err = missinglist.ProcessResendV3(
			req.URL, req.RankType, req.Timestamp, req.Sign, req.Title, req.Content, req.RewardConfig, req.Players,
			h.webConfig.ResendBatchSize, h.webConfig.ResendBatchWorkers, h.webConfig.ResendHTTPTimeoutSec,
		)
	} else {
		// 使用原来的分组发送接口（send_mail）
		results, err = missinglist.ProcessResendV2(
			req.URL, req.RankType, req.Group, req.Timestamp, req.Sign, req.Title, req.Content, req.RewardConfig, req.Players,
			h.webConfig.ResendRankWorkers, h.webConfig.ResendHTTPTimeoutSec,
		)
	}
	
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "补发失败: " + err.Error(),
			"results": results,
			"total":   len(results),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"results": results,
		"total":   len(results),
	})
}

// ParseLuaConfigRequest Lua配置解析请求
type ParseLuaConfigRequest struct {
	Content  string `json:"content" binding:"required"`  // Lua文件内容
	GameType string `json:"game_type" binding:"required"` // 游戏类型：bydr, dsc, cqsj, yxds
}

// ParseLuaConfig 解析Lua配置文件
func (h *MissingListHandler) ParseLuaConfig(c *gin.Context) {
	var req ParseLuaConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "参数错误: " + err.Error(),
		})
		return
	}

	// 解析Lua配置
	config, err := missinglist.ParseLuaConfig(req.Content, req.GameType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "解析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"config":  config,
	})
}

// LoadDefaultConfigRequest 加载默认配置请求
type LoadDefaultConfigRequest struct {
	Iid string `json:"iid" binding:"required"` // 游戏ID：bydr, dsc, cqsj, yxds
}

// LoadDefaultConfig 加载默认配置文件
func (h *MissingListHandler) LoadDefaultConfig(c *gin.Context) {
	var req LoadDefaultConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "参数错误: " + err.Error(),
		})
		return
	}

	// 加载默认配置
	configPath := missinglist.GetDefaultConfigPath()
	content, err := missinglist.LoadDefaultConfig(configPath, req.Iid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "加载配置文件失败: " + err.Error(),
		})
		return
	}

	// 解析配置
	config, err := missinglist.ParseLuaConfig(content, req.Iid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "解析配置文件失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"config":  config,
		"content": content, // 返回原始内容，用于显示
	})
}

// GetItemMap 获取道具ID到道具名称的映射
func (h *MissingListHandler) GetItemMap(c *gin.Context) {
	itemMap, err := missinglist.LoadItemMap()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "加载道具映射失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"item_map": itemMap,
	})
}
