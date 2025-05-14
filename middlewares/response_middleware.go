package middlewares

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// responseBodyWriter 是一个自定义的ResponseWriter，用于捕获响应内容
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写写入方法，同时将内容写入到原始ResponseWriter和缓冲区
func (r *responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// WriteString 重写字符串写入方法
func (r *responseBodyWriter) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}

func ResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建一个自定义ResponseWriter来捕获响应
		w := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = w

		// 处理请求
		c.Next()

		// 检查响应是否已经发送
		if c.Writer.Written() {
			return
		}

		// 定义标准响应结构
		type StandardResponse struct {
			Code    string      `json:"code"`
			Message string      `json:"message,omitempty"`
			Data    interface{} `json:"data,omitempty"`
		}

		var standardResp StandardResponse
		statusCode := c.Writer.Status()
		strStatusCode := strconv.Itoa(statusCode)

		// 情况1: 存在错误
		if len(c.Errors) > 0 {
			// 1.1 处理状态码
			if statusCode == http.StatusOK {
				statusCode = http.StatusInternalServerError
				strStatusCode = "500"
			}

			// 1.2 处理消息内容
			responseBody := w.body.Bytes()
			if len(responseBody) > 0 {
				// 尝试解析已有的响应
				var existingResp map[string]interface{}
				if err := json.Unmarshal(responseBody, &existingResp); err == nil {
					if msg, exists := existingResp["message"].(string); exists && msg != "" {
						standardResp.Message = msg
					}
				}
			}

			// 如果消息仍为空，使用错误信息
			if standardResp.Message == "" {
				standardResp.Message = "服务器错误：" + c.Errors.Last().Error()
			}

			// 设置响应码
			standardResp.Code = "1"
		} else {
			// 情况2: 不存在错误
			// 2.1 处理状态码
			if statusCode >= 200 && statusCode < 300 {
				// 2xx 状态码，检查是否已经是标准格式，如果是则不处理
				responseBody := w.body.Bytes()
				if len(responseBody) > 0 {
					var existingResp map[string]interface{}
					if err := json.Unmarshal(responseBody, &existingResp); err == nil {
						// 检查是否已经符合标准格式
						if _, hasCode := existingResp["code"]; hasCode {
							_, hasData := existingResp["data"]
							_, hasMessage := existingResp["message"]
							if hasData || hasMessage {
								return // 已经是标准格式，不做处理
							}
						}
					}
				}
				standardResp.Code = "0"
				return // 2xx 且不是标准格式，不做处理
			} else if statusCode >= 400 && statusCode < 500 {
				// 4xx 状态码
				responseBody := w.body.Bytes()
				if len(responseBody) > 0 {
					var existingResp map[string]interface{}
					if err := json.Unmarshal(responseBody, &existingResp); err == nil {
						if msg, exists := existingResp["message"].(string); exists && msg != "" {
							standardResp.Message = msg
							standardResp.Code = strStatusCode
							// 重置响应
							c.Writer.Header().Set("Content-Type", "application/json")
							c.Writer.WriteHeader(statusCode)
							respData, _ := json.Marshal(standardResp)
							c.Writer.Write(respData)
							return
						}
					}
				}
				// 没有消息，添加默认消息
				standardResp.Message = "请求错误：" + http.StatusText(statusCode)
				standardResp.Code = "1"
			} else {
				// 其他状态码或未设置状态码
				standardResp.Code = "1"
				standardResp.Message = "服务器发生未知错误。"
				statusCode = http.StatusInternalServerError
			}
		}

		// 重置响应并写入标准格式
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.WriteHeader(statusCode)
		respData, _ := json.Marshal(standardResp)
		c.Writer.Write(respData)
	}
}
