package dao

import (
	"TeamTickBackend/global"
	"context"

	"gorm.io/gorm"
)

// 事务的实例化和调用应该在service层具体分析
type Transaction struct {
	tx *gorm.DB
}

//手动开启事务，并不推荐
func NewTransaction(ctx context.Context) (*Transaction, error) {
	tx := global.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &Transaction{tx: tx}, nil
}

func (t *Transaction) Commit() error {
	return t.tx.Commit().Error
}

func (t *Transaction) Rollback() error {
	return t.tx.Rollback().Error
}

func (t *Transaction) GetTx() *gorm.DB {
	return t.tx
}

// WithTransaction 提供函数式事务API，推荐
func WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
    return global.DB.WithContext(ctx).Transaction(fn)
}
