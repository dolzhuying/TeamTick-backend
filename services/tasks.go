package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	"context"
	"errors"
	"math"
	"time"

	appErrors "TeamTickBackend/pkg/errors"

	"TeamTickBackend/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TaskService struct {
	taskDao            dao.TaskDAO
	taskRecordDao      dao.TaskRecordDAO
	groupDao           dao.GroupDAO
	taskRedisDao       dao.TaskRedisDAO
	taskRecordRedisDao dao.TaskRecordRedisDAO
	transactionManager dao.TransactionManager
	groupMemberDao     dao.GroupMemberDAO
}

func NewTaskService(
	taskDao dao.TaskDAO,
	taskRecordDao dao.TaskRecordDAO,
	taskRecordRedisDao dao.TaskRecordRedisDAO,
	taskRedisDao dao.TaskRedisDAO,
	transactionManager dao.TransactionManager,
	groupDao dao.GroupDAO,
	groupMemberDao dao.GroupMemberDAO,
) *TaskService {
	return &TaskService{
		taskDao:            taskDao,
		taskRecordDao:      taskRecordDao,
		taskRecordRedisDao: taskRecordRedisDao,
		taskRedisDao:       taskRedisDao,
		transactionManager: transactionManager,
		groupDao:           groupDao,
		groupMemberDao:     groupMemberDao,
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
	ssid, bssid, tagId, tagName string,
) (*models.Task, error) {
	var createdTask models.Task
	var members []*models.GroupMember

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
			CreatedAt:   time.Now(),
		}

		if wifi {
			task.BSSID = bssid
			task.SSID = ssid
		}

		if nfc {
			task.TagID = tagId
			if tagName != "" {
				task.TagName = tagName
			}
		}
		// 获取用户组成员列表，用于写入缓存
		groupMembers, err := s.groupMemberDao.GetMembersByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("创建任务失败：群组成员不存在",
					zap.Int("groupID", groupID),
					zap.String("operation", "GetMembersByGroupID"),
					zap.Error(err),
				)
				return appErrors.ErrGroupMemberNotFound
			}
			logger.Error("创建任务失败：数据库操作错误",
				zap.Int("groupID", groupID),
				zap.String("operation", "GetMembersByGroupID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		members = groupMembers

		if err := s.taskDao.Create(ctx, &task, tx); err != nil {
			logger.Error("创建任务失败：数据库操作错误",
				zap.String("taskName", taskName),
				zap.Int("groupID", groupID),
				zap.String("operation", "Create"),
				zap.Error(err),
			)
			return appErrors.ErrTaskCreationFailed.WithError(err)
		}
		createdTask = task
		logger.Info("成功创建任务",
			zap.String("taskName", taskName),
			zap.Int("taskID", createdTask.TaskID),
			zap.Int("groupID", groupID),
			zap.Time("startTime", startTime),
			zap.Time("endTime", endTime),
			zap.Float64("latitude", latitude),
			zap.Float64("longitude", longitude),
			zap.Int("radius", radius),
			zap.Bool("gps", gps),
			zap.Bool("face", face),
			zap.Bool("wifi", wifi),
			zap.Bool("nfc", nfc),
			zap.String("operation", "CreateTask"),
		)
		return nil
	})
	if err != nil {
		logger.Error("创建任务事务失败",
			zap.String("taskName", taskName),
			zap.Int("groupID", groupID),
			zap.String("operation", "CreateTaskTransaction"),
			zap.Error(err),
		)
		return nil, err
	}

	// 将任务写入缓存
	if err := s.taskRedisDao.SetByTaskID(ctx, createdTask.TaskID, &createdTask); err != nil {
		logger.Error("任务缓存失败：Redis操作错误",
			zap.Int("taskID", createdTask.TaskID),
			zap.String("operation", "SetByTaskID"),
			zap.Error(err),
		)
	}

	// 写入用户组任务缓存
	taskList, _ := s.taskRedisDao.GetByGroupID(ctx, createdTask.GroupID)
	if taskList == nil {
		taskList = []*models.Task{}
	}
	taskList = append(taskList, &createdTask)
	if err := s.taskRedisDao.SetByGroupID(ctx, createdTask.GroupID, taskList); err != nil {
		logger.Error("用户组任务缓存失败：Redis操作错误",
			zap.Int("groupID", createdTask.GroupID),
			zap.String("operation", "SetByGroupID"),
			zap.Error(err),
		)
	}

	// 写入用户组成员任务缓存
	for _, member := range members {
		taskList, _ := s.taskRedisDao.GetByUserID(ctx, member.UserID)
		if taskList == nil {
			taskList = []*models.Task{}
		}
		taskList = append(taskList, &createdTask)
		if err := s.taskRedisDao.SetByUserID(ctx, member.UserID, taskList); err != nil {
			logger.Error("用户组成员任务缓存失败：Redis操作错误",
				zap.Int("userID", member.UserID),
				zap.String("operation", "SetByUserID"),
				zap.Error(err),
			)
		}
	}

	return &createdTask, nil
}

// 通过GroupID查询签到任务
func (s *TaskService) GetTasksByGroupID(ctx context.Context, groupID int) ([]*models.Task, error) {
	// 先查缓存
	existTasks, err := s.taskRedisDao.GetByGroupID(ctx, groupID)
	if err == nil && existTasks != nil && len(existTasks) > 0 {
		return existTasks, nil
	}

	// 缓存未命中，从数据库查询
	var tasks []*models.Task

	err = s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		var groupsTasks []*models.Task
		groupsTasks, err = s.taskDao.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取群组任务失败：任务不存在",
					zap.Int("groupID", groupID),
					zap.String("operation", "GetByGroupID"),
					zap.Error(err),
				)
				return appErrors.ErrTaskNotFound
			}
			logger.Error("获取群组任务失败：数据库操作错误",
				zap.Int("groupID", groupID),
				zap.String("operation", "GetByGroupID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}

		tasks = groupsTasks
		logger.Info("成功获取群组任务列表",
			zap.Int("groupID", groupID),
			zap.Int("taskCount", len(tasks)),
			zap.String("operation", "GetTasksByGroupID"),
		)
		return nil
	})
	if err != nil {
		logger.Error("获取群组任务事务失败",
			zap.Int("groupID", groupID),
			zap.String("operation", "GetTasksByGroupIDTransaction"),
			zap.Error(err),
		)
		return nil, err
	}

	// 将任务写入缓存
	if err := s.taskRedisDao.SetByGroupID(ctx, groupID, tasks); err != nil {
		logger.Error("群组任务缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.String("operation", "SetByGroupID"),
			zap.Error(err),
		)
	}

	return tasks, nil
}

// 通过UserID查询签到任务
func (s *TaskService) GetTasksByUserID(ctx context.Context, userID int) ([]*models.Task, error) {
	// 通过缓存查询用户任务
	existTasks, err := s.taskRedisDao.GetByUserID(ctx, userID)
	if err == nil && existTasks != nil && len(existTasks) > 0 {
		return existTasks, nil
	}

	// 缓存未命中，从数据库查询
	var tasks []*models.Task

	err = s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		var userTasks []*models.Task
		userTasks, err = s.taskDao.GetByUserID(ctx, userID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取用户任务失败：任务不存在",
					zap.Int("userID", userID),
					zap.String("operation", "GetByUserID"),
					zap.Error(err),
				)
				return appErrors.ErrTaskNotFound
			}
			logger.Error("获取用户任务失败：数据库操作错误",
				zap.Int("userID", userID),
				zap.String("operation", "GetByUserID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		tasks = userTasks
		logger.Info("成功获取用户任务列表",
			zap.Int("userID", userID),
			zap.Int("taskCount", len(tasks)),
			zap.String("operation", "GetTasksByUserID"),
		)
		return nil
	})
	if err != nil {
		logger.Error("获取用户任务事务失败",
			zap.Int("userID", userID),
			zap.String("operation", "GetTasksByUserIDTransaction"),
			zap.Error(err),
		)
		return nil, err
	}

	// 将任务写入用户缓存
	if err := s.taskRedisDao.SetByUserID(ctx, userID, tasks); err != nil {
		logger.Error("用户任务缓存失败：Redis操作错误",
			zap.Int("userID", userID),
			zap.String("operation", "SetByUserID"),
			zap.Error(err),
		)
	}

	return tasks, nil
}

// 通过TaskID查询指定签到任务
func (s *TaskService) GetTaskByTaskID(ctx context.Context, taskID int) (*models.Task, error) {
	// 先查缓存
	existTask, err := s.taskRedisDao.GetByTaskID(ctx, taskID)
	if err == nil && existTask != nil {
		return existTask, nil
	}

	// 缓存未命中，从数据库查询
	var task *models.Task

	err = s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		task, err = s.taskDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取任务失败：任务不存在",
					zap.Int("taskID", taskID),
					zap.String("operation", "GetByTaskID"),
					zap.Error(err),
				)
				return appErrors.ErrTaskNotFound
			}
			logger.Error("获取任务失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.String("operation", "GetByTaskID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		logger.Info("成功获取任务信息",
			zap.Int("taskID", taskID),
			zap.String("taskName", task.TaskName),
			zap.Int("groupID", task.GroupID),
			zap.Time("startTime", task.StartTime),
			zap.Time("endTime", task.EndTime),
			zap.Float64("latitude", task.Latitude),
			zap.Float64("longitude", task.Longitude),
			zap.Int("radius", task.Radius),
			zap.Bool("gps", task.GPS),
			zap.Bool("face", task.Face),
			zap.Bool("wifi", task.WiFi),
			zap.Bool("nfc", task.NFC),
			zap.String("operation", "GetTaskByTaskID"),
		)
		return nil
	})
	if err != nil {
		logger.Error("获取任务事务失败",
			zap.Int("taskID", taskID),
			zap.String("operation", "GetTaskByTaskIDTransaction"),
			zap.Error(err),
		)
		return nil, err
	}

	// 将任务写入缓存
	if err := s.taskRedisDao.SetByTaskID(ctx, taskID, task); err != nil {
		logger.Error("任务缓存失败：Redis操作错误",
			zap.Int("taskID", taskID),
			zap.String("operation", "SetByTaskID"),
			zap.Error(err),
		)
	}

	return task, nil
}

// 验证用户位置是否在任务范围内
func (s *TaskService) VerifyLocation(ctx context.Context, latitude, longitude float64, taskID int) bool {
	var isValid bool

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		task, err := s.taskDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("位置验证失败：任务不存在",
					zap.Int("taskID", taskID),
					zap.String("operation", "GetByTaskID"),
					zap.Error(err),
				)
				return appErrors.ErrTaskNotFound
			}
			logger.Error("位置验证失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.String("operation", "GetByTaskID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
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
		logger.Info("位置验证完成",
			zap.Int("taskID", taskID),
			zap.Float64("distance", distance),
			zap.Int("radius", radius),
			zap.Bool("isValid", isValid),
			zap.Float64("userLatitude", latitude),
			zap.Float64("userLongitude", longitude),
			zap.Float64("taskLatitude", taskLatitude),
			zap.Float64("taskLongitude", taskLongitude),
			zap.String("operation", "VerifyLocation"),
		)
		return nil
	})
	if err != nil {
		logger.Error("位置验证事务失败",
			zap.Int("taskID", taskID),
			zap.String("operation", "VerifyLocationTransaction"),
			zap.Error(err),
		)
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
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("NFC验证失败：任务不存在",
					zap.Int("taskID", taskID),
					zap.String("operation", "GetByTaskID"),
					zap.Error(err),
				)
				return appErrors.ErrTaskNotFound
			}
			logger.Error("NFC验证失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.String("operation", "GetByTaskID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		taskTagName, taskTagID := task.TagName, task.TagID
		isValid = tagID == taskTagID && tagName == taskTagName
		logger.Info("NFC验证完成",
			zap.Int("taskID", taskID),
			zap.String("inputTagID", tagID),
			zap.String("inputTagName", tagName),
			zap.String("taskTagID", taskTagID),
			zap.String("taskTagName", taskTagName),
			zap.Bool("isValid", isValid),
			zap.String("operation", "VerifyNFC"),
		)
		return nil
	})
	if err != nil {
		logger.Error("NFC验证事务失败",
			zap.Int("taskID", taskID),
			zap.String("operation", "VerifyNFCTransaction"),
			zap.Error(err),
		)
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
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("WiFi验证失败：任务不存在",
					zap.Int("taskID", taskID),
					zap.String("operation", "GetByTaskID"),
					zap.Error(err),
				)
				return appErrors.ErrTaskNotFound
			}
			logger.Error("WiFi验证失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.String("operation", "GetByTaskID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		taskSSID, taskBSSID := task.SSID, task.BSSID
		isValid = ssid == taskSSID && bssid == taskBSSID
		logger.Info("WiFi验证完成",
			zap.Int("taskID", taskID),
			zap.String("inputSSID", ssid),
			zap.String("inputBSSID", bssid),
			zap.String("taskSSID", taskSSID),
			zap.String("taskBSSID", taskBSSID),
			zap.Bool("isValid", isValid),
			zap.String("operation", "VerifyWiFi"),
		)
		return nil
	})
	if err != nil {
		logger.Error("WiFi验证事务失败",
			zap.Int("taskID", taskID),
			zap.String("operation", "VerifyWiFiTransaction"),
			zap.Error(err),
		)
		return false
	}
	return isValid
}

// 签到记录写入
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
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("签到失败：任务不存在",
					zap.Int("taskID", taskID),
					zap.Int("userID", userID),
					zap.String("operation", "GetByTaskID"),
					zap.Error(err),
				)
				return appErrors.ErrTaskNotFound
			}
			logger.Error("签到失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.Int("userID", userID),
				zap.String("operation", "GetByTaskID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}

		//时间校验
		if signedInTime.Before(task.StartTime) {
			return appErrors.ErrTaskNotInRange
		}
		if signedInTime.After(task.EndTime) {
			return appErrors.ErrTaskHasEnded
		}

		//检查用户是否已签到
		record, err := s.taskRecordDao.GetByTaskIDAndUserID(ctx, taskID, userID, tx)
		if err == nil && record != nil {
			logger.Error("签到失败：用户已签到",
				zap.Int("taskID", taskID),
				zap.Int("userID", userID),
				zap.Time("signedTime", record.SignedTime),
				zap.String("operation", "GetByTaskIDAndUserID"),
				zap.Error(err),
			)
			return appErrors.ErrTaskRecordAlreadyExists
		}

		group, err := s.groupDao.GetByGroupID(ctx, task.GroupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("签到失败：群组不存在",
					zap.Int("taskID", taskID),
					zap.Int("userID", userID),
					zap.Int("groupID", task.GroupID),
					zap.String("operation", "GetByGroupID"),
					zap.Error(err),
				)
				return appErrors.ErrGroupNotFound
			}
			logger.Error("签到失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.Int("userID", userID),
				zap.Int("groupID", task.GroupID),
				zap.String("operation", "GetByGroupID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		createdTaskRecord := models.TaskRecord{
			TaskID:     taskID,
			TaskName:   task.TaskName,
			GroupID:    task.GroupID,
			UserID:     userID,
			Latitude:   latitude,
			Longitude:  longitude,
			GroupName:  group.GroupName,
			SignedTime: signedInTime,
			CreatedAt:  time.Now(),
		}

		if err := s.taskRecordDao.Create(ctx, &createdTaskRecord, tx); err != nil {
			logger.Error("签到失败：创建签到记录失败",
				zap.Int("taskID", taskID),
				zap.Int("userID", userID),
				zap.String("operation", "Create"),
				zap.Error(err),
			)
			return appErrors.ErrTaskRecordCreationFailed.WithError(err)
		}
		taskRecord = createdTaskRecord

		logger.Info("成功创建签到记录",
			zap.Int("taskID", taskID),
			zap.Int("userID", userID),
			zap.Int("groupID", task.GroupID),
			zap.String("groupName", group.GroupName),
			zap.Float64("latitude", latitude),
			zap.Float64("longitude", longitude),
			zap.Time("signedTime", signedInTime),
			zap.String("operation", "CheckInTask"),
		)
		return nil
	})

	if err != nil {
		logger.Error("签到事务失败",
			zap.Int("taskID", taskID),
			zap.Int("userID", userID),
			zap.String("operation", "CheckInTaskTransaction"),
			zap.Error(err),
		)
		return nil, err
	}

	// 数据库操作成功后，更新/写入缓存

	// 缓存单个新创建的记录
	if err := s.taskRecordRedisDao.SetTaskIDAndUserID(ctx, &taskRecord); err != nil {
		logger.Error("签到记录缓存失败：Redis操作错误",
			zap.Int("taskID", taskID),
			zap.Int("userID", userID),
			zap.String("operation", "SetTaskIDAndUserID"),
			zap.Error(err),
		)
	}

	// 更新任务的签到记录列表缓存
	taskRecordsList, _ := s.taskRecordRedisDao.GetByTaskID(ctx, taskID)
	if taskRecordsList == nil {
		taskRecordsList = []*models.TaskRecord{}
	}
	taskRecordsList = append(taskRecordsList, &taskRecord)
	if err := s.taskRecordRedisDao.SetByTaskID(ctx, taskID, taskRecordsList); err != nil {
		logger.Error("任务签到记录列表缓存失败：Redis操作错误",
			zap.Int("taskID", taskID),
			zap.String("operation", "SetByTaskID"),
			zap.Error(err),
		)
	}

	// 更新用户的签到记录列表缓存
	userRecordsList, _ := s.taskRecordRedisDao.GetByUserID(ctx, userID)
	if userRecordsList == nil {
		userRecordsList = []*models.TaskRecord{}
	}
	userRecordsList = append(userRecordsList, &taskRecord)
	if err := s.taskRecordRedisDao.SetByUserID(ctx, userID, userRecordsList); err != nil {
		logger.Error("用户签到记录列表缓存失败：Redis操作错误",
			zap.Int("userID", userID),
			zap.String("operation", "SetByUserID"),
			zap.Error(err),
		)
	}

	// 更新用户组成员的任务列表缓存
	taskList, err := s.taskRedisDao.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("获取用户组成员任务列表缓存失败：Redis操作错误",
			zap.Int("userID", userID),
			zap.String("operation", "GetByUserID"),
			zap.Error(err),
		)
	}
	if taskList != nil && len(taskList) > 0 {
		for i, task := range taskList {
			if task.TaskID == taskID {
				taskList = append(taskList[:i], taskList[i+1:]...)
				break
			}
		}
		if err := s.taskRedisDao.SetByUserID(ctx, userID, taskList); err != nil {
			logger.Error("用户组成员任务列表缓存失败：Redis操作错误",
				zap.Int("userID", userID),
				zap.String("operation", "SetByUserID"),
				zap.Error(err),
			)
		}
	}

	return &taskRecord, nil
}

// 通过TaskID查询特定任务签到记录，待完善
func (s *TaskService) GetTaskRecordsByTaskID(ctx context.Context, taskID int) ([]*models.TaskRecord, error) {
	// 先尝试从缓存获取
	records, err := s.taskRecordRedisDao.GetByTaskID(ctx, taskID)
	if err == nil && records != nil && len(records) > 0 {
		return records, nil
	}

	// 缓存未命中，从数据库查询
	var taskRecords []*models.TaskRecord
	err = s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		taskRecords, err = s.taskRecordDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取任务签到记录失败：任务不存在",
					zap.Int("taskID", taskID),
					zap.String("operation", "GetByTaskID"),
					zap.Error(err),
				)
				return appErrors.ErrTaskNotFound
			}
			logger.Error("获取任务签到记录失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.String("operation", "GetByTaskID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		logger.Info("成功获取任务签到记录",
			zap.Int("taskID", taskID),
			zap.Int("recordCount", len(taskRecords)),
			zap.String("operation", "GetTaskRecordsByTaskID"),
		)
		return nil
	})
	if err != nil {
		logger.Error("获取任务签到记录事务失败",
			zap.Int("taskID", taskID),
			zap.String("operation", "GetTaskRecordsByTaskIDTransaction"),
			zap.Error(err),
		)
		return nil, err
	}

	// 将结果写入缓存
	if err := s.taskRecordRedisDao.SetByTaskID(ctx, taskID, taskRecords); err != nil {
		logger.Error("任务签到记录缓存失败：Redis操作错误",
			zap.Int("taskID", taskID),
			zap.String("operation", "SetByTaskID"),
			zap.Error(err),
		)
	}

	return taskRecords, nil
}

// 查询个人签到历史记录(涉及多种查询策略，考虑多个函数实现，待完善)
func (s *TaskService) GetTaskRecordsByUserID(ctx context.Context, userID int) ([]*models.TaskRecord, error) {
	// 先尝试从缓存获取
	records, err := s.taskRecordRedisDao.GetByUserID(ctx, userID)
	if err == nil && records != nil && len(records) > 0 {
		return records, nil
	}

	// 缓存未命中，从数据库查询
	var taskRecords []*models.TaskRecord
	err = s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		taskRecords, err = s.taskRecordDao.GetByUserID(ctx, userID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取用户签到记录失败：未找到记录",
					zap.Int("userID", userID),
					zap.String("operation", "GetByUserID"),
					zap.Error(err),
				)
				return appErrors.ErrTaskNotFound
			}
			logger.Error("获取用户签到记录失败：数据库操作错误",
				zap.Int("userID", userID),
				zap.String("operation", "GetByUserID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		logger.Info("成功获取用户签到记录",
			zap.Int("userID", userID),
			zap.Int("recordCount", len(taskRecords)),
			zap.String("operation", "GetTaskRecordsByUserID"),
		)
		return nil
	})
	if err != nil {
		logger.Error("获取用户签到记录事务失败",
			zap.Int("userID", userID),
			zap.String("operation", "GetTaskRecordsByUserIDTransaction"),
			zap.Error(err),
		)
		return nil, err
	}

	// 将结果写入缓存
	if err := s.taskRecordRedisDao.SetByUserID(ctx, userID, taskRecords); err != nil {
		logger.Error("用户签到记录缓存失败：Redis操作错误",
			zap.Int("userID", userID),
			zap.String("operation", "SetByUserID"),
			zap.Error(err),
		)
	}

	return taskRecords, nil
}

// 获取指定任务和用户的签到记录
func (s *TaskService) GetTaskRecordByTaskIDAndUserID(ctx context.Context, taskID, userID int) (*models.TaskRecord, error) {
	// 先尝试从缓存获取

	record, err := s.taskRecordRedisDao.GetByTaskIDAndUserID(ctx, taskID, userID)
	if err == nil && record != nil {
		return record, nil
	}

	// 缓存未命中，从数据库查询
	var taskRecord *models.TaskRecord
	err = s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		taskRecord, err = s.taskRecordDao.GetByTaskIDAndUserID(ctx, taskID, userID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取任务签到记录失败：记录不存在",
					zap.Int("taskID", taskID),
					zap.Int("userID", userID),
					zap.String("operation", "GetByTaskIDAndUserID"),
					zap.Error(err),
				)
				return appErrors.ErrTaskNotFound
			}
			logger.Error("获取任务签到记录失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.Int("userID", userID),
				zap.String("operation", "GetByTaskIDAndUserID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		logger.Info("成功获取任务签到记录",
			zap.Int("taskID", taskID),
			zap.Int("userID", userID),
			zap.Time("signedTime", taskRecord.SignedTime),
			zap.Float64("latitude", taskRecord.Latitude),
			zap.Float64("longitude", taskRecord.Longitude),
			zap.String("groupName", taskRecord.GroupName),
			zap.String("operation", "GetTaskRecordByTaskIDAndUserID"),
		)
		return nil
	})
	if err != nil {
		logger.Error("获取任务签到记录事务失败",
			zap.Int("taskID", taskID),
			zap.Int("userID", userID),
			zap.String("operation", "GetTaskRecordByTaskIDAndUserIDTransaction"),
			zap.Error(err),
		)
		return nil, err
	}

	// 将结果写入缓存
	if err = s.taskRecordRedisDao.SetTaskIDAndUserID(ctx, taskRecord); err != nil {
		logger.Error("任务签到记录缓存失败：Redis操作错误",
			zap.Int("taskID", taskID),
			zap.Int("userID", userID),
			zap.String("operation", "SetTaskIDAndUserID"),
			zap.Error(err),
		)
	}

	return taskRecord, nil
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
	var newTask models.Task
	var members []*models.GroupMember
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		task, err := s.taskDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("更新任务失败：任务不存在",
					zap.Int("taskID", taskID),
					zap.String("operation", "GetByTaskID"),
					zap.Error(err),
				)
				return appErrors.ErrTaskNotFound
			}
			logger.Error("更新任务失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.String("operation", "GetByTaskID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}

		// 获取用户组成员列表
		nowMembers, err := s.groupMemberDao.GetMembersByGroupID(ctx, task.GroupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("更新任务失败：群组成员不存在",
					zap.Int("taskID", taskID),
					zap.String("operation", "GetMembersByGroupID"),
					zap.Error(err),
				)
				return appErrors.ErrGroupMemberNotFound
			}
			logger.Error("更新任务失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.String("operation", "GetMembersByGroupID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		members = nowMembers

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
			UpdatedAt:   time.Now(),
		}
		if len(wifiAndNFCInfo) > 0 {
			//todo 根据info选择字段
		}
		//更新签到任务
		if err := s.taskDao.UpdateTask(ctx, taskID, newTask, tx); err != nil {
			logger.Error("更新任务失败：更新数据库失败",
				zap.Int("taskID", taskID),
				zap.String("operation", "UpdateTask"),
				zap.Error(err),
			)
			return appErrors.ErrTaskUpdateFailed.WithError(err)
		}
		//获取更新后的任务
		nowTask, err := s.taskDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			logger.Error("获取更新后的任务失败：Redis操作错误",
				zap.Int("taskID", taskID),
				zap.String("operation", "GetByTaskIDAfterUpdate"),
				zap.Error(err),
			)
		}
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("更新任务失败：获取更新后的任务失败",
					zap.Int("taskID", taskID),
					zap.String("operation", "GetByTaskIDAfterUpdate"),
					zap.Error(err),
				)
				return appErrors.ErrTaskNotFound
			}
			logger.Error("更新任务失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.String("operation", "GetByTaskIDAfterUpdate"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		newTask = nowTask
		logger.Info("成功更新任务",
			zap.Int("taskID", taskID),
			zap.String("taskName", taskName),
			zap.Time("startTime", startTime),
			zap.Time("endTime", endTime),
			zap.Float64("latitude", latitude),
			zap.Float64("longitude", longitude),
			zap.Int("radius", radius),
			zap.Bool("gps", gps),
			zap.Bool("face", face),
			zap.Bool("wifi", wifi),
			zap.Bool("nfc", nfc),
			zap.String("operation", "UpdateTask"),
		)
		return nil
	})
	if err != nil {
		logger.Error("更新任务事务失败",
			zap.Int("taskID", taskID),
			zap.String("operation", "UpdateTaskTransaction"),
			zap.Error(err),
		)
		return nil, err
	}

	//将任务从缓存中删除
	if err := s.taskRedisDao.DeleteCacheByTaskID(ctx, taskID); err != nil {
		logger.Error("删除任务缓存失败：Redis操作错误",
			zap.Int("taskID", taskID),
			zap.String("operation", "DeleteCacheByTaskID"),
			zap.Error(err),
		)
	}
	if err := s.taskRedisDao.DeleteCacheByGroupID(ctx, newTask.GroupID); err != nil {
		logger.Error("删除群组任务缓存失败：Redis操作错误",
			zap.Int("groupID", newTask.GroupID),
			zap.String("operation", "DeleteCacheByGroupID"),
			zap.Error(err),
		)
	}
	for _, member := range members {
		if err := s.taskRedisDao.DeleteCacheByUserID(ctx, member.UserID); err != nil {
			logger.Error("删除用户任务缓存失败：Redis操作错误",
				zap.Int("userID", member.UserID),
				zap.String("operation", "DeleteCacheByUserID"),
				zap.Error(err),
			)
		}
	}

	return &newTask, nil
}

// 删除签到任务
func (s *TaskService) DeleteTask(ctx context.Context, taskID int) error {
	var groupID int
	var members []*models.GroupMember
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//先查询任务获取组ID
		task, err := s.taskDao.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("删除任务失败：任务不存在",
					zap.Int("taskID", taskID),
					zap.String("operation", "GetByTaskID"),
					zap.Error(err),
				)
				return appErrors.ErrTaskNotFound
			}
			logger.Error("删除任务失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.String("operation", "GetByTaskID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		groupID = task.GroupID
		// 获取用户组成员列表
		groupMembers, err := s.groupMemberDao.GetMembersByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("删除任务失败：群组成员不存在",
					zap.Int("groupID", groupID),
					zap.String("operation", "GetMembersByGroupID"),
					zap.Error(err),
				)
				return appErrors.ErrGroupMemberNotFound
			}
			logger.Error("删除任务失败：数据库操作错误",
				zap.Int("groupID", groupID),
				zap.String("operation", "GetMembersByGroupID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		members = groupMembers

		if err := s.taskDao.Delete(ctx, taskID, tx); err != nil {
			logger.Error("删除任务失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.String("operation", "Delete"),
				zap.Error(err),
			)
			return appErrors.ErrTaskDeleteFailed.WithError(err)
		}
		logger.Info("成功删除任务",
			zap.Int("taskID", taskID),
			zap.Int("groupID", groupID),
			zap.String("operation", "DeleteTask"),
		)
		return nil
	})
	if err != nil {
		logger.Error("删除任务事务失败",
			zap.Int("taskID", taskID),
			zap.String("operation", "DeleteTaskTransaction"),
			zap.Error(err),
		)
		return err
	}

	//将任务从缓存中删除
	if err := s.taskRedisDao.DeleteCacheByTaskID(ctx, taskID); err != nil {
		logger.Error("删除任务缓存失败：Redis操作错误",
			zap.Int("taskID", taskID),
			zap.String("operation", "DeleteCacheByTaskID"),
			zap.Error(err),
		)
	}
	if err := s.taskRedisDao.DeleteCacheByGroupID(ctx, groupID); err != nil {
		logger.Error("删除群组任务缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.String("operation", "DeleteCacheByGroupID"),
			zap.Error(err),
		)
	}
	for _, member := range members {
		if err := s.taskRedisDao.DeleteCacheByUserID(ctx, member.UserID); err != nil {
			logger.Error("删除用户任务缓存失败：Redis操作错误",
				zap.Int("userID", member.UserID),
				zap.String("operation", "DeleteCacheByUserID"),
				zap.Error(err),
			)
		}
	}

	return nil
}
