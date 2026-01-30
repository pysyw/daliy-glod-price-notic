# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

每日实时金价推送工具，从工商银行网站抓取积存金实时价格，通过飞书机器人推送通知。支持价格区间监控和动态配置管理。

## 常用命令

```bash
# 运行程序
go run main.go

# 构建
go build -o daliy-glod-price-notic

# 下载依赖
go mod tidy
```

## 配置

### 环境变量配置

复制 `.env.example` 为 `.env`，配置以下参数：

```bash
# 飞书机器人 Access Token
FEI_SHU_ACCESS_TOKEN=your-token-here

# 价格阈值（从大到小排序，逗号分隔）
THRESHOLD_PRICE=1051,1047,1045

# 需要@的飞书用户ID列表
FEI_SHU_AT_USER=ou_xxx,ou_yyy

# 每个区间最大告警次数
MAX_ALERT_COUNT=10

# HTTP 服务器端口
HTTP_PORT=8080

# 服务基础 URL（用于生成配置链接）
BASE_URL=http://localhost:8080
```

### 动态配置管理

程序运行后，飞书卡片中会显示"🔧 修改配置"按钮，点击可以通过 Web 表单实时修改：
- 价格区间阈值
- @用户列表
- 最大告警次数

配置更新后立即生效，无需重启程序。

## 架构

```
├── main.go                          # 入口，同时启动定时任务和 HTTP 服务
├── cfg/config.go                    # 配置管理（支持运行时动态修改）
├── internal/
│   ├── handler/get_price.go        # 抓取 ICBC 积存金页面并解析表格
│   ├── cron/icbc/icbc.go           # 定时任务和飞书卡片构建
│   ├── server/server.go            # HTTP 服务器（配置管理接口）
│   └── global/                     # 全局配置（工作时间、假期检查）
```

**数据流**：
1. 定时任务每分钟执行一次（工作时间且非假期）
2. `handler.GetICBCGoldPrice()` 请求工行积存金页面
3. 使用 goquery 解析 HTML 表格，提取金价数据
4. `checkPriceAlert()` 判断当前价格是否在告警区间内
5. `buildFeishuCard()` 构建飞书卡片（包含配置信息和修改按钮）
6. 通过飞书 webhook 发送消息

**配置更新流程**：
1. 用户点击飞书卡片中的"修改配置"按钮
2. 浏览器打开 `{BASE_URL}/config` 配置页面
3. 提交表单后更新运行时配置（存储在内存中）
4. 下次推送时使用新配置

## 关键依赖

- `github.com/imroc/req/v3` - HTTP 客户端
- `github.com/PuerkitoBio/goquery` - HTML 解析
- `github.com/CatchZeng/feishu` - 飞书消息推送
- `github.com/robfig/cron/v3` - 定时任务调度
- `github.com/joho/godotenv` - 环境变量加载
- `github.com/patrickmn/go-cache` - 内存缓存（告警计数）

## API 接口

### GET /config
配置管理页面（Web 表单）

### POST /config
更新配置接口
- `threshold_price`: 价格阈值（逗号分隔）
- `at_users`: @用户列表（逗号分隔）
- `max_alert_count`: 最大告警次数

### GET /health
健康检查接口，返回服务状态

## 价格区间告警规则

阈值数组从大到小排序，例如 `[1051, 1047, 1045]`：
- 价格 >= 1051：不告警
- 1047 ≤ 价格 < 1051：触发区间 0 告警
- 1045 ≤ 价格 < 1047：触发区间 1 告警
- 价格 < 1045：不告警

每个区间在1小时内最多发送 `MAX_ALERT_COUNT` 次@消息，超过后仅推送不@。