# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

每日实时金价推送工具，从工商银行网站抓取积存金实时价格，通过飞书机器人推送通知。

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

复制 `config.example.toml` 为 `config.toml`，填入飞书机器人的 accessToken：

```toml
[feishu]
accessToken = "your-token-here"
```

## 架构

```
├── main.go                      # 入口，调用handler获取金价并通过飞书发送
├── cfg/config.go                # 配置加载（viper读取config.toml）
└── internal/handler/get_price.go # 核心逻辑：抓取ICBC积存金页面并解析表格
```

**数据流**：
1. `handler.GetICBCGoldPrice()` 请求工行积存金页面
2. 使用 goquery 解析 HTML 表格，提取金价数据
3. `main.go` 将数据格式化后通过飞书 webhook 发送

## 关键依赖

- `github.com/imroc/req/v3` - HTTP 客户端
- `github.com/PuerkitoBio/goquery` - HTML 解析
- `github.com/CatchZeng/feishu` - 飞书消息推送
- `github.com/spf13/viper` - 配置管理
