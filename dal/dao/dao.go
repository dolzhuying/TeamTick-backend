package dao

type DAO struct {
	userDAO             UserDAO
	taskDAO             TaskDAO
	groupDAO            GroupDAO
	groupMemberDAO      GroupMemberDAO
	joinApplicationDAO  JoinApplicationDAO
	checkApplicationDAO CheckApplicationDAO
	taskRecordDAO       TaskRecordDAO
}

//为service层提供dao全局访问接口
var Factory = newDAO()

func newDAO() *DAO{
	return &DAO{
		userDAO:             &UserDAOImpl{},
		taskDAO:             &TaskDAOImpl{},
		groupDAO:            &GroupDAOImpl{},
		groupMemberDAO:      &GroupMemberDAOImpl{},
		joinApplicationDAO:  &JoinApplicationDAOImpl{},
		checkApplicationDAO: &CheckApplicationDAOImpl{},
		taskRecordDAO:       &TaskRecordDAOImpl{},
	}
}

func (f *DAO) GetUserDAO() UserDAO {
	return f.userDAO
}

func (f *DAO) GetTaskDAO() TaskDAO {
	return f.taskDAO
}

func (f *DAO) GetGroupDAO() GroupDAO {
	return f.groupDAO
}

func (f *DAO) GetGroupMemberDAO() GroupMemberDAO {
	return f.groupMemberDAO
}

func (f *DAO) GetJoinApplicationDAO() JoinApplicationDAO {
	return f.joinApplicationDAO
}

func (f *DAO) GetCheckApplicationDAO() CheckApplicationDAO {
	return f.checkApplicationDAO
}

func (f *DAO) GetTaskRecordDAO() TaskRecordDAO {
	return f.taskRecordDAO
}


