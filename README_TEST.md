# 测试框架使用指南

## 概述

项目已实现测试框架基础，使用 `testify` 作为断言库，支持单元测试和集成测试。

## 测试结构

```
.
├── api/
│   └── validators/
│       └── validator_test.go      # 参数验证器测试
├── pkg/
│   ├── errors/
│   │   └── errors_test.go         # 错误处理测试
│   └── response/
│       └── response_test.go      # 响应格式测试
└── test/
    ├── test_config.yaml          # 测试配置文件
    └── helpers.go                 # 测试辅助函数
```

## 运行测试

### 运行所有测试
```bash
go test ./...
```

### 运行特定包的测试
```bash
go test ./api/validators
go test ./pkg/errors
go test ./pkg/response
```

### 运行测试并显示详细输出
```bash
go test -v ./...
```

### 运行测试并显示覆盖率
```bash
go test -cover ./...
```

### 生成覆盖率报告
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 测试配置

测试使用独立的配置文件 `test/test_config.yaml`，使用独立的测试数据库 `gm_server_test`，避免影响开发环境。

## 编写测试

### 单元测试示例

```go
package handlers

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestHandlerFunction(t *testing.T) {
	// 准备测试数据
	input := "test"
	
	// 执行测试
	result := handlerFunction(input)
	
	// 断言结果
	assert.Equal(t, "expected", result)
}
```

### 集成测试示例

```go
package handlers_test

import (
	"testing"
	"redis_data/test"
	"github.com/stretchr/testify/assert"
)

func TestIntegration(t *testing.T) {
	// 设置测试数据库
	db, err := test.SetupTestDB()
	assert.NoError(t, err)
	defer test.CleanupTestDB(db)
	
	// 执行测试
	// ...
}
```

## 测试最佳实践

1. **测试命名**：使用 `Test` 前缀，描述测试的功能
2. **测试隔离**：每个测试应该独立，不依赖其他测试
3. **清理资源**：使用 `defer` 清理测试资源
4. **测试数据**：使用独立的测试数据库和配置文件
5. **断言清晰**：使用 `testify` 的断言方法，提供清晰的错误信息

## 目标覆盖率

当前目标：>60% 代码覆盖率

查看覆盖率：
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## 持续集成

建议在 CI/CD 流程中添加测试步骤：

```yaml
# .github/workflows/test.yml
- name: Run tests
  run: go test -v -coverprofile=coverage.out ./...
  
- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    file: ./coverage.out
```
