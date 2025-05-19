package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	appErrors "TeamTickBackend/pkg/errors"
	"context"
	"errors"
	"time"

	"TeamTickBackend/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuditRequestService struct {
	transactionManager       dao.TransactionManager
	checkApplicationDAO      dao.CheckApplicationDAO
	taskRecordDAO            dao.TaskRecordDAO
	taskDAO                  dao.TaskDAO
	groupDAO                 dao.GroupDAO
	checkApplicationRedisDAO dao.CheckApplicationRedisDAO
}

func NewAuditRequestService(
	transactionManager dao.TransactionManager,
	checkApplicationDAO dao.CheckApplicationDAO,
	taskRecordDAO dao.TaskRecordDAO,
	taskDAO dao.TaskDAO,
	groupDAO dao.GroupDAO,
	checkApplicationRedisDAO dao.CheckApplicationRedisDAO,
) *AuditRequestService {
	return &AuditRequestService{
		transactionManager,
		checkApplicationDAO,
		taskRecordDAO,
		taskDAO,
		groupDAO,
		checkApplicationRedisDAO,
	}
}

// GetAuditRequestListByUserID 获取用户签到申请列表
func (s *AuditRequestService) GetAuditRequestListByUserID(ctx context.Context, userID int) ([]*models.CheckApplication, error) {
	// 先从缓存中获取
	existRequests, err := s.checkApplicationRedisDAO.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("获取用户签到申请缓存失败",
			zap.Int("userID", userID),
			zap.String("operation", "GetByUserID"),
			zap.Error(err),
		)
	}
	if existRequests != nil && len(existRequests) > 0 {
		return existRequests, nil
	}

	// 从数据库中查询
	var requests []*models.CheckApplication

	err = s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		auditRequests, err := s.checkApplicationDAO.GetByUserID(ctx, userID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取用户签到申请失败：未找到记录",
					zap.Int("userID", userID),
					zap.String("operation", "GetByUserID"),
					zap.Error(err),
				)
				return appErrors.ErrAuditRequestNotFound
			}
			logger.Error("获取用户签到申请失败：数据库操作错误",
				zap.Int("userID", userID),
				zap.String("operation", "GetByUserID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		requests = auditRequests
		logger.Info("成功获取用户签到申请列表",
			zap.Int("userID", userID),
			zap.Int("requestCount", len(requests)),
			zap.String("operation", "GetAuditRequestListByUserID"),
		)
		return nil
	})
	if err != nil {
		logger.Error("获取用户签到申请事务失败",
			zap.Int("userID", userID),
			zap.String("operation", "GetAuditRequestListByUserIDTransaction"),
			zap.Error(err),
		)
		return nil, err
	}

	// 将结果缓存
	err = s.checkApplicationRedisDAO.SetByUserID(ctx, userID, requests)
	if err != nil {
		logger.Error("缓存用户签到申请失败",
			zap.Int("userID", userID),
			zap.String("operation", "SetByUserID"),
			zap.Error(err),
		)
	}
	return requests, nil
}

// GetAuditRequestByGroupID 获取组签到申请列表
func (s *AuditRequestService) GetAuditRequestByGroupID(ctx context.Context, groupID int) ([]*models.CheckApplication, error) {
	// 先从缓存中获取
	existRequests, err := s.checkApplicationRedisDAO.GetByGroupID(ctx, groupID)
	if err != nil {
		logger.Error("获取组签到申请缓存失败",
			zap.Int("groupID", groupID),
			zap.String("operation", "GetByGroupID"),
			zap.Error(err),
		)
	}
	if existRequests != nil && len(existRequests) > 0 {
		return existRequests, nil
	}

	// 从数据库中查询
	var requests []*models.CheckApplication
	err = s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		auditRequests, err := s.checkApplicationDAO.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取组签到申请失败：未找到记录",
					zap.Int("groupID", groupID),
					zap.String("operation", "GetByGroupID"),
					zap.Error(err),
				)
				return appErrors.ErrAuditRequestNotFound
			}
			logger.Error("获取组签到申请失败：数据库操作错误",
				zap.Int("groupID", groupID),
				zap.String("operation", "GetByGroupID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		requests = auditRequests
		logger.Info("成功获取组签到申请列表",
			zap.Int("groupID", groupID),
			zap.Int("requestCount", len(requests)),
			zap.String("operation", "GetAuditRequestByGroupID"),
		)
		return nil
	})
	if err != nil {
		logger.Error("获取组签到申请事务失败",
			zap.Int("groupID", groupID),
			zap.String("operation", "GetAuditRequestByGroupIDTransaction"),
			zap.Error(err),
		)
		return nil, err
	}

	// 缓存结果
	err = s.checkApplicationRedisDAO.SetByGroupID(ctx, groupID, requests)
	if err != nil {
		logger.Error("缓存组签到申请失败",
			zap.Int("groupID", groupID),
			zap.String("operation", "SetByGroupID"),
			zap.Error(err),
		)
	}
	return requests, nil
}

// GetAuditRequestByGroupIDWithStatus 获取组签到申请列表，支持状态筛选
func (s *AuditRequestService) GetAuditRequestByGroupIDWithStatus(ctx context.Context, groupID int, status string) ([]*models.CheckApplication, error) {
	// 先从缓存中获取
	existRequests, err := s.checkApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, status)
	if err != nil {
		logger.Error("获取组签到申请缓存失败",
			zap.Int("groupID", groupID),
			zap.String("status", status),
			zap.String("operation", "GetByGroupIDAndStatus"),
			zap.Error(err),
		)
	}
	if existRequests != nil && len(existRequests) > 0 {
		return existRequests, nil
	}

	// 从数据库中查询
	var requests []*models.CheckApplication
	err = s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		auditRequests, err := s.checkApplicationDAO.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取组签到申请失败：未找到记录",
					zap.Int("groupID", groupID),
					zap.String("status", status),
					zap.String("operation", "GetByGroupIDWithStatus"),
					zap.Error(err),
				)
				return appErrors.ErrAuditRequestNotFound
			}
			logger.Error("获取组签到申请失败：数据库操作错误",
				zap.Int("groupID", groupID),
				zap.String("status", status),
				zap.String("operation", "GetByGroupIDWithStatus"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		// 根据状态筛选
		if status == "all" {
			requests = auditRequests
		} else if status == "pending" {
			for _, req := range auditRequests {
				if req.Status == "pending" {
					requests = append(requests, req)
				}
			}
		} else if status == "processed" {
			for _, req := range auditRequests {
				if req.Status == "approved" || req.Status == "rejected" {
					requests = append(requests, req)
				}
			}
		}
		logger.Info("成功获取组签到申请列表（带状态）",
			zap.Int("groupID", groupID),
			zap.String("status", status),
			zap.Int("requestCount", len(requests)),
			zap.String("operation", "GetAuditRequestByGroupIDWithStatus"),
		)
		return nil
	})
	if err != nil {
		logger.Error("获取组签到申请（带状态）事务失败",
			zap.Int("groupID", groupID),
			zap.String("status", status),
			zap.String("operation", "GetAuditRequestByGroupIDWithStatusTransaction"),
			zap.Error(err),
		)
		return nil, err
	}

	// 将结果缓存
	err = s.checkApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, status, requests)
	if err != nil {
		logger.Error("缓存组签到申请（带状态）失败",
			zap.Int("groupID", groupID),
			zap.String("status", status),
			zap.String("operation", "SetByGroupIDAndStatus"),
			zap.Error(err),
		)
	}
	return requests, nil
}

// CreateAuditRequest 创建签到申请
func (s *AuditRequestService) CreateAuditRequest(
	ctx context.Context,
	taskID, userID int,
	username, reason string,
	image string,
) (*models.CheckApplication, error) {
	var request models.CheckApplication
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		// 查询任务
		task, err := s.taskDAO.GetByTaskID(ctx, taskID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("创建签到申请失败：任务不存在",
					zap.Int("taskID", taskID),
					zap.Int("userID", userID),
					zap.String("username", username),
					zap.String("operation", "GetByTaskID"),
					zap.Error(err),
				)
				return appErrors.ErrTaskNotFound
			}
			logger.Error("创建签到申请失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.Int("userID", userID),
				zap.String("username", username),
				zap.String("operation", "GetByTaskID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		// 查询是否存在未审批申请
		existingRequest, err := s.checkApplicationDAO.GetByTaskIDAndUserID(ctx, taskID, userID, tx)
		if err == nil && existingRequest != nil && existingRequest.Status == "pending" {
			logger.Error("创建签到申请失败：已存在未审批申请",
				zap.Int("taskID", taskID),
				zap.Int("userID", userID),
				zap.String("username", username),
				zap.String("operation", "GetByTaskIDAndUserID"),
			)
			return appErrors.ErrAuditRequestAlreadyExists
		}
		// 查询组
		group, err := s.groupDAO.GetByGroupID(ctx, task.GroupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("创建签到申请失败：群组不存在",
					zap.Int("taskID", taskID),
					zap.Int("userID", userID),
					zap.String("username", username),
					zap.Int("groupID", task.GroupID),
					zap.String("operation", "GetByGroupID"),
					zap.Error(err),
				)
				return appErrors.ErrGroupNotFound
			}
			logger.Error("创建签到申请失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.Int("userID", userID),
				zap.String("username", username),
				zap.Int("groupID", task.GroupID),
				zap.String("operation", "GetByGroupID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		newRequest := models.CheckApplication{
			TaskID:        taskID,
			GroupID:       task.GroupID,
			TaskName:      task.TaskName,
			UserID:        userID,
			Reason:        reason,
			Image:         image,
			Username:      username,
			RequestAt:     time.Now(),
			AdminID:       group.CreatorID,
			AdminUsername: group.CreatorName,
			CreatedAt:     time.Now(),
		}
		// 创建申请
		if err := s.checkApplicationDAO.Create(ctx, &newRequest, tx); err != nil {
			logger.Error("创建签到申请失败：数据库操作错误",
				zap.Int("taskID", taskID),
				zap.Int("userID", userID),
				zap.String("username", username),
				zap.String("operation", "Create"),
				zap.Error(err),
			)
			return appErrors.ErrAuditRequestCreateFailed.WithError(err)
		}
		request = newRequest
		logger.Info("成功创建签到申请",
			zap.Int("taskID", taskID),
			zap.Int("userID", userID),
			zap.String("username", username),
			zap.Int("groupID", task.GroupID),
			zap.String("groupName", group.GroupName),
			zap.String("reason", reason),
			zap.String("operation", "CreateAuditRequest"),
		)
		return nil

	})
	if err != nil {
		logger.Error("创建签到申请事务失败",
			zap.Int("taskID", taskID),
			zap.Int("userID", userID),
			zap.String("username", username),
			zap.String("operation", "CreateAuditRequestTransaction"),
			zap.Error(err),
		)
		return nil, err
	}

	// 将记录缓存

	// 缓存用户签到申请
	records, err := s.checkApplicationRedisDAO.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("获取用户签到申请缓存失败",
			zap.Int("userID", userID),
			zap.String("operation", "GetByUserID"),
			zap.Error(err),
		)
	}
	if records == nil {
		records = []*models.CheckApplication{}
	}
	records = append(records, &request)
	if err := s.checkApplicationRedisDAO.SetByUserID(ctx, userID, records); err != nil {
		logger.Error("缓存签到申请失败",
			zap.Int("userID", userID),
			zap.String("operation", "SetByUserID"),
			zap.Error(err),
		)
	}

	// 缓存组签到申请
	records, err = s.checkApplicationRedisDAO.GetByGroupID(ctx, request.GroupID)
	if err != nil {
		logger.Error("获取组签到申请缓存失败",
			zap.Int("groupID", request.GroupID),
			zap.String("operation", "GetByGroupID"),
			zap.Error(err),
		)
	}
	if records == nil {
		records = []*models.CheckApplication{}
	}
	records = append(records, &request)
	if err := s.checkApplicationRedisDAO.SetByGroupID(ctx, request.GroupID, records); err != nil {
		logger.Error("缓存组签到申请失败",
			zap.Int("groupID", request.GroupID),
			zap.String("operation", "SetByGroupID"),
			zap.Error(err),
		)
	}

	// 缓存组签到申请（审核中）
	records, err = s.checkApplicationRedisDAO.GetByGroupIDAndStatus(ctx, request.GroupID, "pending")
	if err != nil {
		logger.Error("获取审核中的组签到申请缓存失败",
			zap.Int("groupID", request.GroupID),
			zap.String("operation", "GetByGroupIDAndStatus"),
			zap.Error(err),
		)
	}
	if records == nil {
		records = []*models.CheckApplication{}
	}
	records = append(records, &request)
	if err := s.checkApplicationRedisDAO.SetByGroupIDAndStatus(ctx, request.GroupID, "pending", records); err != nil {
		logger.Error("缓存组签到申请（审核中）失败",
			zap.Int("groupID", request.GroupID),
			zap.String("operation", "SetByGroupIDAndStatus"),
			zap.Error(err),
		)
	}

	return &request, nil

}

// UpdateAuditRequest 更新签到申请
func (s *AuditRequestService) UpdateAuditRequest(ctx context.Context, requestID int, action string) error {
	var req *models.CheckApplication
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		// 查询申请
		request, err := s.checkApplicationDAO.GetByID(ctx, requestID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("更新签到申请失败：申请不存在",
					zap.Int("requestID", requestID),
					zap.String("action", action),
					zap.String("operation", "GetByID"),
					zap.Error(err),
				)
				return appErrors.ErrAuditRequestNotFound
			}
			logger.Error("更新签到申请失败：数据库操作错误",
				zap.Int("requestID", requestID),
				zap.String("action", action),
				zap.String("operation", "GetByID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		req = request
		switch action {
		case "approve":
			if err := s.checkApplicationDAO.Update(ctx, "approved", requestID, tx); err != nil {
				logger.Error("审批签到申请失败：更新状态失败",
					zap.Int("requestID", requestID),
					zap.String("action", action),
					zap.String("operation", "UpdateApprove"),
					zap.Error(err),
				)
				return appErrors.ErrAuditRequestUpdateFailed.WithError(err)
			}
			// 创建签到记录
			record := models.TaskRecord{
				TaskID:     request.TaskID,
				UserID:     request.UserID,
				Username:   request.Username,
				SignedTime: time.Now(),
				Status:     2,
				CreatedAt:  time.Now(),
			}
			if err := s.taskRecordDAO.Create(ctx, &record, tx); err != nil {
				logger.Error("审批签到申请失败：创建签到记录失败",
					zap.Int("requestID", requestID),
					zap.String("action", action),
					zap.String("operation", "CreateTaskRecord"),
					zap.Error(err),
				)
				return appErrors.ErrTaskRecordCreationFailed.WithError(err)
			}
			logger.Info("成功审批签到申请（通过）",
				zap.Int("requestID", requestID),
				zap.String("action", action),
				zap.String("operation", "UpdateAuditRequestApprove"),
			)
			return nil
		case "reject":
			if err := s.checkApplicationDAO.Update(ctx, "rejected", requestID, tx); err != nil {
				logger.Error("审批签到申请失败：更新状态失败",
					zap.Int("requestID", requestID),
					zap.String("action", action),
					zap.String("operation", "UpdateReject"),
					zap.Error(err),
				)
				return appErrors.ErrAuditRequestUpdateFailed.WithError(err)
			}
			logger.Info("成功审批签到申请（拒绝）",
				zap.Int("requestID", requestID),
				zap.String("action", action),
				zap.String("operation", "UpdateAuditRequestReject"),
			)
		}
		return nil
	})
	if err != nil {
		logger.Error("更新签到申请事务失败",
			zap.Int("requestID", requestID),
			zap.String("action", action),
			zap.String("operation", "UpdateAuditRequestTransaction"),
			zap.Error(err),
		)
		return err
	}

	// 删除缓存

	// 更新用户签到申请缓存
	records, err := s.checkApplicationRedisDAO.GetByUserID(ctx, req.UserID)
	if records != nil && err == nil {
		for i, record := range records {
			if record.ID == requestID {
				records = append(records[:i], records[i+1:]...)
				break
			}
		}
		if err := s.checkApplicationRedisDAO.SetByUserID(ctx, req.UserID, records); err != nil {
			logger.Error("更新用户签到申请缓存失败",
				zap.Int("userID", req.UserID),
				zap.String("operation", "SetByUserID"),
				zap.Error(err),
			)
		}
	}

	// 更新组签到申请缓存
	records, err = s.checkApplicationRedisDAO.GetByGroupID(ctx, req.GroupID)
	if records != nil && err == nil {
		for i, record := range records {
			if record.ID == requestID {
				records = append(records[:i], records[i+1:]...)
				break
			}
		}
		if err := s.checkApplicationRedisDAO.SetByGroupID(ctx, req.GroupID, records); err != nil {
			logger.Error("更新组签到申请缓存失败",
				zap.Int("groupID", req.GroupID),
				zap.String("operation", "SetByGroupID"),
				zap.Error(err),
			)
		}
	}

	// 更新组签到申请（审核中）缓存
	records, err = s.checkApplicationRedisDAO.GetByGroupIDAndStatus(ctx, req.GroupID, "pending")
	if records != nil && err == nil {
		for i, record := range records {
			if record.ID == requestID {
				records = append(records[:i], records[i+1:]...)
				break
			}
		}
		if err := s.checkApplicationRedisDAO.SetByGroupIDAndStatus(ctx, req.GroupID, "pending", records); err != nil {
			logger.Error("更新组签到申请（审核中）缓存失败",
				zap.Int("groupID", req.GroupID),
				zap.String("operation", "SetByGroupIDAndStatus"),
				zap.Error(err),
			)
		}
	}
	return nil
}

// GetGroupIDByAuditRequestID 获取签到申请的组ID
func (s *AuditRequestService) GetGroupIDByAuditRequestID(ctx context.Context, requestID int) (int, error) {
	var groupID int
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		request, err := s.checkApplicationDAO.GetByID(ctx, requestID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取签到申请组ID失败：申请不存在",
					zap.Int("requestID", requestID),
					zap.String("operation", "GetByID"),
					zap.Error(err),
				)
				return appErrors.ErrAuditRequestNotFound
			}
			logger.Error("获取签到申请组ID失败：数据库操作错误",
				zap.Int("requestID", requestID),
				zap.String("operation", "GetByID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		groupID = request.GroupID
		logger.Info("成功获取签到申请组ID",
			zap.Int("requestID", requestID),
			zap.Int("groupID", groupID),
			zap.String("operation", "GetGroupIDByAuditRequestID"),
		)
		return nil
	})
	if err != nil {
		logger.Error("获取签到申请组ID事务失败",
			zap.Int("requestID", requestID),
			zap.String("operation", "GetGroupIDByAuditRequestIDTransaction"),
			zap.Error(err),
		)
		return 0, err
	}
	return groupID, nil
}
