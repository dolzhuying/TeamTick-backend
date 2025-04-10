package dao

import (
	"TeamTickBackend/dal/dao/impl"

	"gorm.io/gorm"
)

type DAOFactory struct {
	Db                 *gorm.DB
	TransactionManager TransactionManager

	UserDAO             UserDAO
	GroupDAO            GroupDAO
	TaskDAO             TaskDAO
	GroupMemberDAO      GroupMemberDAO
	TaskRecordDAO       TaskRecordDAO
	JoinApplicationDAO  JoinApplicationDAO
	CheckApplicationDAO CheckApplicationDAO
}

func NewDAOFactory(db *gorm.DB) *DAOFactory {
	return &DAOFactory{
		Db:                  db,
		TransactionManager:  NewTransactionManager(db),
		UserDAO:             &impl.UserDAOMySQLImpl{DB: db},
		GroupDAO:            &impl.GroupDAOMySQLImpl{DB: db},
		TaskDAO:             &impl.TaskDAOMySQLImpl{DB: db},
		GroupMemberDAO:      &impl.GroupMemberDAOMySQLImpl{DB: db},
		TaskRecordDAO:       &impl.TaskRecordDAOMySQLImpl{DB: db},
		JoinApplicationDAO:  &impl.JoinApplicationDAOMySQLImpl{DB: db},
		CheckApplicationDAO: &impl.CheckApplicationDAOMySQLImpl{DB: db},
	}
}
