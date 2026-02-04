package missinglist

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xuri/excelize/v2"
)

var (
	itemMap     map[string]string // 道具ID -> 道具名称
	itemMapOnce sync.Once
	itemMapMu   sync.RWMutex
)

// LoadItemMap 加载道具ID到道具名称的映射
func LoadItemMap() (map[string]string, error) {
	var err error
	itemMapOnce.Do(func() {
		itemMapMu.Lock()
		defer itemMapMu.Unlock()

		// 尝试从项目内路径加载
		itemPath := filepath.Join("config", "item.xlsx")
		itemMap, err = loadItemMapFromFile(itemPath)
		if err != nil {
			// 如果项目内路径失败，尝试从外部路径加载
			itemPath = "D:\\workSpace\\Data\\Excel\\Function\\Excel\\item.xlsx"
			itemMap, err = loadItemMapFromFile(itemPath)
		}
	})

	if err != nil {
		return nil, err
	}

	itemMapMu.RLock()
	defer itemMapMu.RUnlock()

	// 返回副本
	result := make(map[string]string)
	for k, v := range itemMap {
		result[k] = v
	}
	return result, nil
}

// loadItemMapFromFile 从Excel文件加载道具映射
func loadItemMapFromFile(filePath string) (map[string]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开item.xlsx失败: %w", err)
	}
	defer f.Close()

	// 获取第一个工作表
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("item.xlsx中没有工作表")
	}
	sheetName := sheetList[0]

	// 读取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("item.xlsx数据不足（至少需要表头和数据行）")
	}

	// 第一行是表头，查找道具ID和道具名称列
	headers := rows[0]
	itemIdCol := -1
	itemNameCol := -1

	for i, header := range headers {
		headerLower := strings.ToLower(strings.TrimSpace(header))
		// 可能的列名：id, item_id, 道具id, itemid, 编号
		if headerLower == "id" || headerLower == "item_id" || headerLower == "道具id" || 
		   headerLower == "itemid" || headerLower == "编号" || headerLower == "道具编号" {
			itemIdCol = i
		}
		// 可能的列名：name, item_name, 道具名称, itemname, 名称
		if headerLower == "name" || headerLower == "item_name" || headerLower == "道具名称" || 
		   headerLower == "itemname" || headerLower == "名称" {
			itemNameCol = i
		}
	}

	if itemIdCol == -1 {
		return nil, fmt.Errorf("未找到道具ID列（请确保表头包含id、item_id或道具id）")
	}
	if itemNameCol == -1 {
		return nil, fmt.Errorf("未找到道具名称列（请确保表头包含name、item_name或道具名称）")
	}

	// 构建映射
	result := make(map[string]string)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) <= itemIdCol || len(row) <= itemNameCol {
			continue
		}

		itemId := strings.TrimSpace(row[itemIdCol])
		itemName := strings.TrimSpace(row[itemNameCol])

		if itemId != "" && itemName != "" {
			result[itemId] = itemName
		}
	}

	return result, nil
}

// GetItemName 获取道具名称（如果找不到则返回ID）
func GetItemName(itemId string) string {
	itemMapMu.RLock()
	defer itemMapMu.RUnlock()

	if itemMap == nil {
		// 如果映射未加载，尝试加载
		itemMapMu.RUnlock()
		_, err := LoadItemMap()
		itemMapMu.RLock()
		if err != nil {
			return itemId
		}
	}

	if name, ok := itemMap[itemId]; ok {
		return name
	}
	return itemId
}

// GetItemNames 批量获取道具名称
func GetItemNames(itemIds []string) map[string]string {
	result := make(map[string]string)
	
	itemMapMu.RLock()
	defer itemMapMu.RUnlock()

	if itemMap == nil {
		// 如果映射未加载，尝试加载
		itemMapMu.RUnlock()
		_, err := LoadItemMap()
		itemMapMu.RLock()
		if err != nil {
			// 如果加载失败，返回ID本身
			for _, id := range itemIds {
				result[id] = id
			}
			return result
		}
	}

	for _, id := range itemIds {
		if name, ok := itemMap[id]; ok {
			result[id] = name
		} else {
			result[id] = id
		}
	}

	return result
}
