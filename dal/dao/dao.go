package dao

import(
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

//为service层提供dao全局访问接口
var DAOInstance = newDAO()

func newDAO() *DAO{
	return &DAO{
		UserDAO:             &impl.MySQLUserDAOImpl{},
		TaskDAO:             &impl.MySQLTaskDAOImpl{},
		GroupDAO:            &impl.MySQLGroupDAOImpl{},
		GroupMemberDAO:      &impl.MySQLGroupMemberDAOImpl{},
		JoinApplicationDAO:  &impl.MySQLJoinApplicationDAOImpl{},
		CheckApplicationDAO: &impl.MySQLCheckApplicationDAOImpl{},
		TaskRecordDAO:       &impl.MySQLTaskRecordDAOImpl{},
	}
}


