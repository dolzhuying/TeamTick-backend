package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	appErrors "TeamTickBackend/pkg/errors"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type AuditRequestService struct {
	transactionManager  dao.TransactionManager
	checkApplicationDAO dao.CheckApplicationDAO
	taskRecordDAO       dao.TaskRecordDAO
	taskDAO             dao.TaskDAO
	groupDAO            dao.GroupDAO
}

func NewAuditRequestService(
	transactionManager dao.TransactionManager,
	checkApplicationDAO dao.CheckApplicationDAO,
	taskRecordDAO dao.TaskRecordDAO,
	taskDAO dao.TaskDAO,
	groupDAO dao.GroupDAO,
) *AuditRequestService {
	return &AuditRequestService{
		transactionManager,
		checkApplicationDAO,
		taskRecordDAO,
		taskDAO,
		groupDAO,
	}
}

// GetAuditRequestListByUserID 获取用户签到申请列表
func (s *AuditRequestService) GetAuditRequestListByUserID(ctx context.Context, userID int) ([]*models.CheckApplication, error) {
	var requests []*models.CheckApplication

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		auditRequests, err := s.checkApplicationDAO.GetByUserID(ctx, userID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrAuditRequestNotFound
			}
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		requests = auditRequests
		return nil
	})
	if err != nil {
		return nil, err
	}
	return requests, nil
}

// GetAuditRequestByGroupID 获取组签到申请列表
func (s *AuditRequestService) GetAuditRequestByGroupID(ctx context.Context, groupID int) ([]*models.CheckApplication, error) {
	var requests []*models.CheckApplication
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		auditRequests, err := s.checkApplicationDAO.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrAuditRequestNotFound
			}
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		requests = auditRequests
		return nil
	})
	if err != nil {
		return nil, err
	}
	return requests, nil
}

// GetAuditRequestByGroupIDWithStatus 获取组签到申请列表，支持状态筛选
func (s *AuditRequestService) GetAuditRequestByGroupIDWithStatus(ctx context.Context, groupID int, status string) ([]*models.CheckApplication, error) {
	var requests []*models.CheckApplication
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		auditRequests, err := s.checkApplicationDAO.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrAuditRequestNotFound
			}
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
		return nil
	})
	if err != nil {
		return nil, err
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
				return appErrors.ErrTaskNotFound
			}
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		// 查询是否存在未审批申请
		existingRequest, err := s.checkApplicationDAO.GetByTaskIDAndUserID(ctx, taskID, userID, tx)
		if err == nil && existingRequest != nil && existingRequest.Status == "pending" {
			return appErrors.ErrAuditRequestAlreadyExists
		}
		// 查询组
		group, err := s.groupDAO.GetByGroupID(ctx, task.GroupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrGroupNotFound
			}
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
		}
		// 创建申请
		if err := s.checkApplicationDAO.Create(ctx, &newRequest, tx); err != nil {
			return appErrors.ErrAuditRequestCreateFailed.WithError(err)
		}
		request = newRequest
		return nil

	})
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// UpdateAuditRequest 更新签到申请
func (s *AuditRequestService) UpdateAuditRequest(ctx context.Context, requestID int, action string) error {
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		// 查询申请
		request, err := s.checkApplicationDAO.GetByID(ctx, requestID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrAuditRequestNotFound
			}
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		switch action {
		case "approve":
			if err := s.checkApplicationDAO.Update(ctx, "approved", requestID, tx); err != nil {
				return appErrors.ErrAuditRequestUpdateFailed.WithError(err)
			}
			// 创建签到记录
			record := models.TaskRecord{
				TaskID:     request.TaskID,
				UserID:     request.UserID,
				Username:   request.Username,
				SignedTime: time.Now(),
				Status:     2,
			}
			if err := s.taskRecordDAO.Create(ctx, &record, tx); err != nil {
				return appErrors.ErrTaskRecordCreationFailed.WithError(err)
			}
			return nil
		case "reject":
			if err := s.checkApplicationDAO.Update(ctx, "rejected", requestID, tx); err != nil {
				return appErrors.ErrAuditRequestUpdateFailed.WithError(err)
			}
		}
		return nil
	})
	if err != nil {
		return err
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
				return appErrors.ErrAuditRequestNotFound
			}
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		groupID = request.GroupID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return groupID, nil
}
