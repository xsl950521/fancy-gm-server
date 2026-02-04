# Web 服务器

Redis 数据解析工具的 Web 服务器入口。

## 运行

```bash
go run main.go
```

或

```bash
go build -o web.exe
./web.exe
```

## 环境变量

- `PORT`: 服务器端口（默认：8080）

## 功能

- 文件上传接口
- 任务处理接口
- 数据查询接口
- 文件下载接口
- 静态文件服务
