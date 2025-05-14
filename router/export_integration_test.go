package router

import (
	"TeamTickBackend/gen"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ---- FAKE HANDLER ----
type fakeExportHandler struct{}

func (h *fakeExportHandler) GetExportCheckinsXlsx(c *gin.Context, params gen.GetExportCheckinsXlsxParams) {
	// 生成一个临时xlsx文件
	file := excelize.NewFile()
	filePath := "test_statistics.xlsx"
	err := file.SaveAs(filePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "文件生成失败")
		return
	}
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "文件读取失败")
		return
	}
	fileReader := bytes.NewReader(fileContent)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=statistics.xlsx")
	c.Header("Content-Length", fmt.Sprintf("%d", len(fileContent)))
	c.Status(http.StatusOK)
	_, _ = fileReader.WriteTo(c.Writer)
}

func setupExportOnlyRouterWithFakeHandler() *gin.Engine {
	r := gin.Default()
	h := &fakeExportHandler{}
	gen.RegisterExportHandlers(r, h)
	return r
}

func TestExportCheckinsXlsxRoute(t *testing.T) {
	ginRouter := setupExportOnlyRouterWithFakeHandler()

	// 构造请求参数
	start := int(time.Now().AddDate(0, 0, -7).Unix())
	end := int(time.Now().Unix())
	url := "/export/checkins.xlsx?groupIds=1&groupIds=2&dateStart=%d&dateEnd=%d&statuses=success&statuses=absent&statuses=exception"
	url = fmt.Sprintf(url, start, end)

	req, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	ginRouter.ServeHTTP(w, req)

	_ = os.Remove("test_statistics.xlsx")

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，得到 %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("期望Content-Type为application/vnd.openxmlformats-officedocument.spreadsheetml.sheet，得到 %s", contentType)
	}

	contentDisposition := w.Header().Get("Content-Disposition")
	if contentDisposition != "attachment; filename=statistics.xlsx" {
		t.Errorf("期望Content-Disposition为attachment; filename=statistics.xlsx，得到 %s", contentDisposition)
	}

	body := w.Body.Bytes()
	if len(body) < 4 || string(body[:4]) != "PK\x03\x04" {
		t.Error("响应内容不是有效的XLSX文件")
	}
}
