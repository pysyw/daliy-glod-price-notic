package server

import (
	"daliy-glod-price-notic/cfg"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Server HTTP 服务器
type Server struct {
	port   string
	engine *gin.Engine
}

// NewServer 创建 HTTP 服务器
func NewServer() *Server {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080" // 默认端口
	}

	// 设置 Gin 模式（生产环境使用 release 模式）
	gin.SetMode(gin.ReleaseMode)

	engine := gin.Default()

	s := &Server{
		port:   port,
		engine: engine,
	}

	// 注册路由
	s.setupRoutes()

	return s
}

// setupRoutes 注册路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.engine.GET("/health", s.handleHealth)

	// 配置管理
	s.engine.GET("/config", s.showConfigForm)
	s.engine.POST("/config", s.updateConfig)
}

// Start 启动 HTTP 服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.port)
	fmt.Printf("[%s] HTTP 服务器启动在端口 %s\n", time.Now().Format("2006-01-02 15:04:05"), s.port)
	return s.engine.Run(addr)
}

// handleHealth 健康检查接口
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format("2006-01-02 15:04:05"),
	})
}

// showConfigForm 显示配置表单页面
func (s *Server) showConfigForm(c *gin.Context) {
	runtimeCfg := cfg.GetRuntimeConfig()

	// 获取当前配置
	thresholds := runtimeCfg.GetThresholdPrice()
	atUsers := runtimeCfg.GetFeiShuAtUser()
	maxAlert := runtimeCfg.GetMaxAlertCount()

	// 格式化阈值字符串（逗号分隔）
	thresholdStr := ""
	if len(thresholds) > 0 {
		var parts []string
		for _, t := range thresholds {
			parts = append(parts, fmt.Sprintf("%.2f", t))
		}
		thresholdStr = strings.Join(parts, ",")
	}

	// 格式化@用户字符串（逗号分隔）
	atUserStr := strings.Join(atUsers, ",")

	// 准备模板数据
	data := gin.H{
		"ThresholdPrice": thresholdStr,
		"AtUsers":        atUserStr,
		"MaxAlertCount":  maxAlert,
		"Success":        c.Query("success") == "true",
	}

	// 渲染 HTML 页面
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, renderConfigPage(data))
}

// updateConfig 更新配置
func (s *Server) updateConfig(c *gin.Context) {
	runtimeCfg := cfg.GetRuntimeConfig()

	// 解析价格阈值
	thresholdStr := c.PostForm("threshold_price")
	if thresholdStr != "" {
		parts := strings.Split(thresholdStr, ",")
		var thresholds []float64
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if val, err := strconv.ParseFloat(part, 64); err == nil {
				thresholds = append(thresholds, val)
			}
		}
		if len(thresholds) > 0 {
			runtimeCfg.SetThresholdPrice(thresholds)
		}
	}

	// 解析@用户列表
	atUserStr := c.PostForm("at_users")
	if atUserStr != "" {
		parts := strings.Split(atUserStr, ",")
		var users []string
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				users = append(users, part)
			}
		}
		if len(users) > 0 {
			runtimeCfg.SetFeiShuAtUser(users)
		}
	}

	// 解析最大告警次数
	maxAlertStr := c.PostForm("max_alert_count")
	if maxAlertStr != "" {
		if val, err := strconv.Atoi(maxAlertStr); err == nil && val > 0 {
			runtimeCfg.SetMaxAlertCount(val)
		}
	}

	fmt.Printf("[%s] 配置更新成功 - 阈值: %v, @用户: %v, 最大告警次数: %d\n",
		time.Now().Format("2006-01-02 15:04:05"),
		runtimeCfg.GetThresholdPrice(),
		runtimeCfg.GetFeiShuAtUser(),
		runtimeCfg.GetMaxAlertCount())

	// 重定向回配置页面并显示成功消息
	c.Redirect(http.StatusSeeOther, "/config?success=true")
}

// renderConfigPage 渲染配置页面 HTML
func renderConfigPage(data gin.H) string {
	thresholdPrice := data["ThresholdPrice"].(string)
	atUsers := data["AtUsers"].(string)
	maxAlertCount := data["MaxAlertCount"].(int)
	success := data["Success"].(bool)

	successHTML := ""
	if success {
		successHTML = `
        <div class="success-message">
            ✓ 配置更新成功！新配置已立即生效。
        </div>`
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>金价推送配置</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }

        .container {
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            max-width: 600px;
            width: 100%%;
            padding: 40px;
        }

        h1 {
            font-size: 28px;
            color: #333;
            margin-bottom: 10px;
            text-align: center;
        }

        .subtitle {
            text-align: center;
            color: #666;
            margin-bottom: 30px;
            font-size: 14px;
        }

        .success-message {
            background: #d4edda;
            border: 1px solid #c3e6cb;
            color: #155724;
            padding: 12px 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            text-align: center;
            font-size: 14px;
        }

        .form-group {
            margin-bottom: 24px;
        }

        label {
            display: block;
            font-weight: 600;
            color: #333;
            margin-bottom: 8px;
            font-size: 14px;
        }

        .help-text {
            color: #666;
            font-size: 13px;
            margin-top: 6px;
            line-height: 1.5;
        }

        input[type="text"],
        input[type="number"] {
            width: 100%%;
            padding: 12px 16px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 15px;
            transition: border-color 0.3s;
        }

        input[type="text"]:focus,
        input[type="number"]:focus {
            outline: none;
            border-color: #667eea;
        }

        .button-group {
            display: flex;
            gap: 12px;
            margin-top: 32px;
        }

        button {
            flex: 1;
            padding: 14px 24px;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s;
        }

        .btn-primary {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
        }

        .btn-primary:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 20px rgba(102, 126, 234, 0.4);
        }

        .btn-secondary {
            background: #f5f5f5;
            color: #666;
        }

        .btn-secondary:hover {
            background: #e0e0e0;
        }

        .example {
            background: #f8f9fa;
            border-left: 4px solid #667eea;
            padding: 12px 16px;
            margin-top: 8px;
            border-radius: 4px;
            font-size: 13px;
            color: #555;
        }

        .example strong {
            color: #667eea;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>⚙️ 金价推送配置</h1>
        <p class="subtitle">动态修改价格区间和告警设置</p>

        %s

        <form method="POST" action="/config">
            <div class="form-group">
                <label for="threshold_price">价格阈值区间</label>
                <input type="text" id="threshold_price" name="threshold_price"
                       value="%s"
                       placeholder="1051,1047,1045" required>
                <div class="help-text">
                    多个阈值用英文逗号分隔，从大到小排序。系统会在价格进入区间时发送@提醒。
                </div>
                <div class="example">
                    <strong>示例：</strong> 1051,1047,1045 表示价格在 [1047, 1051) 和 [1045, 1047) 区间时会触发告警
                </div>
            </div>

            <div class="form-group">
                <label for="at_users">@用户列表</label>
                <input type="text" id="at_users" name="at_users"
                       value="%s"
                       placeholder="ou_xxx,ou_yyy">
                <div class="help-text">
                    需要@的飞书用户 ID，多个用户用英文逗号分隔。留空则不@任何人。
                </div>
                <div class="example">
                    <strong>获取用户 ID：</strong> 在飞书开放平台查看用户的 open_id 或 user_id
                </div>
            </div>

            <div class="form-group">
                <label for="max_alert_count">最大告警次数</label>
                <input type="number" id="max_alert_count" name="max_alert_count"
                       value="%d"
                       min="1" max="100" required>
                <div class="help-text">
                    每个价格区间最多发送@消息的次数，超过后仅推送不@。有效期1小时。
                </div>
            </div>

            <div class="button-group">
                <button type="button" class="btn-secondary" onclick="window.history.back()">取消</button>
                <button type="submit" class="btn-primary">保存配置</button>
            </div>
        </form>
    </div>
</body>
</html>
`, successHTML, thresholdPrice, atUsers, maxAlertCount)
}
