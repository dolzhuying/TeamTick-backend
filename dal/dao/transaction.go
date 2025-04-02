package dao

import (
	"TeamTickBackend/global"

	"gorm.io/gorm"
)

//事务的实例化和调用应该在service层具体情况具体分析
type Transaction struct {
	tx *gorm.DB
}

func NewTransaction() (*Transaction, error) {
	tx := global.DB.Begin()
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
