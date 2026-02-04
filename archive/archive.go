package archive

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CreateArchiveDir 创建基于日期的归档目录
// 返回归档目录路径
func CreateArchiveDir(basePath string) (string, error) {
	// 使用当前日期创建目录名
	dateStr := time.Now().Format("2006-01-02")
	archiveDir := filepath.Join(basePath, "archive", dateStr)

	// 创建目录（如果不存在）
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create archive directory: %w", err)
	}

	return archiveDir, nil
}

// ArchiveFile 将文件移动到归档目录
func ArchiveFile(filePath, archiveDir string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	fileName := filepath.Base(filePath)
	destPath := filepath.Join(archiveDir, fileName)

	// 如果目标文件已存在，添加时间戳后缀
	if _, err := os.Stat(destPath); err == nil {
		timestamp := time.Now().Format("150405")
		ext := filepath.Ext(fileName)
		name := fileName[:len(fileName)-len(ext)]
		destPath = filepath.Join(archiveDir, fmt.Sprintf("%s_%s%s", name, timestamp, ext))
	}

	// 移动文件
	if err := os.Rename(filePath, destPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// ArchiveFiles 归档多个文件
func ArchiveFiles(filePaths []string, archiveDir string) []error {
	var errors []error
	for _, filePath := range filePaths {
		if err := ArchiveFile(filePath, archiveDir); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

