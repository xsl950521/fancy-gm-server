package model

// FileInfo 文件信息
type FileInfo struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Type     string `json:"type"`
	Path     string `json:"path"`
}

// SheetResult 工作表结果
type SheetResult struct {
	SheetName string          `json:"sheetName"`
	RowCount  int             `json:"rowCount"`
	Columns   []string        `json:"columns"`
	Data      [][]interface{} `json:"data,omitempty"`
}
