package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"redis_data/api/models"
	"redis_data/clubpid"
	"redis_data/dailyrank"
	"redis_data/maillog"
	"redis_data/monthrank"
	"redis_data/rewardrecord"
	"redis_data/totalrank"

	"github.com/xuri/excelize/v2"
)

// Parser 解析器接口
type Parser interface {
	ProcessFiles(files []models.FileInfo, excelFile *excelize.File, job *models.Job) error
}

// NewParser 根据模式创建解析器
func NewParser(mode string) (Parser, error) {
	switch mode {
	case "daily":
		return &DailyRankParser{}, nil
	case "total":
		return &TotalRankParser{}, nil
	case "reward":
		return &RewardRecordParser{}, nil
	case "mail":
		return &MailLogParser{}, nil
	case "month":
		return &MonthRankParser{}, nil
	case "clubpid":
		return &ClubPidParser{}, nil
	case "all":
		return &AllParser{}, nil
	default:
		return nil, fmt.Errorf("unsupported mode: %s", mode)
	}
}

// DailyRankParser 日榜解析器
type DailyRankParser struct{}

func (p *DailyRankParser) ProcessFiles(files []models.FileInfo, excelFile *excelize.File, job *models.Job) error {
	for i, file := range files {
		if err := dailyrank.ProcessDailyRankFile(file.Path, excelFile, "个人榜", "公会榜"); err != nil {
			job.ErrorFiles++
			return fmt.Errorf("failed to process %s: %w", file.Filename, err)
		}
		job.UpdateProgress(i+1, len(files))
	}
	return nil
}

// TotalRankParser 总榜解析器
type TotalRankParser struct{}

func (p *TotalRankParser) ProcessFiles(files []models.FileInfo, excelFile *excelize.File, job *models.Job) error {
	for i, file := range files {
		if err := totalrank.ProcessTotalRankFile(file.Path, excelFile, "总榜-个人", "总榜-公会"); err != nil {
			job.ErrorFiles++
			return fmt.Errorf("failed to process %s: %w", file.Filename, err)
		}
		job.UpdateProgress(i+1, len(files))
	}
	return nil
}

// RewardRecordParser 奖励记录解析器
type RewardRecordParser struct{}

func (p *RewardRecordParser) ProcessFiles(files []models.FileInfo, excelFile *excelize.File, job *models.Job) error {
	for i, file := range files {
		if err := rewardrecord.ProcessRewardRecordFile(file.Path, excelFile, "奖励记录"); err != nil {
			job.ErrorFiles++
			return fmt.Errorf("failed to process %s: %w", file.Filename, err)
		}
		job.UpdateProgress(i+1, len(files))
	}
	return nil
}

// MailLogParser 邮件日志解析器
type MailLogParser struct{}

func (p *MailLogParser) ProcessFiles(files []models.FileInfo, excelFile *excelize.File, job *models.Job) error {
	for i, file := range files {
		if err := maillog.ProcessMailLogFile(file.Path, excelFile, "邮件日志"); err != nil {
			job.ErrorFiles++
			return fmt.Errorf("failed to process %s: %w", file.Filename, err)
		}
		job.UpdateProgress(i+1, len(files))
	}
	return nil
}

// MonthRankParser 月榜解析器
type MonthRankParser struct{}

func (p *MonthRankParser) ProcessFiles(files []models.FileInfo, excelFile *excelize.File, job *models.Job) error {
	for i, file := range files {
		if err := monthrank.ProcessMonthRankFile(file.Path, excelFile, "月榜"); err != nil {
			job.ErrorFiles++
			return fmt.Errorf("failed to process %s: %w", file.Filename, err)
		}
		job.UpdateProgress(i+1, len(files))
	}
	return nil
}

// ClubPidParser 俱乐部PID解析器
type ClubPidParser struct{}

func (p *ClubPidParser) ProcessFiles(files []models.FileInfo, excelFile *excelize.File, job *models.Job) error {
	for i, file := range files {
		if err := clubpid.ProcessClubPidFile(file.Path, excelFile, "俱乐部PID"); err != nil {
			job.ErrorFiles++
			return fmt.Errorf("failed to process %s: %w", file.Filename, err)
		}
		job.UpdateProgress(i+1, len(files))
	}
	return nil
}

// AllParser 全部模式解析器
type AllParser struct{}

func (p *AllParser) ProcessFiles(files []models.FileInfo, excelFile *excelize.File, job *models.Job) error {
	// 按类型分组文件
	filesByType := make(map[string][]models.FileInfo)
	for _, file := range files {
		filesByType[file.Type] = append(filesByType[file.Type], file)
	}

	// 按顺序处理各类型
	processors := []struct {
		mode  string
		files []models.FileInfo
	}{
		{"daily", filesByType["daily"]},
		{"total", filesByType["total"]},
		{"reward", filesByType["reward"]},
		{"mail", filesByType["mail"]},
		{"month", filesByType["month"]},
		{"clubpid", filesByType["clubpid"]},
	}

	totalProcessed := 0
	for _, proc := range processors {
		if len(proc.files) == 0 {
			continue
		}

		parser, err := NewParser(proc.mode)
		if err != nil {
			continue
		}

		if err := parser.ProcessFiles(proc.files, excelFile, job); err != nil {
			// 记录错误但继续处理其他类型
			fmt.Printf("Error processing %s: %v\n", proc.mode, err)
		}

		totalProcessed += len(proc.files)
		job.UpdateProgress(totalProcessed, len(files))
	}

	return nil
}

// ExtractSheetResults 从Excel文件提取工作表信息
func ExtractSheetResults(excelPath string) ([]models.SheetResult, error) {
	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	results := make([]models.SheetResult, 0, len(sheets))

	for _, sheetName := range sheets {
		rows, err := f.GetRows(sheetName)
		if err != nil {
			continue
		}

		if len(rows) == 0 {
			continue
		}

		// 第一行是表头
		columns := rows[0]
		rowCount := len(rows) - 1 // 减去表头

		results = append(results, models.SheetResult{
			SheetName: sheetName,
			RowCount:  rowCount,
			Columns:   columns,
		})
	}

	return results, nil
}

// SaveExcelFile 保存Excel文件
func SaveExcelFile(excelFile *excelize.File, outputPath string) error {
	// 确保目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 删除默认的Sheet1（如果存在且为空）
	sheetList := excelFile.GetSheetList()
	if len(sheetList) > 0 {
		excelFile.DeleteSheet("Sheet1")
	}

	return excelFile.SaveAs(outputPath)
}
