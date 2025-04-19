package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	"context"
	"fmt"
	"time"

	appErrors "TeamTickBackend/pkg/errors"

	"gorm.io/gorm"
)

type TaskService struct {
	taskDao            dao.TaskDAO
	taskRecordDao      dao.TaskRecordDAO
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

// 通过GroupID查询签到任务，包含状态筛选
func (s *TaskService) GetTasksByGroupID(ctx context.Context, groupID int, status string) ([]*models.Task, error) {
	var tasks []*models.Task

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		var groupsTasks []*models.Task
		switch status {
		case "active":
			groupsTasks, err = s.taskDao.GetActiveTasksByGroupID(ctx, groupID, tx)
		case "ended":
			groupsTasks, err = s.taskDao.GetEndedTasksByGroupID(ctx, groupID, tx)
		default:
			groupsTasks, err = s.taskDao.GetByGroupID(ctx, groupID, tx)
		}
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

// 通过UserID查询签到任务，包含状态筛选
func (s *TaskService) GetTasksByUserID(ctx context.Context, userID int, status string) ([]*models.Task, error) {
	var tasks []*models.Task

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		var userTasks []*models.Task
		switch status {
		case "active":
			userTasks, err = s.taskDao.GetActiveTasksByUserID(ctx, userID, tx)
		case "ended":
			userTasks, err = s.taskDao.GetEndedTasksByUserID(ctx, userID, tx)
		case "all":
			userTasks, err = s.taskDao.GetByUserID(ctx, userID, tx)
		}
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
			return appErrors.ErrTaskNotFound.WithError(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return task, nil
}

// 执行签到任务，校验逻辑待完善（gps，人脸等）
func (s *TaskService) CheckInTask(ctx context.Context,
	taskID, userID, groupID int,
	latitude, longitude float64,
	signedInTime time.Time,
	username string,
) (*models.TaskRecord, error) {
	var taskRecord models.TaskRecord
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		// 检查任务记录
		task, err := s.taskDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			return appErrors.ErrTaskNotFound.WithError(err)
		}
		//检查签到时间是否过期
		if signedInTime.After(task.EndTime) {
			return appErrors.ErrTaskHasEnded.WithError(fmt.Errorf("task has ended"))
		}

		//检查是否已经签到
		if record, err := s.taskRecordDao.GetByTaskIDAndUserID(ctx, taskID, userID, tx); err == nil && record != nil {
			return appErrors.ErrTaskRecordAlreadyExists.WithError(err)
		}

		//创建签到记录，字段根据校验策略进行选择
		var record models.TaskRecord
		record=models.TaskRecord{
			TaskID:taskID,
			UserID:userID,
			GroupID:groupID,
			SignedTime:signedInTime,
			Username:username,
		}

		//校验策略选择
		if task.GPS {
			// todo 检查gps定位范围
		}
		if task.Face {
			//todo 人脸识别
		}
		if task.WiFi {
			//todo 检查wifi
		}
		if task.NFC {
			//todo 检查nfc
		}

		//创建签到记录
		if err := s.taskRecordDao.Create(ctx, &record, tx); err != nil {
			return appErrors.ErrTaskRecordCreationFailed.WithError(err)
		}
		taskRecord = record
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
