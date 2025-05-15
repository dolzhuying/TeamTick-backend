package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	appErrors "TeamTickBackend/pkg/errors"
	"TeamTickBackend/pkg/logger"
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GroupsService struct {
	groupDao                dao.GroupDAO
	groupMemberDao          dao.GroupMemberDAO
	joinApplicationDao      dao.JoinApplicationDAO
	transactionManager      dao.TransactionManager
	groupRedisDAO           dao.GroupRedisDAO
	groupMemberRedisDAO     dao.GroupMemberRedisDAO
	joinApplicationRedisDAO dao.JoinApplicationRedisDAO
	taskRedisDAO            dao.TaskRedisDAO
}

func NewGroupsService(
	groupDao dao.GroupDAO,
	groupMemberDao dao.GroupMemberDAO,
	joinApplicationDao dao.JoinApplicationDAO,
	transactionManager dao.TransactionManager,
	groupRedisDAO dao.GroupRedisDAO,
	groupMemberRedisDAO dao.GroupMemberRedisDAO,
	joinApplicationRedisDAO dao.JoinApplicationRedisDAO,
	taskRedisDAO dao.TaskRedisDAO,
) *GroupsService {

	return &GroupsService{
		groupDao:                groupDao,
		groupMemberDao:          groupMemberDao,
		joinApplicationDao:      joinApplicationDao,
		transactionManager:      transactionManager,
		groupRedisDAO:           groupRedisDAO,
		groupMemberRedisDAO:     groupMemberRedisDAO,
		joinApplicationRedisDAO: joinApplicationRedisDAO,
		taskRedisDAO:            taskRedisDAO,
	}
}

// 创建用户组
func (s *GroupsService) CreateGroup(ctx context.Context, groupName, description, creatorName string, creatorID int) (*models.Group, error) {
	var createdGroup models.Group

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		group := models.Group{
			GroupName:   groupName,
			Description: description,
			CreatorID:   creatorID,
			CreatorName: creatorName,
			CreatedAt:   time.Now(),
		}
		//创建用户组
		if err := s.groupDao.Create(ctx, &group, tx); err != nil {
			logger.Error("创建用户组失败：数据库操作错误",
				zap.String("groupName", groupName),
				zap.Int("creatorID", creatorID),
				zap.Error(err),
			)
			return appErrors.ErrGroupCreationFailed.WithError(err)
		}
		//添加用户组管理员
		if err := s.groupMemberDao.Create(ctx, &models.GroupMember{
			GroupID:   group.GroupID,
			UserID:    creatorID,
			Username:  creatorName,
			Role:      "admin",
			CreatedAt: time.Now(),
		}, tx); err != nil {
			logger.Error("创建用户组失败：添加管理员失败",
				zap.String("groupName", groupName),
				zap.Int("creatorID", creatorID),
				zap.Error(err),
			)
			return appErrors.ErrGroupMemberCreationFailed.WithError(err)
		}
		createdGroup = group
		logger.Info("成功创建用户组",
			zap.String("groupName", groupName),
			zap.Int("groupID", group.GroupID),
			zap.Int("creatorID", creatorID),
		)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 缓存用户组
	if err := s.groupRedisDAO.SetByGroupID(ctx, createdGroup.GroupID, &createdGroup); err != nil {
		logger.Error("用户组缓存失败：Redis操作错误",
			zap.Int("groupID", createdGroup.GroupID),
			zap.String("groupName", createdGroup.GroupName),
			zap.Int("creatorID", createdGroup.CreatorID),
			zap.String("creatorName", createdGroup.CreatorName),
			zap.Error(err),
		)
	}
	return &createdGroup, nil
}

// 通过GroupID查询用户组信息
func (s *GroupsService) GetGroupByGroupID(ctx context.Context, groupID int) (*models.Group, error) {
	// 从缓存中查找用户组
	existGroup, err := s.groupRedisDAO.GetByGroupID(ctx, groupID)
	if err == nil && existGroup != nil {
		logger.Info("从缓存中获取用户组信息成功",
			zap.Int("groupID", groupID),
			zap.String("groupName", existGroup.GroupName),
			zap.Int("creatorID", existGroup.CreatorID),
			zap.String("creatorName", existGroup.CreatorName),
		)
		return existGroup, nil
	}

	// 从数据库中查找用户组
	var group models.Group
	err = s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		Group, err := s.groupDao.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取用户组信息失败：用户组不存在",
					zap.Int("groupID", groupID),
					zap.Error(err),
				)
				return appErrors.ErrGroupNotFound
			}
			logger.Error("获取用户组信息失败：数据库操作错误",
				zap.Int("groupID", groupID),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		group = *Group
		logger.Info("成功从数据库获取用户组信息",
			zap.Int("groupID", groupID),
			zap.String("groupName", group.GroupName),
			zap.Int("creatorID", group.CreatorID),
			zap.String("creatorName", group.CreatorName),
			zap.String("description", group.Description),
		)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 缓存用户组
	if err := s.groupRedisDAO.SetByGroupID(ctx, group.GroupID, &group); err != nil {
		logger.Error("用户组缓存失败：Redis操作错误",
			zap.Int("groupID", group.GroupID),
			zap.String("groupName", group.GroupName),
			zap.Int("creatorID", group.CreatorID),
			zap.String("creatorName", group.CreatorName),
			zap.Error(err),
		)
	}
	return &group, nil
}

// 查询特定用户所创建或加入的所有用户组
func (s *GroupsService) GetGroupsByUserID(ctx context.Context, userID int, filter ...string) ([]*models.Group, error) {
	var groups []*models.Group
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		var Groups []*models.Group
		var err error
		if len(filter) > 0 {
			Groups, err = s.groupDao.GetGroupsByUserIDAndfilter(ctx, userID, filter[0], tx)
		} else {
			Groups, err = s.groupDao.GetGroupsByUserID(ctx, userID, tx)
		}
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取用户组列表失败：未找到用户组",
					zap.Int("userID", userID),
					zap.Strings("filter", filter),
					zap.Error(err),
				)
				return appErrors.ErrGroupNotFound
			}
			logger.Error("获取用户组列表失败：数据库操作错误",
				zap.Int("userID", userID),
				zap.Strings("filter", filter),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		groups = Groups
		logger.Info("成功获取用户组列表",
			zap.Int("userID", userID),
			zap.Int("groupCount", len(groups)),
			zap.Strings("filter", filter),
			zap.Any("groups", groups),
		)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// 更新用户组信息
func (s *GroupsService) UpdateGroup(ctx context.Context, groupID, operatorID int, groupName, description string) (*models.Group, error) {
	var updatedGroup models.Group
	var members []*models.GroupMember
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查操作员权限
		if err := s.CheckMemberPermission(ctx, groupID, operatorID); err != nil {
			logger.Error("更新用户组失败：权限不足",
				zap.Int("groupID", groupID),
				zap.Int("operatorID", operatorID),
				zap.String("groupName", groupName),
				zap.Error(err),
			)
			return appErrors.ErrRolePermissionDenied.WithError(err)
		}
		//更新用户组信息
		if err := s.groupDao.UpdateMessage(ctx, groupID, groupName, description, tx); err != nil {
			logger.Error("更新用户组失败：数据库操作错误",
				zap.Int("groupID", groupID),
				zap.String("groupName", groupName),
				zap.String("description", description),
				zap.Error(err),
			)
			return appErrors.ErrGroupUpdateFailed.WithError(err)
		}
		//查询更新后的用户组信息
		group, err := s.groupDao.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			logger.Error("更新用户组失败：获取更新后的信息失败",
				zap.Int("groupID", groupID),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		// 查询用户组成员列表以删除缓存
		members, err = s.groupMemberDao.GetMembersByGroupID(ctx, groupID, tx)
		if err != nil {
			logger.Error("更新用户组失败：获取成员列表失败",
				zap.Int("groupID", groupID),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		updatedGroup = *group
		logger.Info("成功更新用户组信息",
			zap.Int("groupID", groupID),
			zap.String("groupName", groupName),
			zap.String("description", description),
			zap.Int("operatorID", operatorID),
			zap.Int("memberCount", len(members)),
		)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 删除用户组及成员缓存
	if err := s.groupRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		logger.Error("删除用户组缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("operatorID", operatorID),
			zap.Error(err),
		)
	}
	if err := s.groupMemberRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		logger.Error("删除用户组成员缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("operatorID", operatorID),
			zap.Error(err),
		)
	}
	for _, member := range members {
		if err := s.groupMemberRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, member.UserID); err != nil {
			logger.Error("删除用户组成员缓存失败：Redis操作错误",
				zap.Int("groupID", groupID),
				zap.Int("userID", member.UserID),
				zap.String("username", member.Username),
				zap.String("role", member.Role),
				zap.Int("operatorID", operatorID),
				zap.Error(err),
			)
		}
		if err := s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, member.UserID); err != nil {
			logger.Error("删除用户申请缓存失败：Redis操作错误",
				zap.Int("groupID", groupID),
				zap.Int("userID", member.UserID),
				zap.String("username", member.Username),
				zap.String("role", member.Role),
				zap.Int("operatorID", operatorID),
				zap.Error(err),
			)
		}
	}
	return &updatedGroup, nil
}

// 检查用户组成员权限
func (s *GroupsService) CheckMemberPermission(ctx context.Context, groupID, userID int) error {
	// 从缓存查找
	existMember, err := s.groupMemberRedisDAO.GetMemberByGroupIDAndUserID(ctx, groupID, userID)
	if err == nil && existMember != nil {
		if existMember.Role == "admin" {
			return nil
		}
		return appErrors.ErrRolePermissionDenied
	}

	// 从数据库查找
	member, err := s.groupMemberDao.GetMemberByGroupIDAndUserID(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("检查成员权限失败：成员不存在",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
			)
			return appErrors.ErrGroupMemberNotFound
		}
		logger.Error("检查成员权限失败：数据库操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Error(err),
		)
		return appErrors.ErrDatabaseOperation.WithError(err)
	}

	// 缓存用户组成员
	if err := s.groupMemberRedisDAO.SetMemberByGroupIDAndUserID(ctx, member); err != nil {
		logger.Error("缓存用户组成员失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Error(err),
		)
	}

	if member.Role == "admin" {
		return nil
	}
	logger.Error("检查成员权限失败：权限不足",
		zap.Int("groupID", groupID),
		zap.Int("userID", userID),
		zap.String("role", member.Role),
	)
	return appErrors.ErrRolePermissionDenied
}

// 检查用户是否存在于用户组
func (s *GroupsService) CheckUserExistInGroup(ctx context.Context, groupID, userID int) error {
	// 从缓存查找
	existMember, err := s.groupMemberRedisDAO.GetMemberByGroupIDAndUserID(ctx, groupID, userID)
	if err == nil && existMember != nil {
		return nil
	}
	// 从数据库查找
	member, err := s.groupMemberDao.GetMemberByGroupIDAndUserID(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("检查用户组成员失败：成员不存在",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
			)
			return appErrors.ErrGroupMemberNotFound
		}
		logger.Error("检查用户组成员失败：数据库操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Error(err),
		)
		return appErrors.ErrDatabaseOperation.WithError(err)
	}

	// 缓存用户组成员
	if err := s.groupMemberRedisDAO.SetMemberByGroupIDAndUserID(ctx, member); err != nil {
		logger.Error("缓存用户组成员失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Error(err),
		)
	}
	return nil
}

// 删除用户组中的用户
func (s *GroupsService) RemoveMemberFromGroup(ctx context.Context, groupID, userID, operatorID int) error {
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查操作员权限
		if err := s.CheckMemberPermission(ctx, groupID, operatorID); err != nil {
			logger.Error("删除用户组成员失败：权限不足",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Int("operatorID", operatorID),
				zap.Error(err),
			)
			return appErrors.ErrRolePermissionDenied.WithError(err)
		}
		//删除用户组成员
		if err := s.groupMemberDao.Delete(ctx, groupID, userID, tx); err != nil {
			logger.Error("删除用户组成员失败：删除成员记录失败",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Int("operatorID", operatorID),
				zap.Error(err),
			)
			return appErrors.ErrGroupMemberDeletionFailed.WithError(err)
		}
		//更新用户组成员数量
		if err := s.groupDao.UpdateMemberNum(ctx, groupID, false, tx); err != nil {
			logger.Error("删除用户组成员失败：更新成员数量失败",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Int("operatorID", operatorID),
				zap.Error(err),
			)
			return appErrors.ErrGroupUpdateFailed.WithError(err)
		}
		logger.Info("成功删除用户组成员",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("operatorID", operatorID),
		)
		return nil
	})
	if err != nil {
		return err
	}

	// 删除用户组及成员缓存
	if err := s.groupRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		logger.Error("删除用户组缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("operatorID", operatorID),
			zap.Error(err),
		)
	}
	if err := s.groupMemberRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		logger.Error("删除用户组成员缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("operatorID", operatorID),
			zap.Error(err),
		)
	}
	if err := s.groupMemberRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, userID); err != nil {
		logger.Error("删除用户组成员缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("operatorID", operatorID),
			zap.Error(err),
		)
	}

	// 删除用户组任务缓存
	if err := s.taskRedisDAO.DeleteCacheByUserID(ctx, userID); err != nil {
		logger.Error("删除用户组任务缓存失败：Redis操作错误",
			zap.Int("userID", userID),
			zap.Error(err),
		)
	}
	return nil
}

// 查询用户组中的所有成员
func (s *GroupsService) GetMembersByGroupID(ctx context.Context, groupID int) ([]*models.GroupMember, error) {
	// 从缓存查找
	existMembers, err := s.groupMemberRedisDAO.GetMembersByGroupID(ctx, groupID)
	if err == nil && existMembers != nil && len(existMembers) > 0 {
		return existMembers, nil
	}

	// 从数据库查找
	var members []*models.GroupMember
	err = s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		groupMembers, err := s.groupMemberDao.GetMembersByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取用户组成员列表失败：用户组不存在",
					zap.Int("groupID", groupID),
					zap.Error(err),
				)
				return appErrors.ErrGroupNotFound
			}
			logger.Error("获取用户组成员列表失败：数据库操作错误",
				zap.Int("groupID", groupID),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		members = groupMembers
		logger.Info("成功获取用户组成员列表",
			zap.Int("groupID", groupID),
			zap.Int("memberCount", len(members)),
			zap.Any("members", members),
		)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 缓存用户组成员
	if err := s.groupMemberRedisDAO.SetMembersByGroupID(ctx, groupID, members); err != nil {
		logger.Error("缓存用户组成员列表失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("memberCount", len(members)),
			zap.Error(err),
		)
	}
	return members, nil
}

// 创建用户申请加入记录
func (s *GroupsService) CreateJoinApplication(ctx context.Context, groupID, userID int, username, reason string) (*models.JoinApplication, error) {
	// 从缓存查找是否已存在申请记录，避免重复申请
	existApplication, err := s.joinApplicationRedisDAO.GetByGroupIDAndUserID(ctx, groupID, userID)
	if err == nil && existApplication != nil && existApplication.Status == "pending" {
		return nil, appErrors.ErrJoinApplicationAlreadyExists
	}

	// 数据库层面进行查询并防止重复申请
	var application models.JoinApplication
	err = s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查用户组是否存在
		_, err := s.groupDao.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("创建加入申请失败：用户组不存在",
					zap.Int("groupID", groupID),
					zap.Int("userID", userID),
					zap.String("username", username),
					zap.String("reason", reason),
					zap.Error(err),
				)
				return appErrors.ErrGroupNotFound
			}
			logger.Error("创建加入申请失败：数据库操作错误",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.String("username", username),
				zap.String("reason", reason),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		//检查用户是否已加入用户组
		member, err := s.groupMemberDao.GetMemberByGroupIDAndUserID(ctx, groupID, userID, tx)
		if err == nil && member != nil {
			logger.Error("创建加入申请失败：用户已是成员",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.String("username", username),
				zap.String("role", member.Role),
				zap.Error(err),
			)
			return appErrors.ErrGroupMemberAlreadyExists
		}
		//检查用户是否已存在申请记录
		existApplication, err := s.joinApplicationDao.GetByGroupIDAndUserID(ctx, groupID, userID, tx)
		if err == nil && existApplication != nil && existApplication.Status == "pending" {
			logger.Error("创建加入申请失败：已存在待处理的申请",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.String("username", username),
				zap.Int("requestID", existApplication.RequestID),
				zap.String("status", existApplication.Status),
				zap.Error(err),
			)
			return appErrors.ErrJoinApplicationAlreadyExists
		}
		//创建申请记录
		newApplication := models.JoinApplication{
			GroupID:   groupID,
			UserID:    userID,
			Username:  username,
			Reason:    reason,
			CreatedAt: time.Now(),
		}
		if err := s.joinApplicationDao.Create(ctx, &newApplication, tx); err != nil {
			logger.Error("创建加入申请失败：创建申请记录失败",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.String("username", username),
				zap.String("reason", reason),
				zap.Error(err),
			)
			return appErrors.ErrJoinApplicationCreationFailed.WithError(err)
		}
		application = newApplication
		logger.Info("成功创建加入申请",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.String("username", username),
			zap.String("reason", reason),
			zap.Int("requestID", application.RequestID),
			zap.String("status", application.Status),
		)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 缓存单个用户申请记录
	if err := s.joinApplicationRedisDAO.SetByGroupIDAndUserID(ctx, &application); err != nil {
		logger.Error("缓存用户申请记录失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.String("username", username),
			zap.Int("requestID", application.RequestID),
			zap.String("status", application.Status),
			zap.Error(err),
		)
	}

	// 缓存用户组所有申请记录
	existGroupApplication, err := s.joinApplicationRedisDAO.GetByGroupID(ctx, groupID)
	if err != nil {
		logger.Error("缓存用户组申请记录失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Error(err),
		)
	}
	if existGroupApplication == nil {
		existGroupApplication = []*models.JoinApplication{}
	}
	existGroupApplication = append(existGroupApplication, &application)
	if err := s.joinApplicationRedisDAO.SetByGroupID(ctx, groupID, existGroupApplication); err != nil {
		logger.Error("缓存用户组申请记录失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.String("username", username),
			zap.Int("requestID", application.RequestID),
			zap.String("status", application.Status),
			zap.Error(err),
		)
	}

	// 缓存用户组待审批申请记录
	existGroupPendingApplication, err := s.joinApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, "pending")
	if err != nil {
		logger.Error("缓存用户组待审批申请记录失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Error(err),
		)
	}
	if existGroupPendingApplication == nil {
		existGroupPendingApplication = []*models.JoinApplication{}
	}
	existGroupPendingApplication = append(existGroupPendingApplication, &application)
	if err := s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "pending", existGroupPendingApplication); err != nil {
		logger.Error("缓存用户组待审批申请记录失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.String("username", username),
			zap.Int("requestID", application.RequestID),
			zap.String("status", application.Status),
			zap.Error(err),
		)
	}
	return &application, nil
}

// 查看用户组加入申请列表
func (s *GroupsService) GetJoinApplicationsByGroupID(ctx context.Context, groupID, operatorID int, filter ...string) ([]*models.JoinApplication, error) {
	// 从缓存查找,按照有无filter进行不同查询
	if len(filter) == 0 {
		existGroupApplication, err := s.joinApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, "pending")
		if err == nil && existGroupApplication != nil {
			return existGroupApplication, nil
		}
	} else {
		var joinApplicaiton []*models.JoinApplication
		var err error

		switch filter[0] {
		case "pending":
			joinApplicaiton, err = s.joinApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, "pending")
		case "approved":
			joinApplicaiton, err = s.joinApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, "accepted")
		case "rejected":
			joinApplicaiton, err = s.joinApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, "rejected")
		default:
			joinApplicaiton, err = s.joinApplicationRedisDAO.GetByGroupID(ctx, groupID)
		}

		if err == nil && joinApplicaiton != nil {
			return joinApplicaiton, nil
		}
	}

	// 从数据库查找
	var applications []*models.JoinApplication
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查操作员权限
		if err := s.CheckMemberPermission(ctx, groupID, operatorID); err != nil {
			logger.Error("获取加入申请列表失败：权限不足",
				zap.Int("groupID", groupID),
				zap.Int("operatorID", operatorID),
				zap.Strings("filter", filter),
				zap.Error(err),
			)
			return appErrors.ErrRolePermissionDenied.WithError(err)
		}
		//检查用户组是否存在
		if _, err := s.groupDao.GetByGroupID(ctx, groupID, tx); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取加入申请列表失败：用户组不存在",
					zap.Int("groupID", groupID),
					zap.Strings("filter", filter),
				)
				return appErrors.ErrGroupNotFound
			}
			logger.Error("获取加入申请列表失败：数据库操作错误",
				zap.Int("groupID", groupID),
				zap.Strings("filter", filter),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}

		var existApplications []*models.JoinApplication
		var err error

		//按照filter查询申请记录
		if len(filter) == 0 { //默认
			existApplications, err = s.joinApplicationDao.GetByGroupIDAndStatus(ctx, groupID, "pending", tx)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					logger.Error("获取加入申请列表失败：未找到申请记录",
						zap.Int("groupID", groupID),
						zap.String("status", "pending"),
					)
					return appErrors.ErrJoinApplicationNotFound
				}
				logger.Error("获取加入申请列表失败：数据库操作错误",
					zap.Int("groupID", groupID),
					zap.Error(err),
				)
				return appErrors.ErrDatabaseOperation.WithError(err)
			}
		} else {
			switch filter[0] {
			case "pending":
				existApplications, err = s.joinApplicationDao.GetByGroupIDAndStatus(ctx, groupID, "pending", tx)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						logger.Error("获取加入申请列表失败：未找到待处理申请",
							zap.Int("groupID", groupID),
							zap.String("status", "pending"),
						)
						return appErrors.ErrJoinApplicationNotFound
					}
					logger.Error("获取加入申请列表失败：数据库操作错误",
						zap.Int("groupID", groupID),
						zap.Error(err),
					)
					return appErrors.ErrDatabaseOperation.WithError(err)
				}
			case "approved":
				existApplications, err = s.joinApplicationDao.GetByGroupIDAndStatus(ctx, groupID, "accepted", tx)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						logger.Error("获取加入申请列表失败：未找到已通过申请",
							zap.Int("groupID", groupID),
							zap.String("status", "accepted"),
						)
						return appErrors.ErrJoinApplicationNotFound
					}
					logger.Error("获取加入申请列表失败：数据库操作错误",
						zap.Int("groupID", groupID),
						zap.Error(err),
					)
					return appErrors.ErrDatabaseOperation.WithError(err)
				}
			case "rejected":
				existApplications, err = s.joinApplicationDao.GetByGroupIDAndStatus(ctx, groupID, "rejected", tx)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						logger.Error("获取加入申请列表失败：未找到已拒绝申请",
							zap.Int("groupID", groupID),
							zap.String("status", "rejected"),
						)
						return appErrors.ErrJoinApplicationNotFound
					}
					logger.Error("获取加入申请列表失败：数据库操作错误",
						zap.Int("groupID", groupID),
						zap.Error(err),
					)
					return appErrors.ErrDatabaseOperation.WithError(err)
				}
			default:
				existApplications, err = s.joinApplicationDao.GetByGroupID(ctx, groupID, tx)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						logger.Error("获取加入申请列表失败：未找到申请记录",
							zap.Int("groupID", groupID),
						)
						return appErrors.ErrJoinApplicationNotFound
					}
					logger.Error("获取加入申请列表失败：数据库操作错误",
						zap.Int("groupID", groupID),
						zap.Error(err),
					)
					return appErrors.ErrDatabaseOperation.WithError(err)
				}
			}
		}
		applications = existApplications
		logger.Info("成功获取加入申请列表",
			zap.Int("groupID", groupID),
			zap.Int("applicationCount", len(applications)),
			zap.Strings("filter", filter),
			zap.Int("operatorID", operatorID),
		)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 根据对应filter缓存记录
	if len(filter) == 0 {
		if err := s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "pending", applications); err != nil {
			logger.Error("缓存用户组待审批申请记录失败：Redis操作错误",
				zap.Int("groupID", groupID),
				zap.Error(err),
			)
		}
	} else {
		var err error
		switch filter[0] {
		case "pending":
			err = s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "pending", applications)
		case "approved":
			err = s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "accepted", applications)
		case "rejected":
			err = s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "rejected", applications)
		default:
			err = s.joinApplicationRedisDAO.SetByGroupID(ctx, groupID, applications)
		}
		if err != nil {
			logger.Error("缓存用户组申请记录失败：Redis操作错误",
				zap.Int("groupID", groupID),
				zap.Strings("filter", filter),
				zap.Error(err),
			)
		}
	}

	return applications, nil
}

// 审批用户组加入申请（通过申请）
func (s *GroupsService) ApproveJoinApplication(ctx context.Context, groupID, userID, operatorID, requestID int, username string) error {
	// 用于缓存审批通过的记录
	var application models.JoinApplication

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查操作员权限
		if err := s.CheckMemberPermission(ctx, groupID, operatorID); err != nil {
			logger.Error("审批加入申请失败：权限不足",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Int("operatorID", operatorID),
				zap.Int("requestID", requestID),
				zap.String("username", username),
				zap.Error(err),
			)
			return appErrors.ErrRolePermissionDenied.WithError(err)
		}
		//更新用户组成员数量
		if err := s.groupDao.UpdateMemberNum(ctx, groupID, true, tx); err != nil {
			logger.Error("审批加入申请失败：更新成员数量失败",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Int("requestID", requestID),
				zap.String("username", username),
				zap.Error(err),
			)
			return appErrors.ErrGroupUpdateFailed.WithError(err)
		}
		//添加用户组成员
		if err := s.groupMemberDao.Create(ctx, &models.GroupMember{
			GroupID:   groupID,
			UserID:    userID,
			Username:  username,
			CreatedAt: time.Now(),
		}, tx); err != nil {
			logger.Error("审批加入申请失败：添加成员失败",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.String("username", username),
				zap.Int("requestID", requestID),
				zap.Error(err),
			)
			return appErrors.ErrGroupMemberCreationFailed.WithError(err)
		}
		//更新申请记录状态
		if err := s.joinApplicationDao.UpdateStatus(ctx, requestID, "accepted", tx); err != nil {
			logger.Error("审批加入申请失败：更新申请状态失败",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Int("requestID", requestID),
				zap.String("username", username),
				zap.Error(err),
			)
			return appErrors.ErrJoinApplicationUpdateFailed.WithError(err)
		}

		// 接受审批通过的记录，用于缓存
		newApplication, err := s.joinApplicationDao.GetByRequestID(ctx, requestID, tx)
		if err != nil {
			logger.Error("审批加入申请失败：获取更新后的申请记录失败",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Int("requestID", requestID),
				zap.String("username", username),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		application = *newApplication
		logger.Info("成功审批通过加入申请",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.String("username", username),
			zap.Int("requestID", requestID),
			zap.Int("operatorID", operatorID),
			zap.String("status", "accepted"),
		)
		return nil
	})
	if err != nil {
		return err
	}

	// 删除用户组及成员缓存(单个用户记录没有变化，不需要删除)
	if err := s.groupRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		logger.Error("删除用户组缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("requestID", requestID),
			zap.Error(err),
		)
	}
	if err := s.groupMemberRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		logger.Error("删除用户组成员缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("requestID", requestID),
			zap.Error(err),
		)
	}

	// 删除用户申请缓存
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, userID); err != nil {
		logger.Error("删除用户申请缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("requestID", requestID),
			zap.Error(err),
		)
	}

	// 删除用户组所有申请缓存（所有状态）
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		logger.Error("删除用户组申请缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("requestID", requestID),
			zap.Error(err),
		)
	}

	// 更新用户组待审批状态缓存
	existGroupPendingApplication, err := s.joinApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, "pending")
	if err != nil {
		logger.Error("更新用户组待审批状态缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Error(err),
		)
	}
	if existGroupPendingApplication != nil && len(existGroupPendingApplication) > 0 {
		for i := 0; i < len(existGroupPendingApplication); i++ {
			if existGroupPendingApplication[i].RequestID == requestID {
				existGroupPendingApplication = append(existGroupPendingApplication[:i], existGroupPendingApplication[i+1:]...)
				break
			}
		}
	}
	if err := s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "pending", existGroupPendingApplication); err != nil {
		logger.Error("更新用户组待审批状态缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("requestID", requestID),
			zap.Error(err),
		)
	}

	// 更新用户组加入申请审批通过缓存
	existGroupApprovedApplication, err := s.joinApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, "accepted")
	if err != nil {
		logger.Error("更新用户组加入申请审批通过缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Error(err),
		)
	}
	if existGroupApprovedApplication == nil {
		existGroupApprovedApplication = []*models.JoinApplication{}
	}
	existGroupApprovedApplication = append(existGroupApprovedApplication, &application)
	if err := s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "accepted", existGroupApprovedApplication); err != nil {
		logger.Error("更新用户组审批通过缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("requestID", requestID),
			zap.Error(err),
		)
	}

	// 删除用户加入用户组的申请缓存
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, userID); err != nil {
		logger.Error("删除用户申请缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("requestID", requestID),
			zap.Error(err),
		)
	}

	return nil
}

// 拒绝用户组加入申请
func (s *GroupsService) RejectJoinApplication(ctx context.Context, groupID, userID, operatorID, requestID int, username, rejectReason string) error {
	// 用于缓存审批拒绝的记录
	var application models.JoinApplication

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查操作员权限
		if err := s.CheckMemberPermission(ctx, groupID, operatorID); err != nil {
			logger.Error("拒绝加入申请失败：权限不足",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Int("operatorID", operatorID),
				zap.Int("requestID", requestID),
				zap.String("username", username),
				zap.String("rejectReason", rejectReason),
				zap.Error(err),
			)
			return appErrors.ErrRolePermissionDenied.WithError(err)
		}
		//更新申请记录状态
		if err := s.joinApplicationDao.UpdateStatus(ctx, requestID, "rejected", tx); err != nil {
			logger.Error("拒绝加入申请失败：更新申请状态失败",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Int("requestID", requestID),
				zap.String("username", username),
				zap.String("rejectReason", rejectReason),
				zap.Error(err),
			)
			return appErrors.ErrJoinApplicationUpdateFailed.WithError(err)
		}
		//更新拒绝理由
		if err := s.joinApplicationDao.UpdateRejectReason(ctx, requestID, rejectReason, tx); err != nil {
			logger.Error("拒绝加入申请失败：更新拒绝理由失败",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Int("requestID", requestID),
				zap.String("username", username),
				zap.String("rejectReason", rejectReason),
				zap.Error(err),
			)
			return appErrors.ErrJoinApplicationUpdateFailed.WithError(err)
		}

		// 接受审批拒绝的记录，用于缓存
		newApplication, err := s.joinApplicationDao.GetByRequestID(ctx, requestID, tx)
		if err != nil {
			logger.Error("拒绝加入申请失败：获取更新后的申请记录失败",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Int("requestID", requestID),
				zap.String("username", username),
				zap.String("rejectReason", rejectReason),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		application = *newApplication
		logger.Info("成功拒绝加入申请",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.String("username", username),
			zap.Int("requestID", requestID),
			zap.String("rejectReason", rejectReason),
			zap.Int("operatorID", operatorID),
			zap.String("status", "rejected"),
		)
		return nil
	})
	if err != nil {
		return err
	}

	// 删除申请缓存

	// 删除用户申请缓存
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, userID); err != nil {
		logger.Error("删除用户申请缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("requestID", requestID),
			zap.String("username", username),
			zap.Error(err),
		)
	}

	// 删除用户组所有申请缓存（所有状态）
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		logger.Error("删除用户组申请缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("requestID", requestID),
			zap.String("username", username),
			zap.Error(err),
		)
	}

	// 更新用户组待审批状态缓存
	existGroupPendingApplication, err := s.joinApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, "pending")
	if err != nil {
		logger.Error("更新用户组待审批状态缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Error(err),
		)
	}
	if existGroupPendingApplication != nil && len(existGroupPendingApplication) > 0 {
		for i := 0; i < len(existGroupPendingApplication); i++ {
			if existGroupPendingApplication[i].RequestID == requestID {
				existGroupPendingApplication = append(existGroupPendingApplication[:i], existGroupPendingApplication[i+1:]...)
				break
			}
		}
	}
	if err := s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "pending", existGroupPendingApplication); err != nil {
		logger.Error("更新用户组待审批状态缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("requestID", requestID),
			zap.String("username", username),
			zap.Error(err),
		)
	}

	// 更新用户组加入申请审批拒绝缓存
	existGroupRejectedApplication, err := s.joinApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, "rejected")
	if err != nil {
		logger.Error("更新用户组加入申请审批拒绝缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Error(err),
		)
	}
	if existGroupRejectedApplication == nil {
		existGroupRejectedApplication = []*models.JoinApplication{}
	}
	existGroupRejectedApplication = append(existGroupRejectedApplication, &application)
	if err := s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "rejected", existGroupRejectedApplication); err != nil {
		logger.Error("更新用户组审批拒绝缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("requestID", requestID),
			zap.String("username", username),
			zap.Error(err),
		)
	}

	// 删除用户加入用户组的申请缓存
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, userID); err != nil {
		logger.Error("删除用户申请缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("userID", userID),
			zap.Int("requestID", requestID),
			zap.String("username", username),
			zap.Error(err),
		)
	}

	return nil
}

// 删除用户组
func (s *GroupsService) DeleteGroup(ctx context.Context, groupID, operatorID int) error {
	var members []*models.GroupMember
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查操作员权限
		if err := s.CheckMemberPermission(ctx, groupID, operatorID); err != nil {
			logger.Error("删除用户组失败：权限不足",
				zap.Int("groupID", groupID),
				zap.Int("operatorID", operatorID),
				zap.Error(err),
			)
			return appErrors.ErrRolePermissionDenied.WithError(err)
		}
		//删除用户组
		if err := s.groupDao.Delete(ctx, groupID, tx); err != nil {
			logger.Error("删除用户组失败：删除用户组记录失败",
				zap.Int("groupID", groupID),
				zap.Int("operatorID", operatorID),
				zap.Error(err),
			)
			return appErrors.ErrGroupDeletionFailed.WithError(err)
		}
		//删除用户组成员
		groupMembers, err := s.groupMemberDao.GetMembersByGroupID(ctx, groupID, tx)
		if err != nil {
			logger.Error("删除用户组失败：获取成员列表失败",
				zap.Int("groupID", groupID),
				zap.Int("operatorID", operatorID),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		members = groupMembers
		for _, member := range groupMembers {
			if err := s.groupMemberDao.Delete(ctx, groupID, member.UserID, tx); err != nil {
				logger.Error("删除用户组失败：删除成员记录失败",
					zap.Int("groupID", groupID),
					zap.Int("userID", member.UserID),
					zap.String("username", member.Username),
					zap.String("role", member.Role),
					zap.Int("operatorID", operatorID),
					zap.Error(err),
				)
				return appErrors.ErrGroupMemberDeletionFailed.WithError(err)
			}
		}
		logger.Info("成功删除用户组",
			zap.Int("groupID", groupID),
			zap.Int("operatorID", operatorID),
			zap.Int("memberCount", len(groupMembers)),
			zap.Any("members", groupMembers),
		)
		return nil
	})
	if err != nil {
		return err
	}

	// 删除用户组及 成员缓存（组内记录+申请记录）
	if err := s.groupRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		logger.Error("删除用户组缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("operatorID", operatorID),
			zap.Error(err),
		)
	}
	if err := s.groupMemberRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		logger.Error("删除用户组成员缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("operatorID", operatorID),
			zap.Error(err),
		)
	}
	for _, member := range members {
		if err := s.groupMemberRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, member.UserID); err != nil {
			logger.Error("删除用户组成员缓存失败：Redis操作错误",
				zap.Int("groupID", groupID),
				zap.Int("userID", member.UserID),
				zap.String("username", member.Username),
				zap.String("role", member.Role),
				zap.Int("operatorID", operatorID),
				zap.Error(err),
			)
		}
		if err := s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, member.UserID); err != nil {
			logger.Error("删除用户申请缓存失败：Redis操作错误",
				zap.Int("groupID", groupID),
				zap.Int("userID", member.UserID),
				zap.String("username", member.Username),
				zap.String("role", member.Role),
				zap.Int("operatorID", operatorID),
				zap.Error(err),
			)
		}
	}

	// 删除用户组相关所有申请缓存
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		logger.Error("删除用户组申请缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("operatorID", operatorID),
			zap.Error(err),
		)
	}
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndStatus(ctx, groupID, "pending"); err != nil {
		logger.Error("删除用户组待审批申请缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("operatorID", operatorID),
			zap.Error(err),
		)
	}
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndStatus(ctx, groupID, "accepted"); err != nil {
		logger.Error("删除用户组已通过申请缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("operatorID", operatorID),
			zap.Error(err),
		)
	}
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndStatus(ctx, groupID, "rejected"); err != nil {
		logger.Error("删除用户组已拒绝申请缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("operatorID", operatorID),
			zap.Error(err),
		)
	}

	// 删除用户组任务缓存
	if err := s.taskRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		logger.Error("删除用户组任务缓存失败：Redis操作错误",
			zap.Int("groupID", groupID),
			zap.Int("operatorID", operatorID),
			zap.Error(err),
		)
	}

	// 删除用户组成员任务缓存
	for _, member := range members {
		if err:=s.taskRedisDAO.DeleteCacheByUserID(ctx,member.UserID);err!=nil{
			logger.Error("删除用户组成员任务缓存失败：Redis操作错误",
				zap.Int("groupID", groupID),
				zap.Int("userID", member.UserID),
				zap.String("username", member.Username),
				zap.Error(err),
			)
		}
	}
	return nil
}

// 查询当前登录用户在指定用户组中的状态
func (s *GroupsService) GetUserGroupStatus(ctx context.Context, groupID, userID int) (string, int, error) {
	var status string
	var requestID int
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查用户组是否存在
		_, err := s.groupDao.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取用户组状态失败：用户组不存在",
					zap.Int("groupID", groupID),
					zap.Int("userID", userID),
					zap.Error(err),
				)
				return appErrors.ErrGroupNotFound
			}
			logger.Error("获取用户组状态失败：数据库操作错误",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		// 检查是否为组成员
		member, err := s.groupMemberDao.GetMemberByGroupIDAndUserID(ctx, groupID, userID, tx)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取用户组状态失败：数据库操作错误",
					zap.Int("groupID", groupID),
					zap.Int("userID", userID),
					zap.Error(err),
				)
				return appErrors.ErrDatabaseOperation.WithError(err)
			}
		}
		if member != nil {
			status = member.Role
			requestID = 0
			logger.Info("成功获取用户组状态：用户是组成员",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.String("role", status),
				zap.String("username", member.Username),
			)
			return nil
		}
		// 非组成员，查看申请记录
		Application, err := s.joinApplicationDao.GetByGroupIDAndUserID(ctx, groupID, userID, tx)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取用户组状态失败：数据库操作错误",
					zap.Int("groupID", groupID),
					zap.Int("userID", userID),
					zap.Error(err),
				)
				return appErrors.ErrDatabaseOperation.WithError(err)
			}
		}
		if Application == nil {
			status = "none"
			requestID = 0
			logger.Info("成功获取用户组状态：用户未关联",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
			)
		} else {
			status = Application.Status
			requestID = Application.RequestID
			logger.Info("成功获取用户组状态：用户有申请记录",
				zap.Int("groupID", groupID),
				zap.Int("userID", userID),
				zap.String("status", status),
				zap.Int("requestID", requestID),
				zap.String("username", Application.Username),
				zap.String("reason", Application.Reason),
			)
		}
		return nil
	})
	if err != nil {
		return "", 0, err
	}
	return status, requestID, nil
}
