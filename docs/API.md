# GM Server API 接口文档

## 目录

1. [基础信息](#基础信息)
2. [统一响应格式](#统一响应格式)
3. [错误码说明](#错误码说明)
4. [文件上传与处理](#文件上传与处理)
5. [数据查询与下载](#数据查询与下载)
6. [历史记录](#历史记录)
7. [数据分析](#数据分析)
8. [漏发名单管理](#漏发名单管理)
9. [请求示例](#请求示例)

---

## 基础信息

### 基础URL

```
http://localhost:8080/api
```

### 请求头

所有API请求建议包含以下请求头：

```
Content-Type: application/json
X-Request-ID: <唯一请求ID>（可选，用于追踪）
```

### 认证

**⚠️ 重要：** 所有API接口都需要JWT认证（除登录、注册、刷新token接口外）。

在请求头中添加：
```
Authorization: Bearer <token>
```

获取token请参考 [认证文档](AUTH.md)。

---

## 统一响应格式

### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    // 具体数据
  }
}
```

### 错误响应

```json
{
  "code": 1001,
  "message": "参数错误",
  "error": "[1001] 参数错误: 详细错误信息"
}
```

### 分页响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [],
    "total": 100,
    "page": 1,
    "page_size": 20,
    "total_pages": 5
  }
}
```

---

## 错误码说明

| 错误码 | HTTP状态码 | 说明 |
|--------|-----------|------|
| 0 | 200 | 成功 |
| 1001 | 400 | 参数错误 |
| 1002 | 404 | 资源不存在 |
| 1003 | 500 | 内部服务器错误 |
| 1004 | 401 | 未授权 |
| 1005 | 403 | 禁止访问 |
| 1006 | 409 | 资源冲突 |
| 2001 | 404 | 任务不存在 |
| 2002 | 409 | 任务已存在 |
| 2003 | 400 | 任务处理中 |
| 2004 | 404 | 文件不存在 |
| 2005 | 400 | 文件过大 |
| 2006 | 400 | 无效文件类型 |

---

## 文件上传与处理

### 1. 上传文件

上传文件并创建处理任务。

**接口地址：** `POST /api/upload`

**请求类型：** `multipart/form-data`

**请求参数：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| files | File[] | 是 | 要上传的文件列表（.txt格式） |
| mode | String | 否 | 处理模式：all, daily, total, reward, mail, month, clubpid（默认：all） |
| archive | Boolean | 否 | 是否归档文件（默认：false） |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "jobId": "job_1704067200_123456",
    "status": "pending",
    "files": [
      {
        "filename": "daily_rank_20240101.txt",
        "size": 102400,
        "type": "daily",
        "path": "/path/to/file"
      }
    ]
  }
}
```

**cURL示例：**

```bash
curl -X POST http://localhost:8080/api/upload \
  -F "files=@daily_rank_20240101.txt" \
  -F "files=@total_rank_20240101.txt" \
  -F "mode=all" \
  -F "archive=false"
```

**JavaScript示例：**

```javascript
const formData = new FormData();
formData.append('files', file1);
formData.append('files', file2);
formData.append('mode', 'all');
formData.append('archive', 'false');

fetch('http://localhost:8080/api/upload', {
  method: 'POST',
  body: formData
})
.then(res => res.json())
.then(data => console.log(data));
```

### 2. 处理任务

启动任务处理（异步）。

**接口地址：** `POST /api/process/:jobId`

**路径参数：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| jobId | String | 是 | 任务ID |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "jobId": "job_1704067200_123456",
    "status": "processing"
  }
}
```

**cURL示例：**

```bash
curl -X POST http://localhost:8080/api/process/job_1704067200_123456
```

### 3. 查询任务状态

查询任务处理状态。

**接口地址：** `GET /api/status/:jobId`

**路径参数：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| jobId | String | 是 | 任务ID |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "job_1704067200_123456",
    "status": "completed",
    "mode": "all",
    "archive": false,
    "files": [
      {
        "filename": "daily_rank_20240101.txt",
        "size": 102400,
        "type": "daily",
        "path": "/path/to/file"
      }
    ],
    "progress": 100,
    "totalFiles": 2,
    "processedFiles": 2,
    "errorFiles": 0,
    "results": [
      {
        "sheetName": "日榜",
        "rowCount": 1000,
        "columns": ["排名", "玩家ID", "分数"]
      }
    ],
    "createdAt": "2024-01-01T10:00:00Z",
    "completedAt": "2024-01-01T10:05:00Z"
  }
}
```

**任务状态说明：**

- `pending`: 等待处理
- `processing`: 处理中
- `completed`: 已完成
- `failed`: 处理失败

---

## 数据查询与下载

### 1. 查询数据（分页）

查询处理完成的任务数据。

**接口地址：** `GET /api/data`

**查询参数：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| jobId | String | 是 | 任务ID |
| sheet | String | 否 | 工作表名称（默认：第一个工作表） |
| page | Integer | 否 | 页码（默认：1） |
| pageSize | Integer | 否 | 每页数量（默认：100，最大：1000） |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "排名": "1",
        "玩家ID": "123456",
        "分数": "10000",
        "服务器ID": "s1"
      },
      {
        "排名": "2",
        "玩家ID": "123457",
        "分数": "9999",
        "服务器ID": "s1"
      }
    ],
    "columns": ["排名", "玩家ID", "分数", "服务器ID"],
    "total": 1000,
    "page": 1,
    "page_size": 100,
    "total_pages": 10
  }
}
```

**响应字段说明：**

| 字段名 | 类型 | 说明 |
|--------|------|------|
| items | Array | 当前页的数据列表 |
| columns | Array | 列名数组 |
| total | Integer | 总记录数 |
| page | Integer | 当前页码 |
| page_size | Integer | 每页数量 |
| total_pages | Integer | 总页数 |

**cURL示例：**

```bash
curl "http://localhost:8080/api/data?jobId=job_1704067200_123456&sheet=日榜&page=1&pageSize=100"
```

### 2. 下载文件

下载处理结果文件。

**接口地址：** `GET /api/download/:jobId`

**路径参数：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| jobId | String | 是 | 任务ID |

**查询参数：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| format | String | 否 | 下载格式：excel（默认）、csv |
| sheet | String | 否 | 工作表名称（仅CSV格式需要） |

**响应：**

直接返回文件流，Content-Type根据格式设置：
- Excel: `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`
- CSV: `text/csv; charset=utf-8`

**cURL示例：**

```bash
# 下载Excel格式
curl -O http://localhost:8080/api/download/job_1704067200_123456?format=excel

# 下载CSV格式
curl -O "http://localhost:8080/api/download/job_1704067200_123456?format=csv&sheet=日榜"
```

---

## 历史记录

### 1. 获取历史记录列表（分页）

**接口地址：** `GET /api/history`

**查询参数：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | Integer | 否 | 页码（默认：1） |
| pageSize | Integer | 否 | 每页数量（默认：100，最大：1000） |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": "job_1704067200_123456",
        "mode": "all",
        "archive": false,
        "totalFiles": 2,
        "processedFiles": 2,
        "errorFiles": 0,
        "status": "completed",
        "resultPath": "/path/to/result.xlsx",
        "createdAt": "2024-01-01T10:00:00Z",
        "completedAt": "2024-01-01T10:05:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 100,
    "total_pages": 1
  }
}
```

**响应字段说明：**

| 字段名 | 类型 | 说明 |
|--------|------|------|
| items | Array | 当前页的历史记录列表 |
| total | Integer | 总记录数 |
| page | Integer | 当前页码 |
| page_size | Integer | 每页数量 |
| total_pages | Integer | 总页数 |

**cURL示例：**

```bash
curl "http://localhost:8080/api/history?page=1&pageSize=20"
```

### 2. 获取历史记录详情

**接口地址：** `GET /api/history/:id`

**路径参数：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | String | 是 | 历史记录ID（即任务ID） |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "job_1704067200_123456",
    "mode": "all",
    "archive": false,
    "files": [...],
    "results": [...],
    "status": "completed",
    "resultPath": "/path/to/result.xlsx",
    "createdAt": "2024-01-01T10:00:00Z",
    "completedAt": "2024-01-01T10:05:00Z"
  }
}
```

### 3. 删除历史记录

**接口地址：** `DELETE /api/history/:id`

**路径参数：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | String | 是 | 历史记录ID |

**响应示例：**

```json
{
  "code": 0,
  "message": "History deleted successfully"
}
```

---

## 数据分析

### 1. 获取统计数据

**接口地址：** `GET /api/analytics/statistics`

**查询参数：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| jobId | String | 是 | 任务ID |
| sheet | String | 否 | 工作表名称（默认：第一个工作表） |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "totalRows": 1000,
    "totalColumns": 10,
    "columnStats": {
      "分数": {
        "min": 1000,
        "max": 10000,
        "avg": 5000,
        "sum": 5000000
      }
    }
  }
}
```

### 2. 获取工作表列表

**接口地址：** `GET /api/analytics/sheets`

**查询参数：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| jobId | String | 是 | 任务ID |

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "sheets": [
      {
        "sheetName": "日榜",
        "rowCount": 1000
      },
      {
        "sheetName": "总榜",
        "rowCount": 5000
      }
    ]
  }
}
```

---

## 漏发名单管理

### 1. 拉取漏发名单

从游戏服务器拉取漏发名单数据。

**接口地址：** `POST /api/missinglist/fetch`

**请求体：**

```json
{
  "base_url": "http://game-server.com/api",
  "timestamp": "1704067200",
  "sign": "abc123def456",
  "iid": "bydr"
}
```

**参数说明：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| base_url | String | 是 | 游戏服务器API基础URL |
| timestamp | String | 是 | 时间戳 |
| sign | String | 是 | 签名 |
| iid | String | 是 | 游戏ID（bydr, dsc, cqsj, yxds等） |

**响应示例：**

```json
{
  "success": true,
  "data": [
    {
      "day": 1,
      "psid": 123456,
      "score": 10000,
      "server_id": "s1",
      "type": "daily_person",
      "rank": 1,
      "on": 1704067200,
      "hid": 1,
      "iid": "bydr",
      "datetime": "2024-01-01 10:00:00",
      "is_club_rank": false
    }
  ],
  "count": 100
}
```

### 2. 筛选漏发名单

根据条件筛选漏发名单。

**接口地址：** `POST /api/missinglist/filter`

**请求体：**

```json
{
  "items": [
    {
      "day": 1,
      "psid": 123456,
      "score": 10000,
      "server_id": "s1",
      "type": "daily_person",
      "rank": 1,
      "on": 1704067200,
      "hid": 1,
      "iid": "bydr"
    }
  ],
  "hid": 1,
  "iid": "bydr",
  "on": 1704067200,
  "is_club_rank": false,
  "rank_type": "daily"
}
```

**参数说明：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| items | Array | 是 | 要筛选的漏发名单数据 |
| hid | Integer | 否 | 服务器ID筛选 |
| iid | String | 否 | 游戏ID筛选 |
| on | Integer | 否 | 时间戳筛选 |
| is_club_rank | Boolean | 否 | 是否为公会榜 |
| rank_type | String | 否 | 榜单类型：daily（日榜）、total（总榜） |

**响应示例：**

```json
{
  "success": true,
  "data": [
    {
      "day": 1,
      "psid": 123456,
      "score": 10000,
      "rank": 1
    }
  ],
  "count": 50
}
```

### 3. 分组漏发名单

按排名分组漏发名单。

**接口地址：** `POST /api/missinglist/group`

**请求体：**

```json
{
  "items": [
    {
      "day": 1,
      "psid": 123456,
      "rank": 1
    },
    {
      "day": 1,
      "psid": 123457,
      "rank": 1
    },
    {
      "day": 1,
      "psid": 123458,
      "rank": 2
    }
  ]
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "1": [
      {
        "day": 1,
        "psid": 123456,
        "rank": 1
      },
      {
        "day": 1,
        "psid": 123457,
        "rank": 1
      }
    ],
    "2": [
      {
        "day": 1,
        "psid": 123458,
        "rank": 2
      }
    ]
  },
  "count": 2
}
```

### 4. 解析Excel文件

解析Excel文件，提取玩家ID和排名信息。

**接口地址：** `POST /api/missinglist/parse-excel`

**请求体：**

```json
{
  "file_data": "UEsDBBQAAAAIAG..."
}
```

**参数说明：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| file_data | String | 是 | Base64编码的Excel文件内容 |

**响应示例：**

```json
{
  "success": true,
  "data": [
    {
      "roleid": 123456,
      "rank": 1
    },
    {
      "roleid": 123457,
      "rank": 2
    }
  ],
  "count": 100
}
```

### 5. 补发邮件

执行奖励补发操作。

**接口地址：** `POST /api/missinglist/resend`

**请求体：**

```json
{
  "url": "http://game-server.com/api/send_mail",
  "rank_type": "daily_person",
  "group": "1",
  "timestamp": "1704067200",
  "sign": "abc123def456",
  "title": "排行榜奖励",
  "content": "恭喜您获得排行榜奖励",
  "reward_config": {
    "1": {
      "1001": 100,
      "1002": 50
    },
    "2": {
      "1001": 80,
      "1002": 40
    }
  },
  "players": [
    {
      "roleid": 123456,
      "rank": 1
    },
    {
      "roleid": 123457,
      "rank": 2
    }
  ],
  "use_batch": true
}
```

**参数说明：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| url | String | 是 | 游戏服务器补发API地址 |
| rank_type | String | 是 | 榜单类型：daily_person, total_person, daily_club, total_club |
| group | String | 否 | 分组（批量发送时不需要） |
| timestamp | String | 否 | 时间戳 |
| sign | String | 否 | 签名 |
| title | String | 是 | 邮件标题 |
| content | String | 是 | 邮件内容 |
| reward_config | Object | 是 | 奖励配置，key为排名，value为道具ID和数量的映射 |
| players | Array | 是 | 玩家信息列表 |
| use_batch | Boolean | 否 | 是否使用批量发送接口（默认：false） |

**响应示例：**

```json
{
  "success": true,
  "results": [
    {
      "roleid": 123456,
      "rank": 1,
      "success": true,
      "message": "发送成功"
    },
    {
      "roleid": 123457,
      "rank": 2,
      "success": false,
      "message": "发送失败: 玩家不存在"
    }
  ],
  "total": 2
}
```

### 6. 解析Lua配置

解析Lua配置文件，提取奖励配置。

**接口地址：** `POST /api/missinglist/parse-lua`

**请求体：**

```json
{
  "content": "local config = {...}",
  "game_type": "bydr"
}
```

**参数说明：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| content | String | 是 | Lua文件内容 |
| game_type | String | 是 | 游戏类型：bydr, dsc, cqsj, yxds |

**响应示例：**

```json
{
  "success": true,
  "config": {
    "1": {
      "1001": 100,
      "1002": 50
    },
    "2": {
      "1001": 80,
      "1002": 40
    }
  }
}
```

### 7. 加载默认配置

加载默认的Lua配置文件。

**接口地址：** `POST /api/missinglist/load-default-config`

**请求体：**

```json
{
  "iid": "bydr"
}
```

**参数说明：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| iid | String | 是 | 游戏ID |

**响应示例：**

```json
{
  "success": true,
  "config": {
    "1": {
      "1001": 100,
      "1002": 50
    }
  },
  "content": "local config = {...}"
}
```

### 8. 获取道具映射

获取道具ID到道具名称的映射表。

**接口地址：** `GET /api/missinglist/item-map`

**响应示例：**

```json
{
  "success": true,
  "item_map": {
    "1001": "金币",
    "1002": "钻石",
    "2001": "经验药水"
  }
}
```

---

## 请求示例

### Python示例

```python
import requests

# 上传文件
files = {'files': open('daily_rank.txt', 'rb')}
data = {'mode': 'all', 'archive': 'false'}
response = requests.post('http://localhost:8080/api/upload', files=files, data=data)
job_data = response.json()
job_id = job_data['data']['jobId']

# 处理任务
requests.post(f'http://localhost:8080/api/process/{job_id}')

# 查询状态
status = requests.get(f'http://localhost:8080/api/status/{job_id}').json()

# 查询数据
data = requests.get(f'http://localhost:8080/api/data', params={
    'jobId': job_id,
    'page': 1,
    'pageSize': 100
}).json()
```

### JavaScript/TypeScript示例

```typescript
// 上传文件
async function uploadFiles(files: File[], mode: string = 'all') {
  const formData = new FormData();
  files.forEach(file => formData.append('files', file));
  formData.append('mode', mode);
  formData.append('archive', 'false');
  
  const response = await fetch('http://localhost:8080/api/upload', {
    method: 'POST',
    body: formData
  });
  
  return await response.json();
}

// 处理任务
async function processJob(jobId: string) {
  const response = await fetch(`http://localhost:8080/api/process/${jobId}`, {
    method: 'POST'
  });
  return await response.json();
}

// 查询数据
async function getData(jobId: string, page: number = 1, pageSize: number = 100) {
  const params = new URLSearchParams({
    jobId,
    page: page.toString(),
    pageSize: pageSize.toString()
  });
  
  const response = await fetch(`http://localhost:8080/api/data?${params}`);
  return await response.json();
}

// 补发邮件
async function resendMail(config: {
  url: string;
  rankType: string;
  title: string;
  content: string;
  rewardConfig: Record<string, Record<string, number>>;
  players: Array<{roleid: number; rank: number}>;
  useBatch?: boolean;
}) {
  const response = await fetch('http://localhost:8080/api/missinglist/resend', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(config)
  });
  
  return await response.json();
}
```

### Go示例

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "os"
)

// 上传文件
func uploadFile(filePath string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("files", filePath)
    if err != nil {
        return "", err
    }
    io.Copy(part, file)
    writer.WriteField("mode", "all")
    writer.Close()

    req, _ := http.NewRequest("POST", "http://localhost:8080/api/upload", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    data := result["data"].(map[string]interface{})
    return data["jobId"].(string), nil
}

// 查询任务状态
func getJobStatus(jobId string) (map[string]interface{}, error) {
    resp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/status/%s", jobId))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    return result, nil
}
```

---

## 注意事项

1. **文件格式**：上传的文件必须是 `.txt` 格式
2. **文件大小**：建议单个文件不超过 100MB
3. **并发限制**：建议控制并发请求数量，避免服务器过载
4. **任务状态**：任务处理是异步的，需要轮询状态接口获取最新状态
5. **数据保留**：处理结果文件会保留一段时间，建议及时下载
6. **错误处理**：所有接口都可能返回错误，客户端需要处理各种错误情况

---

## 更新日志

### v1.0.0
- 初始版本
- 支持文件上传和处理
- 支持漏发名单管理
- 支持数据查询和下载
- 支持历史记录管理

---

## 技术支持

如有问题，请联系技术支持或查看项目文档。
