package service

import (
	"context"
	"TeamTickBackend/dal/models"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mock 实现 ---
type mockTransactionManager struct{ mock.Mock }

func (m *mockTransactionManager) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	args := m.Called(ctx, fn)
	// 执行回调函数，传入nil作为tx参数
	err := fn(nil)
	// 返回回调错误或配置的mock错误
	if args.Error(0) != nil {
		return args.Error(0)
	}
	return err
}

type mockGroupDAO struct{ mock.Mock }

func (m *mockGroupDAO) GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) (*models.Group, error) {
	args := m.Called(ctx, groupID, tx)
	groupArg := args.Get(0)
	if groupArg == nil {
		return nil, args.Error(1)
	}
	return groupArg.(*models.Group), args.Error(1)
}


type mockGroupMemberDAO struct{ mock.Mock }

func (m *mockGroupMemberDAO) GetMembersByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.GroupMember, error) {
	args := m.Called(ctx, groupID, tx)
	membersArg := args.Get(0)
	if membersArg == nil {
		return nil, args.Error(1)
	}
	return membersArg.([]*models.GroupMember), args.Error(1)
}
