package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type StatisticsService struct {
	statisticsDao dao.StatisticsDAO
	transactionManager dao.TransactionManager
}
// 和接口字段不一致，缺少GroupName
// 这里建议在handlers层首先查询所有用户组，然后循环查询每组的签到统计数据,最后的groupName在handlers层及进行构造补全
type GroupSignInStatistics struct {
	GroupID int 
	SuccessRecords []*models.TaskRecord
	AbsentRecords []*models.AbsentRecord
	ExecptionRecords []*models.CheckApplication
}

// handlers先调用group service获取组内所有成员，再调用本service获取每个成员的签到统计数据
type GroupMemberStatistics struct {
	GroupID int
	UserID int
	SuccessNum int
	AbsentNum int
	ExceptionNum int
}
func NewStatisticsService(
	statisticsDao dao.StatisticsDAO,
	groupDao dao.GroupDAO,
	transactionManager dao.TransactionManager,
) *StatisticsService {
	return &StatisticsService{
		statisticsDao: statisticsDao,
		transactionManager: transactionManager,
	}
}

// 所有查询操作相关错误均定义为500，这里不做额外error包装

// 获取所有用户组
func (s*StatisticsService) GetAllGroups(ctx context.Context) ([]*models.Group,error){
	return s.statisticsDao.GetAllGroups(ctx)
}

// 获取指定组内签到统计数据记录(可在handlers选择构建返回 具体记录/统计数据)
func (s*StatisticsService) GetGroupSignInStatistics(ctx context.Context, groupID int, startTime, endTime time.Time) (*GroupSignInStatistics, error) {
	var dataStatistics *GroupSignInStatistics

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		success,err:=s.statisticsDao.GetGroupSignInSuccess(ctx,groupID,startTime,endTime,tx)
		if err!=nil{
			return err
		}
		absent,err:=s.statisticsDao.GetGroupSignInAbsent(ctx,groupID,startTime,endTime,tx)
		if err!=nil{
			return err
		}
		exception,err:=s.statisticsDao.GetGroupSignInException(ctx,groupID,startTime,endTime,tx)
		if err!=nil{
			return err
		}
		dataStatistics = &GroupSignInStatistics{
			GroupID: groupID,
			SuccessRecords: success,
			AbsentRecords: absent,
			ExecptionRecords: exception,
		}
		return nil
	})
	if err!=nil{
		return nil,err
	}
	return dataStatistics,nil
}

// 获取组内成员签到统计数据
func (s*StatisticsService) GetGroupMemberSignInStatistics(ctx context.Context, groupID,userID int, startTime, endTime time.Time) (*GroupMemberStatistics, error) {
	var dataStatistics *GroupMemberStatistics

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		successNum,err:=s.statisticsDao.GetMemberSignInSuccessNum(ctx,groupID,userID,startTime,endTime,tx)
		if err!=nil{
			return err
		}
		absentNum,err:=s.statisticsDao.GetMemberSignInAbsentNum(ctx,groupID,userID,startTime,endTime,tx)
		if err!=nil{
			return err
		}
		exceptionNum,err:=s.statisticsDao.GetMemberSignInExceptionNum(ctx,groupID,userID,startTime,endTime,tx)
		if err!=nil{
			return err
		}
		dataStatistics = &GroupMemberStatistics{
			GroupID: groupID,
			UserID: userID,
			SuccessNum: successNum,
			AbsentNum: absentNum,
			ExceptionNum: exceptionNum,
		}
		return nil
	})
	if err!=nil{
		return nil,err
	}
	return dataStatistics,nil
}
