package dao

import (
	"TeamTickBackend/dal/dao/impl"
)

type DAO struct {
	UserDAO             UserDAO
	TaskDAO             TaskDAO
	GroupDAO            GroupDAO
	GroupMemberDAO      GroupMemberDAO
	JoinApplicationDAO  JoinApplicationDAO
	CheckApplicationDAO CheckApplicationDAO
	TaskRecordDAO       TaskRecordDAO
}

// 为service层提供dao全局访问接口
var DAOInstance = newDAO()

func newDAO() *DAO {
	return &DAO{
		UserDAO:             &impl.UserDAOMySQLImpl{},
		TaskDAO:             &impl.TaskDAOMySQLImpl{},
		GroupDAO:            &impl.GroupDAOMySQLImpl{},
		GroupMemberDAO:      &impl.GroupMemberDAOMySQLImpl{},
		JoinApplicationDAO:  &impl.JoinApplicationDAOMySQLImpl{},
		CheckApplicationDAO: &impl.CheckApplicationDAOMySQLImpl{},
		TaskRecordDAO:       &impl.TaskRecordDAOMySQLImpl{},
	}
}
