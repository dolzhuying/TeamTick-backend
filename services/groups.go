package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	appErrors "TeamTickBackend/pkg/errors"
	"context"
	"errors"
	"log"

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
}

func NewGroupsService(
	groupDao dao.GroupDAO,
	groupMemberDao dao.GroupMemberDAO,
	joinApplicationDao dao.JoinApplicationDAO,
	transactionManager dao.TransactionManager,
	groupRedisDAO dao.GroupRedisDAO,
	groupMemberRedisDAO dao.GroupMemberRedisDAO,
	joinApplicationRedisDAO dao.JoinApplicationRedisDAO,
) *GroupsService {

	return &GroupsService{
		groupDao:                groupDao,
		groupMemberDao:          groupMemberDao,
		joinApplicationDao:      joinApplicationDao,
		transactionManager:      transactionManager,
		groupRedisDAO:           groupRedisDAO,
		groupMemberRedisDAO:     groupMemberRedisDAO,
		joinApplicationRedisDAO: joinApplicationRedisDAO,
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
		}
		//创建用户组
		if err := s.groupDao.Create(ctx, &group, tx); err != nil {
			return appErrors.ErrGroupCreationFailed.WithError(err)
		}
		//添加用户组管理员
		if err := s.groupMemberDao.Create(ctx, &models.GroupMember{
			GroupID:  group.GroupID,
			UserID:   creatorID,
			Username: creatorName,
			Role:     "admin",
		}, tx); err != nil {
			return appErrors.ErrGroupMemberCreationFailed.WithError(err)
		}
		createdGroup = group
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 缓存用户组
	if err := s.groupRedisDAO.SetByGroupID(ctx, createdGroup.GroupID, &createdGroup); err != nil {
		log.Printf("set group to redis failed, err: %v", err)
	}
	return &createdGroup, nil
}

// 通过GroupID查询用户组信息
func (s *GroupsService) GetGroupByGroupID(ctx context.Context, groupID int) (*models.Group, error) {
	// 从缓存中查找用户组
	existGroup, err := s.groupRedisDAO.GetByGroupID(ctx, groupID)
	if err == nil && existGroup != nil {
		return existGroup, nil
	}

	// 从数据库中查找用户组
	var group models.Group
	err = s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		Group, err := s.groupDao.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrGroupNotFound
			}
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		group = *Group
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 缓存用户组
	if err := s.groupRedisDAO.SetByGroupID(ctx, group.GroupID, &group); err != nil {
		log.Printf("set group to redis failed, err: %v", err)
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
				return appErrors.ErrGroupNotFound
			}
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		groups = Groups
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
			return appErrors.ErrRolePermissionDenied.WithError(err)
		}
		//更新用户组信息
		if err := s.groupDao.UpdateMessage(ctx, groupID, groupName, description, tx); err != nil {
			return appErrors.ErrGroupUpdateFailed.WithError(err)
		}
		//查询更新后的用户组信息
		group, err := s.groupDao.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		// 查询用户组成员列表以删除缓存
		members, err = s.groupMemberDao.GetMembersByGroupID(ctx, groupID, tx)
		if err != nil {
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		updatedGroup = *group
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 删除用户组及成员缓存
	if err := s.groupRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		log.Printf("delete group cache failed, err: %v", err)
	}
	if err := s.groupMemberRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		log.Printf("delete group members cache failed, err: %v", err)
	}
	for _, member := range members {
		if err := s.groupMemberRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, member.UserID); err != nil {
			log.Printf("delete group member cache failed, err: %v", err)
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
			return appErrors.ErrGroupMemberNotFound
		}
		return appErrors.ErrDatabaseOperation.WithError(err)
	}

	// 缓存用户组成员
	if err := s.groupMemberRedisDAO.SetMemberByGroupIDAndUserID(ctx, member); err != nil {
		log.Printf("set group member to redis failed, err: %v", err)
	}

	if member.Role == "admin" {
		return nil
	}
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
			return appErrors.ErrGroupMemberNotFound
		}
		return appErrors.ErrDatabaseOperation.WithError(err)
	}

	// 缓存用户组成员
	if err := s.groupMemberRedisDAO.SetMemberByGroupIDAndUserID(ctx, member); err != nil {
		log.Printf("set group member to redis failed, err: %v", err)
	}
	return nil
}

// 删除用户组中的用户
func (s *GroupsService) RemoveMemberFromGroup(ctx context.Context, groupID, userID, operatorID int) error {
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查操作员权限
		if err := s.CheckMemberPermission(ctx, groupID, operatorID); err != nil {
			return appErrors.ErrRolePermissionDenied.WithError(err)
		}
		//删除用户组成员
		if err := s.groupMemberDao.Delete(ctx, groupID, userID, tx); err != nil {
			return appErrors.ErrGroupMemberDeletionFailed.WithError(err)
		}
		//更新用户组成员数量
		if err := s.groupDao.UpdateMemberNum(ctx, groupID, false, tx); err != nil {
			return appErrors.ErrGroupUpdateFailed.WithError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 删除用户组及成员缓存
	if err := s.groupRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		log.Printf("delete group cache failed, err: %v", err)
	}
	if err := s.groupMemberRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		log.Printf("delete group members cache failed, err: %v", err)
	}
	if err := s.groupMemberRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, userID); err != nil {
		log.Printf("delete group member cache failed, err: %v", err)
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
				return appErrors.ErrGroupNotFound
			}
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		members = groupMembers
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 缓存用户组成员
	if err := s.groupMemberRedisDAO.SetMembersByGroupID(ctx, groupID, members); err != nil {
		log.Printf("set group members to redis failed, err: %v", err)
	}
	return members, nil
}

// 创建用户申请加入记录(返回值？是否需要返回申请记录)
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
				return appErrors.ErrGroupNotFound
			}
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		//检查用户是否已加入用户组
		member, err := s.groupMemberDao.GetMemberByGroupIDAndUserID(ctx, groupID, userID, tx)
		if err == nil && member != nil {
			return appErrors.ErrGroupMemberAlreadyExists
		}
		//检查用户是否已存在申请记录
		existApplication, err := s.joinApplicationDao.GetByGroupIDAndUserID(ctx, groupID, userID, tx)
		if err == nil && existApplication != nil && existApplication.Status == "pending" {
			return appErrors.ErrJoinApplicationAlreadyExists
		}
		//创建申请记录
		newApplication := models.JoinApplication{
			GroupID:  groupID,
			UserID:   userID,
			Username: username,
			Reason:   reason,
		}
		if err := s.joinApplicationDao.Create(ctx, &newApplication, tx); err != nil {
			return appErrors.ErrJoinApplicationCreationFailed.WithError(err)
		}
		application = newApplication
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 缓存申请记录

	// 缓存单个用户申请记录
	if err := s.joinApplicationRedisDAO.SetByGroupIDAndUserID(ctx, &application); err != nil {
		log.Printf("set join application by userid and groupid to redis failed, err: %v", err)
	}

	// 缓存用户组所有申请记录
	existGroupApplication, err := s.joinApplicationRedisDAO.GetByGroupID(ctx, groupID)
	if existGroupApplication == nil {
		existGroupApplication = []*models.JoinApplication{}
	}
	existGroupApplication = append(existGroupApplication, &application)
	if err := s.joinApplicationRedisDAO.SetByGroupID(ctx, groupID, existGroupApplication); err != nil {
		log.Printf("set join application by groupid to redis failed, err: %v", err)
	}

	// 缓存用户组待审批申请记录
	existGroupPendingApplication, err := s.joinApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, "pending")
	if existGroupPendingApplication == nil {
		existGroupPendingApplication = []*models.JoinApplication{}
	}
	existGroupPendingApplication = append(existGroupPendingApplication, &application)
	if err := s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "pending", existGroupPendingApplication); err != nil {
		log.Printf("set join application by groupid and status to redis failed, err: %v", err)
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

		switch filter[0]{
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
			return appErrors.ErrRolePermissionDenied.WithError(err)
		}
		//检查用户组是否存在
		if _, err := s.groupDao.GetByGroupID(ctx, groupID, tx); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrGroupNotFound
			}
			return appErrors.ErrDatabaseOperation.WithError(err)
		}

		var existApplications []*models.JoinApplication
		var err error

		//按照filter查询申请记录
		if len(filter) == 0 { //默认
			existApplications, err = s.joinApplicationDao.GetByGroupIDAndStatus(ctx, groupID, "pending", tx)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return appErrors.ErrJoinApplicationNotFound
				}
				return appErrors.ErrDatabaseOperation.WithError(err)
			}
		} else {
			switch filter[0] {
			case "pending":
				existApplications, err = s.joinApplicationDao.GetByGroupIDAndStatus(ctx, groupID, "pending", tx)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return appErrors.ErrJoinApplicationNotFound
					}
					return appErrors.ErrDatabaseOperation.WithError(err)
				}
			case "approved":
				existApplications, err = s.joinApplicationDao.GetByGroupIDAndStatus(ctx, groupID, "accepted", tx)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return appErrors.ErrJoinApplicationNotFound
					}
					return appErrors.ErrDatabaseOperation.WithError(err)
				}
			case "rejected":
				existApplications, err = s.joinApplicationDao.GetByGroupIDAndStatus(ctx, groupID, "rejected", tx)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return appErrors.ErrJoinApplicationNotFound
					}
					return appErrors.ErrDatabaseOperation.WithError(err)
				}
			default:
				existApplications, err = s.joinApplicationDao.GetByGroupID(ctx, groupID, tx)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return appErrors.ErrJoinApplicationNotFound
					}
					return appErrors.ErrDatabaseOperation.WithError(err)
				}
			}
		}
		applications = existApplications
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 根据对应filter缓存记录
	if len(filter) == 0 {
		if err := s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "pending", applications); err != nil {
			log.Printf("set join application by groupid and status to redis failed, err: %v", err)
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
			log.Printf("set join application by groupid (with status) to redis failed, err: %v", err)
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
			return appErrors.ErrRolePermissionDenied.WithError(err)
		}
		//更新用户组成员数量
		if err := s.groupDao.UpdateMemberNum(ctx, groupID, true, tx); err != nil {
			return appErrors.ErrGroupUpdateFailed.WithError(err)
		}
		//添加用户组成员
		if err := s.groupMemberDao.Create(ctx, &models.GroupMember{
			GroupID:  groupID,
			UserID:   userID,
			Username: username,
		}, tx); err != nil {
			return appErrors.ErrGroupMemberCreationFailed.WithError(err)
		}
		//更新申请记录状态
		if err := s.joinApplicationDao.UpdateStatus(ctx, requestID, "accepted", tx); err != nil {
			return appErrors.ErrJoinApplicationUpdateFailed.WithError(err)
		}
		
		// 接受审批通过的记录，用于缓存
		newApplication,err:=s.joinApplicationDao.GetByRequestID(ctx,requestID,tx)
		if err!=nil{
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		application = *newApplication
		return nil
	})
	if err != nil {
		return err
	}

	// 删除用户组及成员缓存(单个用户记录没有变化，不需要删除)
	if err := s.groupRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		log.Printf("delete group cache failed, err: %v", err)
	}
	if err := s.groupMemberRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		log.Printf("delete group members cache failed, err: %v", err)
	}

	// 删除申请缓存

	// 删除用户申请缓存
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, userID); err != nil {
		log.Printf("delete join application by userid and groupid to redis failed, err: %v", err)
	}

	// 删除用户组所有申请缓存（所有状态）
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		log.Printf("delete join application by groupid to redis failed, err: %v", err)
	}

	// 更新用户组待审批状态缓存
	existGroupPendingApplication, err := s.joinApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, "pending")
	if existGroupPendingApplication !=nil && len(existGroupPendingApplication)>0{
		for i:=0;i<len(existGroupPendingApplication);i++{
			if existGroupPendingApplication[i].RequestID == requestID {
				existGroupPendingApplication=append(existGroupPendingApplication[:i],existGroupPendingApplication[i+1:]...)
				break
			}
		}
	}
	if err := s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "pending", existGroupPendingApplication); err != nil {
		log.Printf("set join application by groupid and status to redis failed, err: %v", err)
	}

	// 更新用户组加入申请审批通过缓存
	existGroupApprovedApplication, err := s.joinApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, "accepted")
	if existGroupApprovedApplication ==nil{
		existGroupApprovedApplication = []*models.JoinApplication{}
	}
	existGroupApprovedApplication = append(existGroupApprovedApplication, &application)
	if err := s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "accepted", existGroupApprovedApplication); err != nil {
		log.Printf("set join application by groupid and status to redis failed, err: %v", err)
	}

	// 删除用户加入用户组的申请缓存
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, userID); err != nil {
		log.Printf("delete join application by userid and groupid to redis failed, err: %v", err)
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
			return appErrors.ErrRolePermissionDenied.WithError(err)
		}
		//更新申请记录状态
		if err := s.joinApplicationDao.UpdateStatus(ctx, requestID, "rejected", tx); err != nil {
			return appErrors.ErrJoinApplicationUpdateFailed.WithError(err)
		}
		//更新拒绝理由
		if err := s.joinApplicationDao.UpdateRejectReason(ctx, requestID, rejectReason, tx); err != nil {
			return appErrors.ErrJoinApplicationUpdateFailed.WithError(err)
		}

		// 接受审批拒绝的记录，用于缓存
		newApplication, err := s.joinApplicationDao.GetByRequestID(ctx, requestID, tx)
		if err != nil {
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		application = *newApplication
		return nil
	})
	if err != nil {
		return err
	}

	// 删除申请缓存

	// 删除用户申请缓存
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, userID); err != nil {
		log.Printf("delete join application by userid and groupid to redis failed, err: %v", err)
	}

	// 删除用户组所有申请缓存（所有状态）
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		log.Printf("delete join application by groupid to redis failed, err: %v", err)
	}

	// 更新用户组待审批状态缓存
	existGroupPendingApplication, err := s.joinApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, "pending")
	if existGroupPendingApplication !=nil && len(existGroupPendingApplication)>0{
		for i:=0;i<len(existGroupPendingApplication);i++{
			if existGroupPendingApplication[i].RequestID == requestID {
				existGroupPendingApplication=append(existGroupPendingApplication[:i],existGroupPendingApplication[i+1:]...)
				break
			}
		}
	}
	if err := s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "pending", existGroupPendingApplication); err != nil {
		log.Printf("set join application by groupid and status to redis failed, err: %v", err)
	}

	// 更新用户组加入申请审批拒绝缓存
	existGroupRejectedApplication, err := s.joinApplicationRedisDAO.GetByGroupIDAndStatus(ctx, groupID, "rejected")
	if existGroupRejectedApplication ==nil{
		existGroupRejectedApplication = []*models.JoinApplication{}
	}
	existGroupRejectedApplication = append(existGroupRejectedApplication, &application)
	if err := s.joinApplicationRedisDAO.SetByGroupIDAndStatus(ctx, groupID, "rejected", existGroupRejectedApplication); err != nil {
		log.Printf("set join application by groupid and status to redis failed, err: %v", err)
	}

	// 删除用户加入用户组的申请缓存
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, userID); err != nil {
		log.Printf("delete join application by userid and groupid to redis failed, err: %v", err)
	}

	return nil
}

// 删除用户组
func (s *GroupsService) DeleteGroup(ctx context.Context, groupID, operatorID int) error {
	var members []*models.GroupMember
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查操作员权限
		if err := s.CheckMemberPermission(ctx, groupID, operatorID); err != nil {
			return appErrors.ErrRolePermissionDenied.WithError(err)
		}
		//删除用户组
		if err := s.groupDao.Delete(ctx, groupID, tx); err != nil {
			return appErrors.ErrGroupDeletionFailed.WithError(err)
		}
		//删除用户组成员
		groupMembers, err := s.groupMemberDao.GetMembersByGroupID(ctx, groupID, tx)
		if err != nil {
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		members = groupMembers
		for _, member := range groupMembers {
			if err := s.groupMemberDao.Delete(ctx, groupID, member.UserID, tx); err != nil {
				return appErrors.ErrGroupMemberDeletionFailed.WithError(err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 删除用户组及 成员缓存（组内记录+申请记录）
	if err := s.groupRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		log.Printf("delete group cache failed, err: %v", err)
	}
	if err := s.groupMemberRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		log.Printf("delete group members cache failed, err: %v", err)
	}
	for _, member := range members {
		if err := s.groupMemberRedisDAO.DeleteCacheByGroupIDAndUserID(ctx, groupID, member.UserID); err != nil {
			log.Printf("delete group member cache failed, err: %v", err)
		}
		if err:=s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndUserID(ctx,groupID,member.UserID);err!=nil{
			log.Printf("delete join application by userid and groupid to redis failed, err: %v", err)
		}
	}

	// 删除用户组相关所有申请缓存
	if err := s.joinApplicationRedisDAO.DeleteCacheByGroupID(ctx, groupID); err != nil {
		log.Printf("delete join application by groupid to redis failed, err: %v", err)
	}
	if err:=s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndStatus(ctx,groupID,"pending");err!=nil{
		log.Printf("delete join application by groupid and status to redis failed, err: %v", err)
	}
	if err:=s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndStatus(ctx,groupID,"accepted");err!=nil{
		log.Printf("delete join application by groupid and status to redis failed, err: %v", err)
	}
	if err:=s.joinApplicationRedisDAO.DeleteCacheByGroupIDAndStatus(ctx,groupID,"rejected");err!=nil{
		log.Printf("delete join application by groupid and status to redis failed, err: %v", err)
	}

	return nil
}

// 查询当前登录用户在指定用户组中的状态，包括未关联、申请中、普通成员、管理员等(返回申请记录，可在handlers层根据记录的status构建对应的响应)
func (s *GroupsService) GetUserGroupStatus(ctx context.Context, groupID, userID int) (string, int, error) {
	var status string
	var requestID int
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查用户组是否存在
		_, err := s.groupDao.GetByGroupID(ctx, groupID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrGroupNotFound
			}
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		// 检查是否为组成员
		member, err := s.groupMemberDao.GetMemberByGroupIDAndUserID(ctx, groupID, userID, tx)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrDatabaseOperation.WithError(err)
			}
		}
		if member != nil {
			status = member.Role
			requestID = 0
			return nil
		}
		// 非组成员，查看申请记录
		Application, err := s.joinApplicationDao.GetByGroupIDAndUserID(ctx, groupID, userID, tx)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrDatabaseOperation.WithError(err)
			}
		}
		if Application == nil {
			status = "none"
			requestID = 0
		} else {
			status = Application.Status
			requestID = Application.RequestID
		}
		return nil
	})
	if err != nil {
		return "", 0, err
	}
	return status, requestID, nil
}
