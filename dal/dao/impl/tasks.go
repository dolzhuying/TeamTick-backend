package impl

import (
	"TeamTickBackend/dal/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type TaskDAOMySQLImpl struct {
	DB *gorm.DB
}

// Create 创建签到任务
func (dao *TaskDAOMySQLImpl) Create(ctx context.Context, task *models.Task, tx ...*gorm.DB) error {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(task).Error
}

// GetByGroupID 按group_id查询所有签到任务（包含进行中以及已结束）
func (dao *TaskDAOMySQLImpl) GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.Task, error) {
	var tasks []*models.Task
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("group_id = ?", groupID).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetActiveTasksByGroupID 按group_id查询当前进行中的签到任务
func (dao *TaskDAOMySQLImpl) GetActiveTasksByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.Task, error) {
	var tasks []*models.Task
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("group_id=? AND NOW() BETWEEN start_time AND end_time", groupID).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetEndedTasksByGroupID 按group_id查询当前已结束的签到任务
func (dao *TaskDAOMySQLImpl) GetEndedTasksByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.Task, error) {
	var tasks []*models.Task
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	now := time.Now()
	err := db.WithContext(ctx).Where("group_id=? AND end_time<?", groupID, now).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetByUserID 获取用户当前所属的所有用户组的签到任务
func (dao *TaskDAOMySQLImpl) GetByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.Task, error) {
	var tasks []*models.Task
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Table("tasks t").
		Select("t.*").
		Joins("JOIN group_member gm ON t.group_id = gm.group_id").
		Where("gm.user_id = ?", userID).
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetEndedTasksByUserID 获取用户当前所属的所有用户组的已结束任务
func (dao *TaskDAOMySQLImpl) GetEndedTasksByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.Task, error) {
	var tasks []*models.Task
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Table("tasks t").
		Select("t.*").
		Joins("JOIN group_member gm ON t.group_id = gm.group_id").
		Where("gm.user_id = ? AND t.end_time<?", userID, time.Now()).
		Order("t.end_time ASC").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetActiveTasksByUserID 获取用户当前所属的所有用户组的待签到任务
func (dao *TaskDAOMySQLImpl) GetActiveTasksByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.Task, error) {
	var tasks []*models.Task
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Table("tasks t").
		Select("t.*").
		Joins("JOIN group_member gm ON t.group_id = gm.group_id").
		Where("gm.user_id = ? AND NOW() BETWEEN t.start_time AND t.end_time", userID).
		Order("t.end_time ASC").
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetByTaskID 按task_id查询签到任务
func (dao *TaskDAOMySQLImpl) GetByTaskID(ctx context.Context, taskID int, tx ...*gorm.DB) (*models.Task, error) {
	var task models.Task
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("task_id=?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTask 更新签到任务
func (dao *TaskDAOMySQLImpl) UpdateTask(ctx context.Context, taskID int, newTask *models.Task, tx ...*gorm.DB) error {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	mp := map[string]interface{}{
		"task_name":   newTask.TaskName,
		"description": newTask.Description,
		"start_time":  newTask.StartTime,
		"end_time":    newTask.EndTime,
		"latitude":    newTask.Latitude,
		"longitude":   newTask.Longitude,
		"radius":      newTask.Radius,
		"gps":         newTask.GPS,
		"face":        newTask.Face,
		"wifi":        newTask.WiFi,
		"nfc":         newTask.NFC,
	}
	if newTask.SSID != "" {
		mp["ssid"] = newTask.SSID
	}
	if newTask.BSSID != "" {
		mp["bssid"] = newTask.BSSID
	}
	if newTask.TagID != "" {
		mp["tagid"] = newTask.TagID
	}
	if newTask.TagName != "" {
		mp["tagname"] = newTask.TagName
	}
	return db.WithContext(ctx).Model(&models.Task{}).Where("task_id=?", taskID).Updates(mp).Error
}

// Delete 删除签到任务
func (dao *TaskDAOMySQLImpl) Delete(ctx context.Context, taskID int, tx ...*gorm.DB) error {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Where("task_id=?", taskID).Delete(&models.Task{}).Error
}