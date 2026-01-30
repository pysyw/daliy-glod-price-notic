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
	priceIntervals := runtimeCfg.GetPriceIntervals()
	atUsers := runtimeCfg.GetFeiShuAtUser()

	// 格式化@用户字符串（逗号分隔）
	atUserStr := strings.Join(atUsers, ",")

	// 准备模板数据
	data := gin.H{
		"PriceIntervals": priceIntervals,
		"AtUsers":        atUserStr,
		"Success":        c.Query("success") == "true",
	}

	// 渲染 HTML 页面
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, renderConfigPage(data))
}

// updateConfig 更新配置
func (s *Server) updateConfig(c *gin.Context) {
	runtimeCfg := cfg.GetRuntimeConfig()

	// 解析价格区间配置（从动态表单字段）
	var intervals []cfg.PriceInterval
	index := 0
	for {
		lowerStr := c.PostForm(fmt.Sprintf("interval_lower_%d", index))
		upperStr := c.PostForm(fmt.Sprintf("interval_upper_%d", index))
		countStr := c.PostForm(fmt.Sprintf("interval_count_%d", index))

		// 如果字段不存在，说明没有更多区间了
		if lowerStr == "" && upperStr == "" && countStr == "" {
			break
		}

		// 解析数值
		lower, err1 := strconv.ParseFloat(lowerStr, 64)
		upper, err2 := strconv.ParseFloat(upperStr, 64)
		count, err3 := strconv.Atoi(countStr)

		// 验证数据有效性
		if err1 == nil && err2 == nil && err3 == nil && lower < upper && count > 0 {
			intervals = append(intervals, cfg.PriceInterval{
				Lower:         lower,
				Upper:         upper,
				MaxAlertCount: count,
			})
		}

		index++
	}

	// 更新配置
	if len(intervals) > 0 {
		runtimeCfg.SetPriceIntervals(intervals)
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

	fmt.Printf("[%s] 配置更新成功\n", time.Now().Format("2006-01-02 15:04:05"))

	// 重定向回配置页面并显示成功消息
	c.Redirect(http.StatusSeeOther, "/config?success=true")
}

// renderConfigPage 渲染配置页面 HTML
func renderConfigPage(data gin.H) string {
	priceIntervals := data["PriceIntervals"].([]cfg.PriceInterval)
	atUsers := data["AtUsers"].(string)
	success := data["Success"].(bool)

	successHTML := ""
	if success {
		successHTML = `
        <div class="success-message">
            ✓ 配置更新成功！新配置已立即生效。
        </div>`
	}

	// 生成区间行HTML
	intervalRowsHTML := ""
	if len(priceIntervals) == 0 {
		// 如果没有配置，添加一个空行
		intervalRowsHTML = `
            <div class="interval-row" data-index="区间 1">
                <input type="number" name="interval_lower_0" placeholder="下限（元）" step="0.01" required>
                <span class="separator">-</span>
                <input type="number" name="interval_upper_0" placeholder="上限（元）" step="0.01" required>
                <span class="separator">:</span>
                <input type="number" name="interval_count_0" placeholder="告警次数" min="1" required>
                <button type="button" class="btn-delete" onclick="removeInterval(this)">删除</button>
            </div>`
	} else {
		for i, interval := range priceIntervals {
			intervalRowsHTML += fmt.Sprintf(`
            <div class="interval-row" data-index="区间 %d">
                <input type="number" name="interval_lower_%d" value="%.2f" placeholder="下限（元）" step="0.01" required>
                <span class="separator">-</span>
                <input type="number" name="interval_upper_%d" value="%.2f" placeholder="上限（元）" step="0.01" required>
                <span class="separator">:</span>
                <input type="number" name="interval_count_%d" value="%d" placeholder="告警次数" min="1" required>
                <button type="button" class="btn-delete" onclick="removeInterval(this)">删除</button>
            </div>`, i+1, i, interval.Lower, i, interval.Upper, i, interval.MaxAlertCount)
		}
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
            max-width: 700px;
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

        .intervals-container {
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            padding: 16px;
            background: #f8f9fa;
        }

        .interval-row {
            display: flex;
            align-items: center;
            gap: 8px;
            margin-bottom: 12px;
            flex-wrap: wrap;
        }

        .interval-row input[type="number"] {
            flex: 1;
            min-width: 80px;
            padding: 10px 12px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 14px;
            transition: border-color 0.3s;
        }

        .interval-row input[type="number"]:focus {
            outline: none;
            border-color: #667eea;
        }

        .interval-row .separator {
            color: #999;
            font-weight: bold;
            font-size: 16px;
            flex-shrink: 0;
        }

        .btn-delete {
            padding: 8px 12px;
            background: #ff4757;
            color: white;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-size: 16px;
            font-weight: bold;
            transition: background 0.3s;
            flex-shrink: 0;
        }

        .btn-delete:hover {
            background: #ee5a6f;
        }

        /* 移动端适配 */
        @media (max-width: 768px) {
            body {
                padding: 10px;
            }

            .container {
                padding: 24px 20px;
            }

            h1 {
                font-size: 24px;
            }

            .subtitle {
                font-size: 13px;
            }

            .interval-row {
                flex-direction: column;
                align-items: stretch;
                gap: 10px;
                padding: 12px;
                background: white;
                border-radius: 8px;
                margin-bottom: 16px;
                position: relative;
            }

            .interval-row input[type="number"] {
                width: 100%%;
                min-width: unset;
                padding: 12px 14px;
                font-size: 16px;
            }

            .interval-row .separator {
                display: none;
            }

            .interval-row::before {
                content: attr(data-index);
                position: absolute;
                top: -8px;
                left: 12px;
                background: #667eea;
                color: white;
                padding: 2px 8px;
                border-radius: 4px;
                font-size: 12px;
                font-weight: bold;
            }

            .btn-delete {
                width: 100%%;
                padding: 12px;
            }

            .intervals-container {
                padding: 12px;
            }

            .btn-add {
                padding: 14px;
                font-size: 15px;
            }

            .button-group {
                flex-direction: column;
            }

            button[type="submit"],
            button[type="button"].btn-secondary {
                width: 100%%;
                padding: 16px;
            }
        }

        /* 小屏手机适配 */
        @media (max-width: 480px) {
            .container {
                padding: 20px 16px;
            }

            h1 {
                font-size: 22px;
            }

            .form-group {
                margin-bottom: 20px;
            }

            .help-text {
                font-size: 12px;
            }

            .example {
                font-size: 12px;
                padding: 10px 12px;
            }
        }

        .btn-add {
            width: 100%%;
            padding: 12px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 600;
            transition: background 0.3s;
            margin-top: 8px;
        }

        .btn-add:hover {
            background: #5568d3;
        }

        input[type="text"] {
            width: 100%%;
            padding: 12px 16px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 15px;
            transition: border-color 0.3s;
        }

        input[type="text"]:focus {
            outline: none;
            border-color: #667eea;
        }

        .button-group {
            display: flex;
            gap: 12px;
            margin-top: 32px;
        }

        button[type="submit"],
        button[type="button"].btn-secondary {
            flex: 1;
            padding: 14px 24px;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s;
        }

        button[type="submit"] {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
        }

        button[type="submit"]:hover {
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
            background: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 12px 16px;
            margin-top: 8px;
            border-radius: 4px;
            font-size: 13px;
            color: #856404;
            line-height: 1.6;
        }

        .example strong {
            color: #e67e22;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>⚙️ 金价推送配置</h1>
        <p class="subtitle">动态修改价格区间和告警设置</p>

        %s

        <form method="POST" action="/config" onsubmit="reindexIntervals()">
            <div class="form-group">
                <label>价格区间配置</label>
                <div class="intervals-container">
                    <div id="intervals-list">
                        %s
                    </div>
                    <button type="button" class="btn-add" onclick="addInterval()">+ 添加区间</button>
                </div>
                <div class="help-text">
                    每个区间包含：<strong>下限</strong>、<strong>上限</strong>、<strong>最大告警次数</strong>
                </div>
                <div class="example">
                    <strong>💡 配置说明：</strong>
                    <br>• 下限 1045 - 上限 1047 : 次数 5
                    <br>&nbsp;&nbsp;&nbsp;→ 表示价格在 [1045, 1047) 区间时最多告警5次
                    <br>• 下限 1047 - 上限 1051 : 次数 10
                    <br>&nbsp;&nbsp;&nbsp;→ 表示价格在 [1047, 1051) 区间时最多告警10次
                    <br><br><strong>⚠️ 重要：</strong>只有当价格在区间内才会触发@提醒！
                    <br>例如：配置 1080-1185:3，则价格在 1080~1185 元之间都会@人
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
            </div>

            <div class="button-group">
                <button type="button" class="btn-secondary" onclick="window.history.back()">取消</button>
                <button type="submit">保存配置</button>
            </div>
        </form>
    </div>

    <script>
        let intervalCount = %d;

        function addInterval() {
            const container = document.getElementById('intervals-list');
            const newRow = document.createElement('div');
            newRow.className = 'interval-row';
            newRow.setAttribute('data-index', '区间 ' + (container.children.length + 1));
            newRow.innerHTML = ` + "`" + `
                <input type="number" name="interval_lower_${intervalCount}" placeholder="下限（元）" step="0.01" required>
                <span class="separator">-</span>
                <input type="number" name="interval_upper_${intervalCount}" placeholder="上限（元）" step="0.01" required>
                <span class="separator">:</span>
                <input type="number" name="interval_count_${intervalCount}" placeholder="告警次数" min="1" required>
                <button type="button" class="btn-delete" onclick="removeInterval(this)">删除</button>
            ` + "`" + `;
            container.appendChild(newRow);
            intervalCount++;
            updateIntervalIndexes();
        }

        function removeInterval(button) {
            const row = button.parentElement;
            row.remove();
            updateIntervalIndexes();
        }

        function updateIntervalIndexes() {
            const rows = document.querySelectorAll('.interval-row');
            rows.forEach((row, index) => {
                row.setAttribute('data-index', '区间 ' + (index + 1));
            });
        }

        function reindexIntervals() {
            const rows = document.querySelectorAll('.interval-row');
            rows.forEach((row, index) => {
                const inputs = row.querySelectorAll('input');
                inputs[0].name = ` + "`" + `interval_lower_${index}` + "`" + `;
                inputs[1].name = ` + "`" + `interval_upper_${index}` + "`" + `;
                inputs[2].name = ` + "`" + `interval_count_${index}` + "`" + `;
            });
        }
    </script>
</body>
</html>
`, successHTML, intervalRowsHTML, atUsers, len(priceIntervals))
}
