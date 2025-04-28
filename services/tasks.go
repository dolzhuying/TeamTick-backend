package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	"context"
	"errors"
	"math"
	"time"

	appErrors "TeamTickBackend/pkg/errors"

	"gorm.io/gorm"
)

type TaskService struct {
	taskDao            dao.TaskDAO
	taskRecordDao      dao.TaskRecordDAO
	groupDao           dao.GroupDAO
	transactionManager dao.TransactionManager
}

func NewTaskService(
	taskDao dao.TaskDAO,
	taskRecordDao dao.TaskRecordDAO,
	transactionManager dao.TransactionManager) *TaskService {
	return &TaskService{
		taskDao:            taskDao,
		taskRecordDao:      taskRecordDao,
		transactionManager: transactionManager,
	}
}

// 创建签到任务
func (s *TaskService) CreateTask(ctx context.Context,
	taskName string,
	description string,
	groupID int,
	startTime time.Time,
	endTime time.Time,
	latitude float64,
	longitude float64,
	radius int,
	gps, face, wifi, nfc bool,
	wifiAndNFCInfo ...string,
) (*models.Task, error) {
	var createdTask models.Task

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		task := models.Task{
			TaskName:    taskName,
			Description: description,
			GroupID:     groupID,
			StartTime:   startTime,
			EndTime:     endTime,
			Latitude:    latitude,
			Longitude:   longitude,
			Radius:      radius,
			GPS:         gps,
			Face:        face,
			WiFi:        wifi,
			NFC:         nfc,
		}
		//根据info选择字段，待完善
		//n:=len(wifiAndNFCInfo)

		if err := s.taskDao.Create(ctx, &task, tx); err != nil {
			return appErrors.ErrTaskCreationFailed.WithError(err)
		}
		createdTask = task
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &createdTask, nil
}

// 通过GroupID查询签到任务
func (s *TaskService) GetTasksByGroupID(ctx context.Context, groupID int) ([]*models.Task, error) {
	var tasks []*models.Task

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		var groupsTasks []*models.Task
		groupsTasks, err = s.taskDao.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			return appErrors.ErrTaskNotFound.WithError(err)
		}

		tasks = groupsTasks
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// 通过UserID查询签到任务
func (s *TaskService) GetTasksByUserID(ctx context.Context, userID int) ([]*models.Task, error) {
	var tasks []*models.Task

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		var userTasks []*models.Task
		userTasks, err = s.taskDao.GetByUserID(ctx, userID, tx)
		if err != nil {
			return appErrors.ErrTaskNotFound.WithError(err)
		}
		tasks = userTasks
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// 通过TaskID查询指定签到任务
func (s *TaskService) GetTaskByTaskID(ctx context.Context, taskID int) (*models.Task, error) {
	var task *models.Task

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		task, err = s.taskDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrTaskNotFound
			}
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return task, nil
}

// 验证用户位置是否在任务范围内
func (s *TaskService) VerifyLocation(ctx context.Context, latitude, longitude float64, taskID int) bool {
	var isValid bool

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		task, err := s.taskDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			return appErrors.ErrTaskNotFound.WithError(err)
		}

		radius := task.Radius
		taskLatitude, taskLongitude := task.Latitude, task.Longitude

		// 使用Haversine公式计算两点之间的距离
		// 地球半径(米)
		const earthRadius = 6371000.0

		lat1 := latitude * (math.Pi / 180.0)
		lon1 := longitude * (math.Pi / 180.0)
		lat2 := taskLatitude * (math.Pi / 180.0)
		lon2 := taskLongitude * (math.Pi / 180.0)

		dLat := lat2 - lat1
		dLon := lon2 - lon1
		a := math.Pow(math.Sin(dLat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dLon/2), 2)
		c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
		distance := earthRadius * c

		isValid = distance <= float64(radius)
		return nil
	})
	if err != nil {
		return false
	}
	return isValid

}

// 验证NFC
func (s *TaskService) VerifyNFC(ctx context.Context, tagID, tagName string, taskID int) bool {
	var isValid bool

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		task, err := s.taskDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			return appErrors.ErrTaskNotFound.WithError(err)
		}
		taskTagName, taskTagID := task.TagName, task.TagID
		isValid = tagID == taskTagID && tagName == taskTagName
		return nil
	})
	if err != nil {
		return false
	}
	return isValid
}

// 验证wifi
func (s *TaskService) VerifyWiFi(ctx context.Context, ssid, bssid string, taskID int) bool {
	var isValid bool

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		task, err := s.taskDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			return appErrors.ErrTaskNotFound.WithError(err)
		}
		taskSSID, taskBSSID := task.SSID, task.BSSID
		isValid = ssid == taskSSID && bssid == taskBSSID
		return nil
	})
	if err != nil {
		return false
	}
	return isValid
}

// 签到记录写入,待完善
func (s *TaskService) CheckInTask(
	ctx context.Context,
	taskID, userID int,
	latitude, longitude float64,
	signedInTime time.Time,
	otherInfo ...string,
) (*models.TaskRecord, error) {
	var taskRecord models.TaskRecord

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		task, err := s.taskDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			return appErrors.ErrTaskNotFound.WithError(err)
		}
		group, err := s.groupDao.GetByGroupID(ctx, task.GroupID, tx)
		if err != nil {
			return appErrors.ErrGroupNotFound.WithError(err)
		}
		createdTaskRecord := models.TaskRecord{
			TaskID:     taskID,
			GroupID:    task.GroupID,
			UserID:     userID,
			Latitude:   latitude,
			Longitude:  longitude,
			GroupName:  group.GroupName,
			SignedTime: signedInTime,
		}

		//根据otherInfo选择字段

		if err := s.taskRecordDao.Create(ctx, &taskRecord, tx); err != nil {
			return appErrors.ErrTaskRecordCreationFailed.WithError(err)
		}
		taskRecord = createdTaskRecord
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &taskRecord, nil

}

// 通过TaskID查询特定任务签到记录，待完善
func (s *TaskService) GetTaskRecordsByTaskID(ctx context.Context, taskID int) ([]*models.TaskRecord, error) {
	var taskRecords []*models.TaskRecord
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		taskRecords, err = s.taskRecordDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			return appErrors.ErrTaskNotFound.WithError(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return taskRecords, nil
}

// 查询个人签到历史记录(涉及多种查询策略，考虑多个函数实现，待完善)
func (s *TaskService) GetTaskRecordsByUserID(ctx context.Context, userID int) ([]*models.TaskRecord, error) {
	var taskRecords []*models.TaskRecord
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		taskRecords, err = s.taskRecordDao.GetByUserID(ctx, userID, tx)
		if err != nil {
			return appErrors.ErrTaskNotFound.WithError(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return taskRecords, nil
}

// 更新签到任务
func (s *TaskService) UpdateTask(
	ctx context.Context,
	taskID int,
	taskName string,
	description string,
	startTime time.Time,
	endTime time.Time,
	latitude float64,
	longitude float64,
	radius int,
	gps, face, wifi, nfc bool,
	wifiAndNFCInfo ...string,
) (*models.Task, error) {
	var task models.Task
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		_, err := s.taskDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			return appErrors.ErrTaskNotFound.WithError(err)
		}
		newTask := &models.Task{
			TaskName:    taskName,
			Description: description,
			StartTime:   startTime,
			EndTime:     endTime,
			Latitude:    latitude,
			Longitude:   longitude,
			Radius:      radius,
			GPS:         gps,
			Face:        face,
			WiFi:        wifi,
			NFC:         nfc,
		}
		if len(wifiAndNFCInfo) > 0 {
			//todo 根据info选择字段
		}
		//更新签到任务
		if err := s.taskDao.UpdateTask(ctx, taskID, newTask, tx); err != nil {
			return appErrors.ErrTaskUpdateFailed.WithError(err)
		}
		//获取更新后的任务
		nowTask, err := s.taskDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		task = *nowTask
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// 删除签到任务
func (s *TaskService) DeleteTask(ctx context.Context, taskID int) error {
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := s.taskDao.Delete(ctx, taskID, tx); err != nil {
			return appErrors.ErrTaskDeleteFailed.WithError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
