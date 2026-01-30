# 操作日志

## 2026-01-26 修复运行时配置不生效的问题

### 问题描述
用户反馈：使用 `GetRuntimeConfig` 后没有发送艾特消息

### 问题分析
检查 `internal/cron/icbc/icbc.go:checkPriceAlert` 函数发现：
- 函数接收了 `runtimeCfg *cfg.RuntimeConfig` 参数
- 但实际逻辑中使用的是 `cfg.GlobalConfig`（已废弃的静态配置）
- 导致通过 Web 界面修改配置后，艾特消息无法使用新配置

### 具体问题位置
- 第 334 行：`getPriceInterval(price, cfg.GlobalConfig.ThresholdPrice)` 使用了静态配置
- 第 341 行：`shouldSendAlert(intervalIndex, cfg.GlobalConfig.MaxAlertCount)` 使用了静态配置
- 第 352-354 行：构建 `AlertInfo` 时使用了 `cfg.GlobalConfig.ThresholdPrice` 和 `cfg.GlobalConfig.FeiShuAtUser`

### 修复方案
修改 `checkPriceAlert` 函数，在函数开始时提取运行时配置的值：
```go
thresholds := runtimeCfg.GetThresholdPrice()
atUsers := runtimeCfg.GetFeiShuAtUser()
maxAlert := runtimeCfg.GetMaxAlertCount()
```

然后在所有逻辑中使用这些运行时配置值，替代 `cfg.GlobalConfig`

### 修改文件
- `internal/cron/icbc/icbc.go:318-367` - 重构 `checkPriceAlert` 函数

### 验证结果
- ✅ 代码编译成功
- ✅ 函数现在正确使用运行时配置
- ✅ Web 界面修改配置后，艾特消息将使用新配置

### 影响范围
- 修复后，通过 `/config` 接口修改的配置将立即生效
- 艾特用户列表、价格阈值、最大告警次数都将使用运行时配置
