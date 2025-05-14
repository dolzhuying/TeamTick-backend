package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	"TeamTickBackend/pkg/logger"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	appErrors "TeamTickBackend/pkg/errors"

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
				return "", appErrors.ErrStatisticsGroupNotFound.WithError(err)
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

		for i, header := range headers {
			cell := fmt.Sprintf("%c1", 'A'+i)
			f.SetCellValue(sheetName, cell, header)
		}

		// 获取组内所有成员
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

		row := 2 // 从第2行开始写入数据
		for _, member := range members {
			statistics, err := s.GetGroupMemberSignInStatistics(ctx, groupID, member.UserID, startTime, endTime)
			if err != nil {
				logger.Error("获取成员统计数据失败：数据库操作错误",
					zap.Int("groupID", groupID),
					zap.Int("userID", member.UserID),
					zap.Error(err),
				)
				return "", appErrors.ErrStatisticsQueryFailed.WithError(err)
			}

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

			for i, value := range data {
				cell := fmt.Sprintf("%c%d", 'A'+i, row)
				f.SetCellValue(sheetName, cell, value)
			}
			row++
		}

		// 设置列宽
		f.SetColWidth(sheetName, "A", fmt.Sprintf("%c", 'A'+len(headers)-1), 15)

		// 写总计行
		sumRow := row + 1
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", sumRow), "总计")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", sumRow), "") // B列空
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", sumRow), "") // C列空
		colIdx := 3                                               // 从D列开始（因为C列是总任务数）
		for range statuses {
			if row == 2 {
				f.SetCellValue(sheetName, fmt.Sprintf("%c%d", 'A'+colIdx, sumRow), 0)
				f.SetCellValue(sheetName, fmt.Sprintf("%c%d", 'A'+colIdx+1, sumRow), "")
			} else {
				formula := fmt.Sprintf("SUM(%c2:%c%d)", 'A'+colIdx, 'A'+colIdx, row-1)
				f.SetCellFormula(sheetName, fmt.Sprintf("%c%d", 'A'+colIdx, sumRow), formula)
				f.SetCellValue(sheetName, fmt.Sprintf("%c%d", 'A'+colIdx+1, sumRow), "")
			}
			colIdx += 2
		}

		// 调试输出
		fmt.Printf("sheetName=%s, row=%d, sumRow=%d\n", sheetName, row, sumRow)

		// 显式设置当前 sheet 为活动 sheet
		idx, _ := f.GetSheetIndex(sheetName)
		f.SetActiveSheet(idx)

		// 柱状图：每种状态一条线，显示所有用户的数据
		quotedSheet := fmt.Sprintf("'%s'", sheetName)
		series := []excelize.ChartSeries{}
		for i := range statuses {
			col := 'D' + i // 从D列开始，因为C列是总任务数
			var catRange, valRange string
			if row-2 == 1 {
				catRange = fmt.Sprintf("%s!$B$2", quotedSheet)
				valRange = fmt.Sprintf("%s!$%c$2", quotedSheet, col)
			} else {
				catRange = fmt.Sprintf("%s!$B$2:$B$%d", quotedSheet, row-1)
				valRange = fmt.Sprintf("%s!$%c$2:$%c$%d", quotedSheet, col, col, row-1)
			}
			series = append(series, excelize.ChartSeries{
				Name:       fmt.Sprintf("%s!$%c$1", quotedSheet, col),
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
			GapWidth: func() *uint { v := uint(500); return &v }(),
		}
		f.AddChart(sheetName, "J2", barChart)
		f.SetCellValue(sheetName, "J18", "用户签到统计")

		// 饼图：用总计行，显示所有用户的总计数据
		var pieCatRange, pieValRange string
		if len(statuses) == 1 {
			pieCatRange = fmt.Sprintf("%s!$D$%d", quotedSheet, sumRow)
			pieValRange = fmt.Sprintf("%s!$D$%d", quotedSheet, sumRow)
		} else {
			pieCatRange = fmt.Sprintf("%s!$D$%d:$%c$%d", quotedSheet, sumRow, 'D'+len(statuses)-1, sumRow)
			pieValRange = fmt.Sprintf("%s!$D$%d:$%c$%d", quotedSheet, sumRow, 'D'+len(statuses)-1, sumRow)
		}
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
		}
		f.AddChart(sheetName, "J20", pieChart)
		f.SetCellValue(sheetName, "J36", "签到状态分布")
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
	if err := f.SaveAs(filePath); err != nil {
		logger.Error("保存Excel文件失败",
			zap.String("filePath", filePath),
			zap.Error(err),
		)
		return "", appErrors.ErrStatisticsFileSaveFailed.WithError(err)
	}

	logger.Info("成功生成用户组签到统计Excel",
		zap.String("filePath", filePath),
	)
	return filePath, nil
}
