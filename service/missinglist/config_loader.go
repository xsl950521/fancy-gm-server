package missinglist

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// LoadDefaultConfig 从指定路径加载默认配置文件
// 文件名格式：iid年月日.lua，例如：bydr20251104.lua
// 返回最新日期的配置文件内容
func LoadDefaultConfig(configPath string, iid string) (string, error) {
	// 构建文件路径
	configDir := configPath
	if configDir == "" {
		// 优先使用项目内的配置文件路径
		workDir, err := os.Getwd()
		if err == nil {
			projectConfigDir := filepath.Join(workDir, "config", "lua")
			if _, err := os.Stat(projectConfigDir); err == nil {
				configDir = projectConfigDir
			} else {
				// 如果项目内路径不存在，回退到外部路径
				configDir = `D:\workSpace\Data\Excel\Activity\Lua`
			}
		} else {
			configDir = `D:\workSpace\Data\Excel\Activity\Lua`
		}
	}

	// 查找匹配的文件（支持两种格式：iid年月日.lua 和 iid_年月日.lua）
	var files []string
	pattern1 := filepath.Join(configDir, fmt.Sprintf("%s*.lua", iid))
	pattern2 := filepath.Join(configDir, fmt.Sprintf("%s_*.lua", iid))
	
	files1, err1 := filepath.Glob(pattern1)
	if err1 == nil {
		files = append(files, files1...)
	}
	
	files2, err2 := filepath.Glob(pattern2)
	if err2 == nil {
		files = append(files, files2...)
	}
	
	if err1 != nil && err2 != nil {
		return "", fmt.Errorf("查找配置文件失败: %v, %v", err1, err2)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("未找到配置文件: %s*.lua 或 %s_*.lua", iid, iid)
	}

	// 按文件名排序（文件名包含日期，最新的在后面）
	sort.Strings(files)

	// 取最后一个文件（最新的）
	latestFile := files[len(files)-1]

	// 读取文件内容
	content, err := ioutil.ReadFile(latestFile)
	if err != nil {
		return "", fmt.Errorf("读取配置文件失败: %v", err)
	}

	return string(content), nil
}

// GetDefaultConfigPath 获取默认配置文件路径
func GetDefaultConfigPath() string {
	// 优先使用项目内的配置文件路径
	workDir, err := os.Getwd()
	if err == nil {
		projectConfigDir := filepath.Join(workDir, "config", "lua")
		if _, err := os.Stat(projectConfigDir); err == nil {
			return projectConfigDir
		}
	}
	// 如果项目内路径不存在，返回外部路径
	return `D:\workSpace\Data\Excel\Activity\Lua`
}

// ListAvailableConfigs 列出可用的配置文件
func ListAvailableConfigs(configPath string, iid string) ([]string, error) {
	configDir := configPath
	if configDir == "" {
		configDir = GetDefaultConfigPath()
	}

	files, err := filepath.Glob(filepath.Join(configDir, fmt.Sprintf("%s*.lua", iid)))
	if err != nil {
		return nil, err
	}

	// 提取文件名并排序
	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		fileNames = append(fileNames, filepath.Base(file))
	}
	sort.Strings(fileNames)

	return fileNames, nil
}

// ParseDateFromFilename 从文件名解析日期
// 文件名格式：iid年月日.lua，例如：bydr20251104.lua -> 2025-11-04
func ParseDateFromFilename(filename string) (time.Time, error) {
	// 移除扩展名
	name := strings.TrimSuffix(filename, ".lua")
	
	// 查找日期部分（8位数字：YYYYMMDD）
	datePattern := `(\d{8})`
	re := regexp.MustCompile(datePattern)
	matches := re.FindStringSubmatch(name)
	
	if len(matches) < 2 {
		return time.Time{}, fmt.Errorf("无法从文件名解析日期: %s", filename)
	}
	
	dateStr := matches[1]
	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析日期失败: %v", err)
	}
	
	return date, nil
}
