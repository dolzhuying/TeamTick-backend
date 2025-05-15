package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	"TeamTickBackend/pkg/logger"
	"context"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"

	appErrors "TeamTickBackend/pkg/errors"

	"bytes"

	"image/color"

	"github.com/disintegration/imaging"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type StatisticsService struct {
	statisticsDao      dao.StatisticsDAO
	transactionManager dao.TransactionManager
	groupDao           dao.GroupDAO
	groupMemberDao     dao.GroupMemberDAO
}

// 和接口字段不一致，缺少GroupName
// 这里建议在handlers层首先查询所有用户组，然后循环查询每组的签到统计数据,最后的groupName在handlers层及进行构造补全
type GroupSignInStatistics struct {
	GroupID          int
	SuccessRecords   []*models.TaskRecord
	AbsentRecords    []*models.AbsentRecord
	ExecptionRecords []*models.CheckApplication
}

// handlers先调用group service获取组内所有成员，再调用本service获取每个成员的签到统计数据
type GroupMemberStatistics struct {
	GroupID      int
	UserID       int
	SuccessNum   int
	AbsentNum    int
	ExceptionNum int
}

func NewStatisticsService(
	statisticsDao dao.StatisticsDAO,
	groupDao dao.GroupDAO,
	transactionManager dao.TransactionManager,
	groupMemberDao dao.GroupMemberDAO,
) *StatisticsService {
	return &StatisticsService{
		statisticsDao:      statisticsDao,
		transactionManager: transactionManager,
		groupDao:           groupDao,
		groupMemberDao:     groupMemberDao,
	}
}

// 所有查询操作相关错误均定义为500，这里不做额外error包装

// 获取所有用户组
func (s *StatisticsService) GetAllGroups(ctx context.Context) ([]*models.Group, error) {
	groups, err := s.statisticsDao.GetAllGroups(ctx)
	if err != nil {
		logger.Error("获取所有用户组失败：数据库操作错误",
			zap.Error(err),
		)
		return nil, appErrors.ErrDatabaseOperation.WithError(err)
	}
	logger.Info("成功获取所有用户组",
		zap.Int("groupCount", len(groups)),
	)
	return groups, nil
}

// 获取指定组内签到统计数据记录
func (s *StatisticsService) GetGroupSignInStatistics(ctx context.Context, groupID int, startTime, endTime time.Time) (*GroupSignInStatistics, error) {
	if startTime.After(endTime) {
		return nil, appErrors.ErrStatisticsInvalidTimeRange
	}

	var dataStatistics *GroupSignInStatistics

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		success, err := s.statisticsDao.GetGroupSignInSuccess(ctx, groupID, startTime, endTime, tx)
		if err != nil {
			logger.Error("获取组签到成功记录失败",
				zap.Int("groupID", groupID),
				zap.Error(err),
			)
			return appErrors.ErrStatisticsQueryFailed.WithError(err)
		}
		absent, err := s.statisticsDao.GetGroupSignInAbsent(ctx, groupID, startTime, endTime, tx)
		if err != nil {
			logger.Error("获取组缺勤记录失败",
				zap.Int("groupID", groupID),
				zap.Error(err),
			)
			return appErrors.ErrStatisticsQueryFailed.WithError(err)
		}
		exception, err := s.statisticsDao.GetGroupSignInException(ctx, groupID, startTime, endTime, tx)
		if err != nil {
			logger.Error("获取组异常记录失败",
				zap.Int("groupID", groupID),
				zap.Error(err),
			)
			return appErrors.ErrStatisticsQueryFailed.WithError(err)
		}
		dataStatistics = &GroupSignInStatistics{
			GroupID:          groupID,
			SuccessRecords:   success,
			AbsentRecords:    absent,
			ExecptionRecords: exception,
		}
		logger.Info("成功获取组签到统计数据",
			zap.Int("groupID", groupID),
			zap.Int("successCount", len(success)),
			zap.Int("absentCount", len(absent)),
			zap.Int("exceptionCount", len(exception)),
		)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dataStatistics, nil
}

// 获取组内成员签到统计数据
func (s *StatisticsService) GetGroupMemberSignInStatistics(ctx context.Context, groupID, userID int, startTime, endTime time.Time) (*GroupMemberStatistics, error) {
	if startTime.After(endTime) {
		return nil, appErrors.ErrStatisticsInvalidTimeRange
	}

	var dataStatistics *GroupMemberStatistics

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		successNum, err := s.statisticsDao.GetMemberSignInSuccessNum(ctx, groupID, userID, startTime, endTime, tx)
		if err != nil {
			logger.Error("获取成员签到成功次数失败",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Error(err),
			)
			return appErrors.ErrStatisticsQueryFailed.WithError(err)
		}
		absentNum, err := s.statisticsDao.GetMemberSignInAbsentNum(ctx, groupID, userID, startTime, endTime, tx)
		if err != nil {
			logger.Error("获取成员缺勤次数失败",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Error(err),
			)
			return appErrors.ErrStatisticsQueryFailed.WithError(err)
		}
		exceptionNum, err := s.statisticsDao.GetMemberSignInExceptionNum(ctx, groupID, userID, startTime, endTime, tx)
		if err != nil {
			logger.Error("获取成员异常次数失败",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Error(err),
			)
			return appErrors.ErrStatisticsQueryFailed.WithError(err)
		}
		dataStatistics = &GroupMemberStatistics{
			GroupID:      groupID,
			UserID:       userID,
			SuccessNum:   successNum,
			AbsentNum:    absentNum,
			ExceptionNum: exceptionNum,
		}
		logger.Info("成功获取成员签到统计数据",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("successNum", successNum),
			zap.Int("absentNum", absentNum),
			zap.Int("exceptionNum", exceptionNum),
		)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dataStatistics, nil
}

// 生成用户组签到统计Excel文件
func (s *StatisticsService) GenerateGroupSignInStatisticsExcel(ctx context.Context, groupIDs []int, startTime, endTime time.Time, statuses []string) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("导出Excel时panic", zap.Any("recover", r))
		}
	}()

	if startTime.After(endTime) {
		logger.Error("生成统计Excel失败：时间范围无效",
			zap.Time("startTime", startTime),
			zap.Time("endTime", endTime),
		)
		return "", appErrors.ErrStatisticsInvalidTimeRange
	}

	// 如果没有指定状态，使用默认状态列表
	if len(statuses) == 0 {
		statuses = []string{"success", "absent", "exception"}
	}

	logger.Info("开始生成用户组签到统计Excel",
		zap.Ints("groupIDs", groupIDs),
		zap.Time("startTime", startTime),
		zap.Time("endTime", endTime),
		zap.Strings("statuses", statuses),
	)

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			logger.Error("关闭Excel文件失败",
				zap.Error(err),
			)
		}
	}()

	// 遍历每个用户组，为每个组创建一个工作表
	for i, groupID := range groupIDs {
		// 获取用户组信息
		group, err := s.groupDao.GetByGroupID(ctx, groupID, nil)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Warn("获取用户组信息失败：用户组不存在",
					zap.Int("groupID", groupID),
					zap.Error(err),
				)
				continue
			}
			logger.Error("获取用户组信息失败",
				zap.Int("groupID", groupID),
				zap.Error(err),
			)
			return "", appErrors.ErrStatisticsQueryFailed.WithError(err)
		}

		// 设置工作表名称
		sheetName := fmt.Sprintf("%d %s", groupID, group.GroupName)
		if i == 0 {
			// 第一个 sheet 重命名为 "ID 用户组名称"
			f.SetSheetName("Sheet1", sheetName)
		} else {
			// 其他 sheet 创建新 sheet
			f.NewSheet(sheetName)
		}

		// 合并A1:A2，logo区域
		logger.Info("[导出统计] 开始合并A1:A2单元格",
			zap.String("sheet", sheetName),
		)
		if err := f.MergeCell(sheetName, "A1", "A2"); err != nil {
			logger.Error("[导出统计] 合并单元格失败：Excel操作错误",
				zap.String("sheet", sheetName),
				zap.Error(err),
			)
		} else {
			logger.Info("[导出统计] 合并单元格成功",
				zap.String("sheet", sheetName),
			)
		}
		// 设置A列宽和A1/A2行高，使区域与图片canvasSize接近
		logger.Info("[导出统计] 设置A列宽和A1/A2行高",
			zap.String("sheet", sheetName),
		)
		f.SetColWidth(sheetName, "A", "A", 12)
		f.SetRowHeight(sheetName, 1, 40)
		f.SetRowHeight(sheetName, 2, 40)
		// 设置A1:A2单元格内容居中且无边框
		logger.Info("[导出统计] 设置A1:A2单元格样式",
			zap.String("sheet", sheetName),
		)
		centerStyle, _ := f.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{
				Horizontal: "center",
				Vertical:   "center",
			},
			Border: []excelize.Border{
				{Type: "top", Color: "#FFFFFF", Style: 2},
				{Type: "bottom", Color: "#FFFFFF", Style: 2},
				{Type: "left", Color: "#FFFFFF", Style: 2},
				{Type: "right", Color: "#FFFFFF", Style: 2},
			},
		})
		f.SetCellStyle(sheetName, "A1", "A2", centerStyle)
		// 1. 导出时间移到B1
		exportTime := time.Now().Format("2006-01-02 15:04:05")
		logger.Info("[导出统计] 写入导出时间",
			zap.String("exportTime", exportTime),
			zap.String("sheet", sheetName),
		)
		f.SetCellValue(sheetName, "B1", "导出时间："+exportTime)
		logger.Info("已写入导出时间", zap.String("exportTime", exportTime), zap.String("sheet", sheetName))
		// 设置B1单元格样式，去掉右边框和底部边框
		timeStyle, _ := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 11, Color: "#1976D2"},
			Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
			Border: []excelize.Border{
				{Type: "left", Color: "#000000", Style: 1},
				{Type: "top", Color: "#000000", Style: 1},
				// 去掉底部边框
			},
		})
		f.SetCellStyle(sheetName, "B1", "B1", timeStyle)
		// 设置C1单元格样式，去掉左边框和底部边框
		c1Style, _ := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 11, Color: "#1976D2"},
			Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
			Border: []excelize.Border{
				// 去掉左边框
				{Type: "top", Color: "#000000", Style: 1},
				{Type: "right", Color: "#000000", Style: 1},
				// 去掉底部边框
			},
		})
		f.SetCellStyle(sheetName, "C1", "C1", c1Style)
		// canvasSize为正方形
		canvasSize := 80
		logoPath := "src/image.png" // 请确保logo已放在此路径
		logger.Info("[导出统计] 读取logo图片",
			zap.String("logoPath", logoPath),
			zap.String("sheet", sheetName),
		)
		if fileBytes, err := os.ReadFile(logoPath); err == nil {
			logger.Info("[导出统计] logo图片读取成功",
				zap.Int("fileSize", len(fileBytes)),
				zap.String("sheet", sheetName),
			)
			img, err := imaging.Decode(bytes.NewReader(fileBytes))
			if err != nil {
				logger.Error("[导出统计] 解码logo图片失败",
					zap.Error(err),
					zap.String("sheet", sheetName),
				)
			} else {
				logger.Info("[导出统计] logo图片解码成功",
					zap.String("sheet", sheetName),
				)
				bounds := img.Bounds()
				w, h := bounds.Dx(), bounds.Dy()
				var logoImg image.Image
				if w > canvasSize || h > canvasSize {
					scale := float64(canvasSize) / float64(w)
					if h > w {
						scale = float64(canvasSize) / float64(h)
					}
					logger.Info("[导出统计] logo图片缩放",
						zap.Float64("scale", scale),
						zap.String("sheet", sheetName),
					)
					logoImg = imaging.Resize(img, int(float64(w)*scale), int(float64(h)*scale), imaging.Lanczos)
				} else {
					logoImg = img
				}
				canvas := imaging.New(canvasSize, canvasSize, color.NRGBA{255, 255, 255, 255})
				canvas = imaging.PasteCenter(canvas, logoImg)
				buf := new(bytes.Buffer)
				if err := imaging.Encode(buf, canvas, imaging.PNG); err != nil {
					logger.Error("[导出统计] 编码图片失败：图片处理错误",
						zap.Error(err),
						zap.String("sheet", sheetName),
					)
				} else {
					logger.Info("[导出统计] 图片编码成功",
						zap.Int("fileSize", buf.Len()),
						zap.String("sheet", sheetName),
					)
					fileBytes = buf.Bytes()
				}
			}
			pic := &excelize.Picture{
				File:      fileBytes,
				Extension: ".png",
			}
			logger.Info("[导出统计] 准备插入Logo图片",
				zap.String("sheet", sheetName),
			)
			if err := f.AddPictureFromBytes(sheetName, "A1", pic); err != nil {
				logger.Error("[导出统计] 插入Logo失败：Excel操作错误",
					zap.Error(err),
					zap.String("sheet", sheetName),
				)
			} else {
				logger.Info("[导出统计] Logo插入成功",
					zap.String("sheet", sheetName),
				)
			}
		} else {
			logger.Warn("[导出统计] Logo文件不存在或读取失败，跳过插入",
				zap.String("logoPath", logoPath),
				zap.Error(err),
			)
		}

		// 设置表头
		headers := []string{"用户ID", "用户名", "总任务数"}
		for _, status := range statuses {
			switch status {
			case "success":
				headers = append(headers, "成功签到次数", "成功签到任务ID")
			case "absent":
				headers = append(headers, "缺勤次数", "缺勤任务ID")
			case "exception":
				headers = append(headers, "异常次数", "异常任务ID")
			}
		}
		logger.Info("[导出统计] 准备写入表头",
			zap.Strings("headers", headers),
			zap.String("sheet", sheetName),
		)
		headerRow := 3
		dataStartRow := 4
		headerStyle, _ := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{"#466A9E"}, Pattern: 1},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			Border: []excelize.Border{
				{Type: "left", Color: "#466A9E", Style: 1},
				{Type: "top", Color: "#466A9E", Style: 1},
				{Type: "bottom", Color: "#466A9E", Style: 1},
				{Type: "right", Color: "#466A9E", Style: 1},
			},
		})
		oddStyle, _ := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Color: "#222222"},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E3EFFC"}, Pattern: 1},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			Border: []excelize.Border{
				{Type: "left", Color: "#B7CAE3", Style: 1},
				{Type: "top", Color: "#B7CAE3", Style: 1},
				{Type: "bottom", Color: "#B7CAE3", Style: 1},
				{Type: "right", Color: "#B7CAE3", Style: 1},
			},
		})
		evenStyle, _ := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Color: "#222222"},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{"#FFFFFF"}, Pattern: 1},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			Border: []excelize.Border{
				{Type: "left", Color: "#B7CAE3", Style: 1},
				{Type: "top", Color: "#B7CAE3", Style: 1},
				{Type: "bottom", Color: "#B7CAE3", Style: 1},
				{Type: "right", Color: "#B7CAE3", Style: 1},
			},
		})

		// 表头
		for i, header := range headers {
			cell := fmt.Sprintf("%c%d", 'A'+i, headerRow)
			f.SetCellValue(sheetName, cell, header)
			f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}
		f.SetRowHeight(sheetName, headerRow, 25)
		logger.Info("[导出统计] 表头写入完成",
			zap.String("sheet", sheetName),
		)

		// 获取组内所有成员
		logger.Info("准备获取组成员",
			zap.Int("groupID", groupID),
			zap.String("sheet", sheetName),
		)
		members, err := s.groupMemberDao.GetMembersByGroupID(ctx, groupID, nil)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Warn("获取组成员失败：没有成员",
					zap.Int("groupID", groupID),
					zap.Error(err),
				)
				continue
			}
			logger.Error("获取组成员失败",
				zap.Int("groupID", groupID),
				zap.Error(err),
			)
			return "", appErrors.ErrStatisticsQueryFailed.WithError(err)
		}
		logger.Info("获取组成员成功", zap.Int("groupID", groupID), zap.Int("memberCount", len(members)), zap.String("sheet", sheetName))

		row := dataStartRow
		for idx, member := range members {
			logger.Info("[导出统计] 开始处理成员",
				zap.Int("index", idx),
				zap.Int("groupID", groupID),
				zap.Int("userID", member.UserID),
				zap.String("sheet", sheetName),
			)
			statistics, err := s.GetGroupMemberSignInStatistics(ctx, groupID, member.UserID, startTime, endTime)
			if err != nil {
				logger.Error("[导出统计] 获取成员统计数据失败",
					zap.Int("index", idx),
					zap.Int("groupID", groupID),
					zap.Int("userID", member.UserID),
					zap.String("sheet", sheetName),
					zap.Error(err),
				)
				return "", appErrors.ErrStatisticsQueryFailed.WithError(err)
			}
			logger.Info("[导出统计] 成员统计数据获取成功",
				zap.Int("index", idx),
				zap.Int("groupID", groupID),
				zap.Int("userID", member.UserID),
				zap.String("sheet", sheetName),
				zap.Any("statistics", statistics),
			)

			// 计算总任务数
			totalTasks := statistics.SuccessNum + statistics.AbsentNum + statistics.ExceptionNum
			data := []interface{}{member.UserID, member.Username, totalTasks}
			// 统计每种状态的taskId
			for _, status := range statuses {
				taskIDs := ""
				switch status {
				case "success":
					num := statistics.SuccessNum
					records, err := s.statisticsDao.GetGroupSignInSuccess(ctx, groupID, startTime, endTime, nil)
					if err != nil {
						logger.Error("获取成功签到记录失败",
							zap.Int("groupID", groupID),
							zap.Error(err),
						)
						return "", appErrors.ErrStatisticsQueryFailed.WithError(err)
					}
					var ids []string
					for _, r := range records {
						if r.UserID == member.UserID {
							ids = append(ids, fmt.Sprintf("%d", r.TaskID))
						}
					}
					taskIDs = strings.Join(ids, ",")
					data = append(data, num, taskIDs)
				case "absent":
					num := statistics.AbsentNum
					records, err := s.statisticsDao.GetGroupSignInAbsent(ctx, groupID, startTime, endTime, nil)
					if err != nil {
						logger.Error("获取缺勤记录失败",
							zap.Int("groupID", groupID),
							zap.Error(err),
						)
						return "", appErrors.ErrStatisticsQueryFailed.WithError(err)
					}
					var ids []string
					for _, r := range records {
						if r.UserID == member.UserID {
							ids = append(ids, fmt.Sprintf("%d", r.TaskID))
						}
					}
					taskIDs = strings.Join(ids, ",")
					data = append(data, num, taskIDs)
				case "exception":
					num := statistics.ExceptionNum
					records, err := s.statisticsDao.GetGroupSignInException(ctx, groupID, startTime, endTime, nil)
					if err != nil {
						logger.Error("获取异常记录失败",
							zap.Int("groupID", groupID),
							zap.Error(err),
						)
						return "", appErrors.ErrStatisticsQueryFailed.WithError(err)
					}
					var ids []string
					for _, r := range records {
						if r.UserID == member.UserID {
							ids = append(ids, fmt.Sprintf("%d", r.TaskID))
						}
					}
					taskIDs = strings.Join(ids, ",")
					data = append(data, num, taskIDs)
				}
			}

			logger.Info("[导出统计] 成员数据写入完成",
				zap.Int("index", idx),
				zap.Int("groupID", groupID),
				zap.Int("userID", member.UserID),
				zap.String("sheet", sheetName),
			)
			// 写入每个单元格并设置样式
			for i, value := range data {
				cell := fmt.Sprintf("%c%d", 'A'+i, row)
				rowIdx := row - dataStartRow
				var styleID int
				if rowIdx%2 == 0 {
					styleID = oddStyle
				} else {
					styleID = evenStyle
				}
				f.SetCellStyle(sheetName, cell, cell, styleID)
				f.SetCellValue(sheetName, cell, value)
			}
			f.SetRowHeight(sheetName, row, 20)
			row++
		}
		logger.Info("成员数据写入完成", zap.String("sheet", sheetName))

		// 设置列宽
		colWidths := map[string]float64{
			"A": 10, // 用户ID
			"B": 20, // 用户名
			"C": 12, // 总任务数
		}
		// 动态设置状态相关列的宽度
		for i := 0; i < len(statuses); i++ {
			colWidths[string(rune('D'+i*2))] = 12 // 次数列
			colWidths[string(rune('E'+i*2))] = 30 // 任务ID列
		}
		logger.Info("[导出统计] 设置列宽",
			zap.Any("colWidths", colWidths),
			zap.String("sheet", sheetName),
		)
		for col, width := range colWidths {
			f.SetColWidth(sheetName, col, col, width)
		}

		// 写总计行
		logger.Info("[导出统计] 准备写入总计行",
			zap.String("sheet", sheetName),
		)
		sumRow := row + 1
		for i := 0; i < len(headers); i++ {
			cell := fmt.Sprintf("%c%d", 'A'+i, sumRow)
			if i == 0 {
				f.SetCellValue(sheetName, cell, "总计")
			} else if i == 1 {
				f.SetCellValue(sheetName, cell, "")
			} else if i == 2 {
				formula := fmt.Sprintf("SUM(%c%d:%c%d)", 'A'+i, dataStartRow, 'A'+i, row-1)
				f.SetCellFormula(sheetName, cell, formula)
			} else if (i-3)%2 == 0 {
				formula := fmt.Sprintf("SUM(%c%d:%c%d)", 'A'+i, dataStartRow, 'A'+i, row-1)
				f.SetCellFormula(sheetName, cell, formula)
			} else {
				f.SetCellValue(sheetName, cell, "")
			}
			// 总计行只加粗字体
			f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}
		f.SetRowHeight(sheetName, sumRow, 25)
		logger.Info("[导出统计] 总计行写入完成",
			zap.String("sheet", sheetName),
		)

		// 饼图
		logger.Info("[导出统计] 准备生成饼图",
			zap.String("sheet", sheetName),
		)
		pieStartCol := 3 // "成功签到次数"列
		pieEndCol := pieStartCol + len(statuses)*2 - 2
		quotedSheet := fmt.Sprintf("'%s'", sheetName)
		pieCatRange := fmt.Sprintf("%s!$%c$1:$%c$1", quotedSheet, 'A'+pieStartCol, 'A'+pieEndCol)
		pieValRange := fmt.Sprintf("%s!$%c$%d:$%c$%d", quotedSheet, 'A'+pieStartCol, sumRow, 'A'+pieEndCol, sumRow)
		pieChart := &excelize.Chart{
			Type: excelize.Pie,
			Series: []excelize.ChartSeries{
				{
					Name:       "签到分布",
					Categories: pieCatRange,
					Values:     pieValRange,
				},
			},
			PlotArea: excelize.ChartPlotArea{
				ShowPercent: true,
			},
			Title: []excelize.RichTextRun{
				{
					Text: "签到分布饼图",
				},
			},
		}
		_ = f.AddChart(sheetName, "J21", pieChart)
		logger.Info("[导出统计] 饼图添加完成",
			zap.String("sheet", sheetName),
		)

		// 柱状图
		series := []excelize.ChartSeries{}
		for i := range statuses {
			col := 'D' + i // 从D列开始，因为C列是总任务数
			var catRange, valRange string
			if row-2 == 1 {
				catRange = fmt.Sprintf("%s!$B$%d", quotedSheet, dataStartRow)
				valRange = fmt.Sprintf("%s!$%c$%d", quotedSheet, col, dataStartRow)
			} else {
				catRange = fmt.Sprintf("%s!$B$%d:$B$%d", quotedSheet, dataStartRow, row-1)
				valRange = fmt.Sprintf("%s!$%c$%d:$%c$%d", quotedSheet, col, dataStartRow, col, row-1)
			}
			series = append(series, excelize.ChartSeries{
				Name:       fmt.Sprintf("%s!$%c$%d", quotedSheet, col, headerRow),
				Categories: catRange,
				Values:     valRange,
			})
		}
		barChart := &excelize.Chart{
			Type:   excelize.Col,
			Series: series,
			PlotArea: excelize.ChartPlotArea{
				ShowPercent: true,
			},
			GapWidth: func() *uint { v := uint(300); return &v }(),
			Title: []excelize.RichTextRun{
				{
					Text: "各成员签到统计柱状图",
				},
			},
		}
		_ = f.AddChart(sheetName, "J3", barChart)
		logger.Info("[导出统计] 柱状图添加完成",
			zap.String("sheet", sheetName),
		)

		// 冻结表头
		logger.Info("设置冻结窗格", zap.String("sheet", sheetName))
		f.SetPanes(sheetName, &excelize.Panes{
			Freeze:      true,
			Split:       false,
			XSplit:      0,
			YSplit:      1,
			TopLeftCell: "A2",
			ActivePane:  "bottomLeft",
		})

		// 调试输出
		fmt.Printf("sheetName=%s, row=%d, sumRow=%d\n", sheetName, row, sumRow)
		logger.Info("sheet生成完成", zap.String("sheet", sheetName), zap.Int("row", row), zap.Int("sumRow", sumRow))

		// 显式设置当前 sheet 为活动 sheet
		idx, _ := f.GetSheetIndex(sheetName)
		f.SetActiveSheet(idx)
	}

	// 生成文件名
	fileName := fmt.Sprintf("statistics_%s_%s.xlsx",
		startTime.Format("20060102"),
		endTime.Format("20060102"),
	)
	filePath := filepath.Join("export_files", fileName)

	// 确保目录存在
	if err := os.MkdirAll("export_files", 0755); err != nil {
		logger.Error("创建导出目录失败",
			zap.Error(err),
		)
		return "", appErrors.ErrStatisticsFileCreationFailed.WithError(err)
	}

	// 保存文件
	logger.Info("[导出统计] 准备保存Excel文件",
		zap.String("filePath", filePath),
	)
	if err := f.SaveAs(filePath); err != nil {
		logger.Error("保存Excel文件失败",
			zap.String("filePath", filePath),
			zap.Error(err),
		)
		return "", appErrors.ErrStatisticsFileSaveFailed.WithError(err)
	}

	logger.Info("[导出统计] 成功生成用户组签到统计Excel",
		zap.String("filePath", filePath),
	)
	return filePath, nil
}
