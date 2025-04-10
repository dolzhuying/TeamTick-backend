package dao

import (
	"context"

	"gorm.io/gorm"
)

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type TransactionManagerImpl struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) TransactionManager {
	return &TransactionManagerImpl{db: db}
}

// WithTransaction 提供函数式事务API，推荐
func (t *TransactionManagerImpl) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return t.db.WithContext(ctx).Transaction(fn)
}
