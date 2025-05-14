package mysqlImpl

import (
	"TeamTickBackend/dal/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type StatisticsDAOMySQLImpl struct {
	DB *gorm.DB
}

// GetAllGroups 获取所有组
func (dao *StatisticsDAOMySQLImpl) GetAllGroups(ctx context.Context, tx ...*gorm.DB) ([]*models.Group, error) {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	var groups []*models.Group
	err := db.WithContext(ctx).Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// GetGroupSignInSuccess 获取组内签到成功记录
func (dao *StatisticsDAOMySQLImpl) GetGroupSignInSuccess(ctx context.Context, groupID int, startTime, endTime time.Time, tx ...*gorm.DB) ([]*models.TaskRecord, error) {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	var records []*models.TaskRecord
	err := db.WithContext(ctx).
		Model(&models.TaskRecord{}).
		Where("group_id=? AND created_at BETWEEN ? AND ?", groupID, startTime, endTime).
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// GetGroupSignInException 获取组内签到异常记录
func (dao *StatisticsDAOMySQLImpl) GetGroupSignInException(ctx context.Context, groupID int, startTime, endTime time.Time, tx ...*gorm.DB) ([]*models.CheckApplication, error) {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}

	var records []*models.CheckApplication
	err := db.WithContext(ctx).
		Model(&models.CheckApplication{}).
		Where("group_id=?", groupID).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Find(&records).Error

	if err != nil {
		return nil, err
	}
	return records, nil
}

// GetGroupSignInAbsent 获取组内签到缺勤人员记录
func (dao *StatisticsDAOMySQLImpl) GetGroupSignInAbsent(ctx context.Context, groupID int, startTime, endTime time.Time, tx ...*gorm.DB) ([]*models.AbsentRecord, error) {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}

	var records []*models.AbsentRecord
	err := db.WithContext(ctx).
		Table("group_member gm").
		Select("gm.*, t.task_id"). 
		Joins(`
      JOIN tasks t 
        ON gm.group_id = t.group_id
       AND t.start_time >= ? 
       AND t.end_time   <= ? 
       AND t.end_time < NOW()
    `, startTime, endTime).
		Joins(`
      LEFT JOIN tasks_record tr 
        ON tr.task_id = t.task_id 
       AND tr.user_id  = gm.user_id
    `).
		Where("gm.group_id = ?", groupID).
		Where("tr.task_id IS NULL").
		Scan(&records).Error

	if err != nil {
		return nil, err
	}
	return records, nil
}

// 获取成员签到成功次数
func (dao *StatisticsDAOMySQLImpl) GetMemberSignInSuccessNum(ctx context.Context, groupID, userID int, startTime, endTime time.Time, tx ...*gorm.DB) (int, error) {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	var successNum int64
	err := db.WithContext(ctx).
		Model(&models.TaskRecord{}).
		Where("group_id=? AND user_id=? AND created_at BETWEEN ? AND ?", groupID, userID, startTime, endTime).
		Count(&successNum).Error
	if err != nil {
		return -1, err
	}
	return int(successNum), nil
}

// 获取成员签到异常次数
func (dao *StatisticsDAOMySQLImpl) GetMemberSignInExceptionNum(ctx context.Context, groupID, userID int, startTime, endTime time.Time, tx ...*gorm.DB) (int, error) {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	var exceptionNum int64
	err := db.WithContext(ctx).
		Model(&models.CheckApplication{}).
		Where("group_id=? AND user_id=? AND created_at BETWEEN ? AND ?", groupID, userID, startTime, endTime).
		Count(&exceptionNum).Error
	if err != nil {
		return -1, err
	}
	return int(exceptionNum), nil
}

// 获取成员签到缺勤次数
func (dao *StatisticsDAOMySQLImpl) GetMemberSignInAbsentNum(ctx context.Context, groupID, userID int, startTime, endTime time.Time, tx ...*gorm.DB) (int, error) {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	var absentNum int64
	err := db.WithContext(ctx).
		Table("tasks t").
		Where("t.group_id = ?", groupID).
		Where("t.start_time >= ?", startTime).
		Where("t.end_time <= ?", endTime).
		Where("t.end_time < ?", time.Now()).
		Where("NOT EXISTS (SELECT 1 FROM tasks_record tr WHERE tr.task_id = t.task_id AND tr.user_id = ?)", userID).
		Where("EXISTS (SELECT 1 FROM group_member gm WHERE gm.group_id = t.group_id AND gm.user_id = ?)", userID).
		Count(&absentNum).Error

	if err != nil {
		return -1, err
	}
	return int(absentNum), nil
}
