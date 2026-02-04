# API 客户端对接示例

本文档提供各种编程语言的客户端对接示例，帮助开发者快速集成GM Server API。

## 目录

- [JavaScript/TypeScript](#javascripttypescript)
- [Python](#python)
- [Go](#go)
- [Java](#java)
- [C#](#c)
- [PHP](#php)
- [cURL](#curl)

---

## JavaScript/TypeScript

### 基础客户端类

```typescript
// api-client.ts
class GMAPIClient {
  private baseURL: string;

  constructor(baseURL: string = 'http://localhost:8080/api') {
    this.baseURL = baseURL;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || '请求失败');
    }

    return response.json();
  }

  // 上传文件
  async uploadFiles(
    files: File[],
    mode: string = 'all',
    archive: boolean = false
  ): Promise<UploadResponse> {
    const formData = new FormData();
    files.forEach(file => formData.append('files', file));
    formData.append('mode', mode);
    formData.append('archive', archive.toString());

    const response = await fetch(`${this.baseURL}/upload`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || '上传失败');
    }

    return response.json();
  }

  // 处理任务
  async processJob(jobId: string): Promise<ProcessResponse> {
    return this.request(`/process/${jobId}`, {
      method: 'POST',
    });
  }

  // 查询任务状态
  async getJobStatus(jobId: string): Promise<JobStatusResponse> {
    return this.request(`/status/${jobId}`);
  }

  // 查询数据
  async getData(
    jobId: string,
    options: {
      sheet?: string;
      page?: number;
      pageSize?: number;
    } = {}
  ): Promise<DataResponse> {
    const params = new URLSearchParams({
      jobId,
      ...(options.sheet && { sheet: options.sheet }),
      page: (options.page || 1).toString(),
      pageSize: (options.pageSize || 100).toString(),
    });

    return this.request(`/data?${params}`);
  }

  // 下载文件
  async downloadFile(
    jobId: string,
    format: 'excel' | 'csv' = 'excel',
    sheet?: string
  ): Promise<Blob> {
    const params = new URLSearchParams({
      format,
      ...(sheet && { sheet }),
    });

    const response = await fetch(
      `${this.baseURL}/download/${jobId}?${params}`
    );

    if (!response.ok) {
      throw new Error('下载失败');
    }

    return response.blob();
  }

  // 补发邮件
  async resendMail(config: ResendMailConfig): Promise<ResendResponse> {
    return this.request('/missinglist/resend', {
      method: 'POST',
      body: JSON.stringify(config),
    });
  }
}

// 使用示例
const client = new GMAPIClient('http://localhost:8080/api');

// 上传并处理文件
async function processFiles(files: File[]) {
  try {
    // 1. 上传文件
    const uploadResult = await client.uploadFiles(files, 'all', false);
    const jobId = uploadResult.data.jobId;
    console.log('任务已创建:', jobId);

    // 2. 启动处理
    await client.processJob(jobId);
    console.log('任务处理已启动');

    // 3. 轮询状态
    const checkStatus = async () => {
      const status = await client.getJobStatus(jobId);
      console.log('任务状态:', status.data.status);

      if (status.data.status === 'completed') {
        // 4. 查询数据
        const data = await client.getData(jobId, { page: 1, pageSize: 100 });
        console.log('数据:', data.data);

        // 5. 下载文件
        const blob = await client.downloadFile(jobId, 'excel');
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${jobId}.xlsx`;
        a.click();
      } else if (status.data.status === 'failed') {
        console.error('任务处理失败:', status.data.error);
      } else {
        // 继续轮询
        setTimeout(checkStatus, 2000);
      }
    };

    checkStatus();
  } catch (error) {
    console.error('处理失败:', error);
  }
}
```

### React Hook示例

```typescript
// useGMAPI.ts
import { useState, useCallback } from 'react';

export function useGMAPI() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const client = new GMAPIClient();

  const uploadAndProcess = useCallback(async (files: File[]) => {
    setLoading(true);
    setError(null);

    try {
      const uploadResult = await client.uploadFiles(files);
      await client.processJob(uploadResult.data.jobId);
      return uploadResult.data.jobId;
    } catch (err) {
      setError(err instanceof Error ? err.message : '上传失败');
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  return { uploadAndProcess, loading, error };
}
```

---

## Python

### 基础客户端类

```python
# api_client.py
import requests
import time
from typing import Optional, Dict, List, Any
from pathlib import Path

class GMAPIClient:
    def __init__(self, base_url: str = "http://localhost:8080/api"):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json'
        })

    def upload_files(
        self,
        file_paths: List[str],
        mode: str = 'all',
        archive: bool = False
    ) -> Dict[str, Any]:
        """上传文件"""
        url = f"{self.base_url}/upload"
        files = []
        for path in file_paths:
            files.append(('files', (Path(path).name, open(path, 'rb'), 'text/plain')))
        
        data = {
            'mode': mode,
            'archive': str(archive).lower()
        }
        
        response = self.session.post(url, files=files, data=data)
        response.raise_for_status()
        
        # 关闭文件
        for _, (_, file_obj, _) in files:
            file_obj.close()
        
        return response.json()

    def process_job(self, job_id: str) -> Dict[str, Any]:
        """处理任务"""
        url = f"{self.base_url}/process/{job_id}"
        response = self.session.post(url)
        response.raise_for_status()
        return response.json()

    def get_job_status(self, job_id: str) -> Dict[str, Any]:
        """查询任务状态"""
        url = f"{self.base_url}/status/{job_id}"
        response = self.session.get(url)
        response.raise_for_status()
        return response.json()

    def wait_for_completion(
        self,
        job_id: str,
        timeout: int = 300,
        interval: int = 2
    ) -> Dict[str, Any]:
        """等待任务完成"""
        start_time = time.time()
        
        while time.time() - start_time < timeout:
            status = self.get_job_status(job_id)
            job_status = status['data']['status']
            
            if job_status == 'completed':
                return status
            elif job_status == 'failed':
                raise Exception(f"任务失败: {status['data'].get('error', '未知错误')}")
            
            time.sleep(interval)
        
        raise TimeoutError(f"任务超时: {job_id}")

    def get_data(
        self,
        job_id: str,
        sheet: Optional[str] = None,
        page: int = 1,
        page_size: int = 100
    ) -> Dict[str, Any]:
        """查询数据"""
        params = {
            'jobId': job_id,
            'page': page,
            'pageSize': page_size
        }
        if sheet:
            params['sheet'] = sheet
        
        url = f"{self.base_url}/data"
        response = self.session.get(url, params=params)
        response.raise_for_status()
        return response.json()

    def download_file(
        self,
        job_id: str,
        output_path: str,
        format: str = 'excel',
        sheet: Optional[str] = None
    ):
        """下载文件"""
        params = {'format': format}
        if sheet:
            params['sheet'] = sheet
        
        url = f"{self.base_url}/download/{job_id}"
        response = self.session.get(url, params=params, stream=True)
        response.raise_for_status()
        
        with open(output_path, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)

    def resend_mail(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """补发邮件"""
        url = f"{self.base_url}/missinglist/resend"
        response = self.session.post(url, json=config)
        response.raise_for_status()
        return response.json()

# 使用示例
if __name__ == '__main__':
    client = GMAPIClient()
    
    # 上传并处理文件
    file_paths = ['daily_rank_20240101.txt', 'total_rank_20240101.txt']
    upload_result = client.upload_files(file_paths, mode='all')
    job_id = upload_result['data']['jobId']
    
    print(f"任务已创建: {job_id}")
    
    # 启动处理
    client.process_job(job_id)
    print("任务处理已启动")
    
    # 等待完成
    try:
        status = client.wait_for_completion(job_id, timeout=300)
        print("任务处理完成")
        
        # 查询数据
        data = client.get_data(job_id, page=1, page_size=100)
        print(f"数据总数: {data['data']['total']}")
        
        # 下载文件
        client.download_file(job_id, f'{job_id}.xlsx', format='excel')
        print("文件已下载")
    except Exception as e:
        print(f"处理失败: {e}")
```

---

## Go

### 基础客户端

```go
// api_client.go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "os"
    "time"
)

type GMAPIClient struct {
    BaseURL    string
    HTTPClient *http.Client
}

func NewGMAPIClient(baseURL string) *GMAPIClient {
    return &GMAPIClient{
        BaseURL: baseURL,
        HTTPClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

type UploadResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    struct {
        JobID  string `json:"jobId"`
        Status string `json:"status"`
        Files  []struct {
            Filename string `json:"filename"`
            Size     int64  `json:"size"`
            Type     string `json:"type"`
        } `json:"files"`
    } `json:"data"`
}

func (c *GMAPIClient) UploadFiles(filePaths []string, mode string, archive bool) (*UploadResponse, error) {
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // 添加文件
    for _, path := range filePaths {
        file, err := os.Open(path)
        if err != nil {
            return nil, err
        }
        defer file.Close()

        part, err := writer.CreateFormFile("files", file.Name())
        if err != nil {
            return nil, err
        }
        io.Copy(part, file)
    }

    writer.WriteField("mode", mode)
    writer.WriteField("archive", fmt.Sprintf("%v", archive))
    writer.Close()

    req, err := http.NewRequest("POST", c.BaseURL+"/upload", body)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result UploadResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    if result.Code != 0 {
        return nil, fmt.Errorf("上传失败: %s", result.Message)
    }

    return &result, nil
}

func (c *GMAPIClient) ProcessJob(jobID string) error {
    req, err := http.NewRequest("POST", c.BaseURL+"/process/"+jobID, nil)
    if err != nil {
        return err
    }

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("处理失败: %d", resp.StatusCode)
    }

    return nil
}

func (c *GMAPIClient) GetJobStatus(jobID string) (map[string]interface{}, error) {
    resp, err := c.HTTPClient.Get(c.BaseURL + "/status/" + jobID)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result, nil
}

func (c *GMAPIClient) WaitForCompletion(jobID string, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    for time.Now().Before(deadline) {
        status, err := c.GetJobStatus(jobID)
        if err != nil {
            return err
        }

        data := status["data"].(map[string]interface{})
        jobStatus := data["status"].(string)

        switch jobStatus {
        case "completed":
            return nil
        case "failed":
            errorMsg := ""
            if err, ok := data["error"].(string); ok {
                errorMsg = err
            }
            return fmt.Errorf("任务失败: %s", errorMsg)
        }

        <-ticker.C
    }

    return fmt.Errorf("任务超时: %s", jobID)
}

// 使用示例
func main() {
    client := NewGMAPIClient("http://localhost:8080/api")

    // 上传文件
    files := []string{"daily_rank_20240101.txt", "total_rank_20240101.txt"}
    uploadResult, err := client.UploadFiles(files, "all", false)
    if err != nil {
        panic(err)
    }

    jobID := uploadResult.Data.JobID
    fmt.Printf("任务已创建: %s\n", jobID)

    // 处理任务
    if err := client.ProcessJob(jobID); err != nil {
        panic(err)
    }
    fmt.Println("任务处理已启动")

    // 等待完成
    if err := client.WaitForCompletion(jobID, 5*time.Minute); err != nil {
        panic(err)
    }
    fmt.Println("任务处理完成")
}
```

---

## Java

### 基础客户端类

```java
// GMAPIClient.java
import com.fasterxml.jackson.databind.ObjectMapper;
import okhttp3.*;
import java.io.File;
import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

public class GMAPIClient {
    private final String baseURL;
    private final OkHttpClient client;
    private final ObjectMapper objectMapper;

    public GMAPIClient(String baseURL) {
        this.baseURL = baseURL;
        this.client = new OkHttpClient.Builder()
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .build();
        this.objectMapper = new ObjectMapper();
    }

    public Map<String, Object> uploadFiles(
        String[] filePaths,
        String mode,
        boolean archive
    ) throws IOException {
        MultipartBody.Builder builder = new MultipartBody.Builder()
            .setType(MultipartBody.FORM);

        for (String path : filePaths) {
            File file = new File(path);
            builder.addFormDataPart(
                "files",
                file.getName(),
                RequestBody.create(file, MediaType.parse("text/plain"))
            );
        }

        builder.addFormDataPart("mode", mode);
        builder.addFormDataPart("archive", String.valueOf(archive));

        Request request = new Request.Builder()
            .url(baseURL + "/upload")
            .post(builder.build())
            .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("上传失败: " + response);
            }
            return objectMapper.readValue(
                response.body().string(),
                Map.class
            );
        }
    }

    public void processJob(String jobId) throws IOException {
        Request request = new Request.Builder()
            .url(baseURL + "/process/" + jobId)
            .post(RequestBody.create("", MediaType.parse("application/json")))
            .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("处理失败: " + response);
            }
        }
    }

    public Map<String, Object> getJobStatus(String jobId) throws IOException {
        Request request = new Request.Builder()
            .url(baseURL + "/status/" + jobId)
            .get()
            .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("查询失败: " + response);
            }
            return objectMapper.readValue(
                response.body().string(),
                Map.class
            );
        }
    }

    // 使用示例
    public static void main(String[] args) throws IOException {
        GMAPIClient client = new GMAPIClient("http://localhost:8080/api");

        // 上传文件
        String[] files = {"daily_rank_20240101.txt", "total_rank_20240101.txt"};
        Map<String, Object> result = client.uploadFiles(files, "all", false);
        
        Map<String, Object> data = (Map<String, Object>) result.get("data");
        String jobId = (String) data.get("jobId");
        System.out.println("任务已创建: " + jobId);

        // 处理任务
        client.processJob(jobId);
        System.out.println("任务处理已启动");
    }
}
```

---

## C#

### 基础客户端类

```csharp
// GMAPIClient.cs
using System;
using System.Collections.Generic;
using System.IO;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

public class GMAPIClient
{
    private readonly string baseURL;
    private readonly HttpClient httpClient;

    public GMAPIClient(string baseURL = "http://localhost:8080/api")
    {
        this.baseURL = baseURL;
        this.httpClient = new HttpClient
        {
            Timeout = TimeSpan.FromSeconds(30)
        };
    }

    public async Task<UploadResponse> UploadFilesAsync(
        string[] filePaths,
        string mode = "all",
        bool archive = false
    )
    {
        using var content = new MultipartFormDataContent();

        foreach (var path in filePaths)
        {
            var fileContent = new ByteArrayContent(File.ReadAllBytes(path));
            fileContent.Headers.ContentType = new System.Net.Http.Headers.MediaTypeHeaderValue("text/plain");
            content.Add(fileContent, "files", Path.GetFileName(path));
        }

        content.Add(new StringContent(mode), "mode");
        content.Add(new StringContent(archive.ToString().ToLower()), "archive");

        var response = await httpClient.PostAsync($"{baseURL}/upload", content);
        response.EnsureSuccessStatusCode();

        var json = await response.Content.ReadAsStringAsync();
        return JsonSerializer.Deserialize<UploadResponse>(json);
    }

    public async Task ProcessJobAsync(string jobId)
    {
        var response = await httpClient.PostAsync(
            $"{baseURL}/process/{jobId}",
            new StringContent("", Encoding.UTF8, "application/json")
        );
        response.EnsureSuccessStatusCode();
    }

    public async Task<JobStatusResponse> GetJobStatusAsync(string jobId)
    {
        var response = await httpClient.GetAsync($"{baseURL}/status/{jobId}");
        response.EnsureSuccessStatusCode();

        var json = await response.Content.ReadAsStringAsync();
        return JsonSerializer.Deserialize<JobStatusResponse>(json);
    }

    public async Task WaitForCompletionAsync(string jobId, TimeSpan timeout)
    {
        var startTime = DateTime.Now;

        while (DateTime.Now - startTime < timeout)
        {
            var status = await GetJobStatusAsync(jobId);
            var jobStatus = status.Data.Status;

            if (jobStatus == "completed")
                return;
            if (jobStatus == "failed")
                throw new Exception($"任务失败: {status.Data.Error}");

            await Task.Delay(2000);
        }

        throw new TimeoutException($"任务超时: {jobId}");
    }
}

// 使用示例
class Program
{
    static async Task Main(string[] args)
    {
        var client = new GMAPIClient();

        // 上传文件
        var files = new[] { "daily_rank_20240101.txt", "total_rank_20240101.txt" };
        var uploadResult = await client.UploadFilesAsync(files, "all", false);
        var jobId = uploadResult.Data.JobId;

        Console.WriteLine($"任务已创建: {jobId}");

        // 处理任务
        await client.ProcessJobAsync(jobId);
        Console.WriteLine("任务处理已启动");

        // 等待完成
        try
        {
            await client.WaitForCompletionAsync(jobId, TimeSpan.FromMinutes(5));
            Console.WriteLine("任务处理完成");
        }
        catch (Exception ex)
        {
            Console.WriteLine($"处理失败: {ex.Message}");
        }
    }
}
```

---

## PHP

### 基础客户端类

```php
<?php
// api_client.php
class GMAPIClient {
    private $baseURL;
    private $timeout;

    public function __construct($baseURL = 'http://localhost:8080/api', $timeout = 30) {
        $this->baseURL = $baseURL;
        $this->timeout = $timeout;
    }

    public function uploadFiles($filePaths, $mode = 'all', $archive = false) {
        $ch = curl_init($this->baseURL . '/upload');
        
        $files = [];
        foreach ($filePaths as $path) {
            $files['files[]'] = new CURLFile($path, 'text/plain', basename($path));
        }
        
        $data = array_merge($files, [
            'mode' => $mode,
            'archive' => $archive ? 'true' : 'false'
        ]);
        
        curl_setopt_array($ch, [
            CURLOPT_POST => true,
            CURLOPT_POSTFIELDS => $data,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_TIMEOUT => $this->timeout,
        ]);
        
        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);
        
        if ($httpCode !== 200) {
            throw new Exception("上传失败: HTTP $httpCode");
        }
        
        return json_decode($response, true);
    }

    public function processJob($jobId) {
        $ch = curl_init($this->baseURL . "/process/$jobId");
        curl_setopt_array($ch, [
            CURLOPT_POST => true,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_TIMEOUT => $this->timeout,
        ]);
        
        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);
        
        if ($httpCode !== 200) {
            throw new Exception("处理失败: HTTP $httpCode");
        }
        
        return json_decode($response, true);
    }

    public function getJobStatus($jobId) {
        $ch = curl_init($this->baseURL . "/status/$jobId");
        curl_setopt_array($ch, [
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_TIMEOUT => $this->timeout,
        ]);
        
        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);
        
        if ($httpCode !== 200) {
            throw new Exception("查询失败: HTTP $httpCode");
        }
        
        return json_decode($response, true);
    }

    public function waitForCompletion($jobId, $timeout = 300) {
        $startTime = time();
        
        while (time() - $startTime < $timeout) {
            $status = $this->getJobStatus($jobId);
            $jobStatus = $status['data']['status'];
            
            if ($jobStatus === 'completed') {
                return $status;
            }
            if ($jobStatus === 'failed') {
                throw new Exception("任务失败: " . ($status['data']['error'] ?? '未知错误'));
            }
            
            sleep(2);
        }
        
        throw new Exception("任务超时: $jobId");
    }
}

// 使用示例
$client = new GMAPIClient();

try {
    // 上传文件
    $files = ['daily_rank_20240101.txt', 'total_rank_20240101.txt'];
    $result = $client->uploadFiles($files, 'all', false);
    $jobId = $result['data']['jobId'];
    
    echo "任务已创建: $jobId\n";
    
    // 处理任务
    $client->processJob($jobId);
    echo "任务处理已启动\n";
    
    // 等待完成
    $status = $client->waitForCompletion($jobId, 300);
    echo "任务处理完成\n";
} catch (Exception $e) {
    echo "错误: " . $e->getMessage() . "\n";
}
?>
```

---

## cURL

### 常用命令示例

```bash
#!/bin/bash

BASE_URL="http://localhost:8080/api"

# 1. 上传文件
echo "上传文件..."
UPLOAD_RESPONSE=$(curl -X POST "$BASE_URL/upload" \
  -F "files=@daily_rank_20240101.txt" \
  -F "files=@total_rank_20240101.txt" \
  -F "mode=all" \
  -F "archive=false")

JOB_ID=$(echo $UPLOAD_RESPONSE | jq -r '.data.jobId')
echo "任务ID: $JOB_ID"

# 2. 处理任务
echo "启动处理..."
curl -X POST "$BASE_URL/process/$JOB_ID"

# 3. 轮询状态
echo "等待处理完成..."
while true; do
  STATUS_RESPONSE=$(curl -s "$BASE_URL/status/$JOB_ID")
  STATUS=$(echo $STATUS_RESPONSE | jq -r '.data.status')
  
  echo "当前状态: $STATUS"
  
  if [ "$STATUS" = "completed" ]; then
    echo "处理完成！"
    break
  elif [ "$STATUS" = "failed" ]; then
    ERROR=$(echo $STATUS_RESPONSE | jq -r '.data.error')
    echo "处理失败: $ERROR"
    exit 1
  fi
  
  sleep 2
done

# 4. 查询数据
echo "查询数据..."
curl -s "$BASE_URL/data?jobId=$JOB_ID&page=1&pageSize=100" | jq '.'

# 5. 下载文件
echo "下载文件..."
curl -O "$BASE_URL/download/$JOB_ID?format=excel"
echo "文件已下载: ${JOB_ID}.xlsx"
```

---

## 最佳实践

### 1. 错误处理

```typescript
try {
  const result = await client.uploadFiles(files);
} catch (error) {
  if (error.response) {
    // HTTP错误
    const status = error.response.status;
    const data = error.response.data;
    console.error(`HTTP ${status}:`, data.message);
  } else {
    // 网络错误
    console.error('网络错误:', error.message);
  }
}
```

### 2. 重试机制

```typescript
async function retryRequest<T>(
  fn: () => Promise<T>,
  maxRetries: number = 3,
  delay: number = 1000
): Promise<T> {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }
  throw new Error('重试失败');
}
```

### 3. 进度追踪

```typescript
async function processWithProgress(jobId: string, onProgress: (progress: number) => void) {
  const checkStatus = async () => {
    const status = await client.getJobStatus(jobId);
    const progress = status.data.progress || 0;
    onProgress(progress);
    
    if (status.data.status === 'completed') {
      return;
    } else if (status.data.status === 'failed') {
      throw new Error(status.data.error);
    } else {
      setTimeout(checkStatus, 2000);
    }
  };
  
  checkStatus();
}
```

### 4. 批量处理

```typescript
async function batchProcess(files: File[][], batchSize: number = 5) {
  const results = [];
  
  for (let i = 0; i < files.length; i += batchSize) {
    const batch = files.slice(i, i + batchSize);
    const promises = batch.map(fileGroup => 
      client.uploadFiles(fileGroup).then(result => 
        client.processJob(result.data.jobId)
      )
    );
    
    const batchResults = await Promise.all(promises);
    results.push(...batchResults);
  }
  
  return results;
}
```

---

## 测试工具

### Postman Collection

可以导入以下Postman Collection进行测试：

```json
{
  "info": {
    "name": "GM Server API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "上传文件",
      "request": {
        "method": "POST",
        "header": [],
        "body": {
          "mode": "formdata",
          "formdata": [
            {
              "key": "files",
              "type": "file",
              "src": []
            },
            {
              "key": "mode",
              "value": "all",
              "type": "text"
            }
          ]
        },
        "url": {
          "raw": "{{base_url}}/upload",
          "host": ["{{base_url}}"],
          "path": ["upload"]
        }
      }
    }
  ]
}
```

---

## 常见问题

### Q1: 如何处理大文件上传？

**A:** 建议分块上传或使用流式上传：

```typescript
// 分块上传示例
async function uploadLargeFile(file: File, chunkSize: number = 1024 * 1024) {
  const chunks = Math.ceil(file.size / chunkSize);
  // 实现分块上传逻辑
}
```

### Q2: 如何实现断点续传？

**A:** 保存任务ID，定期查询状态，失败后可以重新处理。

### Q3: 如何处理超时？

**A:** 设置合理的超时时间，并实现重试机制：

```typescript
const client = new GMAPIClient();
client.setTimeout(60000); // 60秒超时
```

---

更多示例和详细说明，请参考 [API接口文档](API.md)。
