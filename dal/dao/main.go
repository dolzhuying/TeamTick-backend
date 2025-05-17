package dao

import (
	mysqlImpl "TeamTickBackend/dal/dao/impl/mysql"
	redisImpl "TeamTickBackend/dal/dao/impl/redis"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type DAOFactory struct {
	Db                 *gorm.DB
	RedisClient        *redis.Client
	TransactionManager TransactionManager

	UserDAO             UserDAO
	GroupDAO            GroupDAO
	TaskDAO             TaskDAO
	GroupMemberDAO      GroupMemberDAO
	TaskRecordDAO       TaskRecordDAO
	JoinApplicationDAO  JoinApplicationDAO
	CheckApplicationDAO CheckApplicationDAO
	StatisticsDAO       StatisticsDAO

	TaskRecordRedisDAO  TaskRecordRedisDAO
	TaskRedisDAO        TaskRedisDAO
	GroupRedisDAO       GroupRedisDAO
	GroupMemberRedisDAO GroupMemberRedisDAO
	JoinApplicationRedisDAO JoinApplicationRedisDAO
	CheckApplicationRedisDAO CheckApplicationRedisDAO
}

func NewDAOFactory(db *gorm.DB, redisClient *redis.Client) *DAOFactory {

	return &DAOFactory{
		Db:                 db,
		RedisClient:        redisClient,
		TransactionManager: NewTransactionManager(db),

		UserDAO:             &mysqlImpl.UserDAOMySQLImpl{DB: db},
		GroupDAO:            &mysqlImpl.GroupDAOMySQLImpl{DB: db},
		TaskDAO:             &mysqlImpl.TaskDAOMySQLImpl{DB: db},
		GroupMemberDAO:      &mysqlImpl.GroupMemberDAOMySQLImpl{DB: db},
		TaskRecordDAO:       &mysqlImpl.TaskRecordDAOMySQLImpl{DB: db},
		JoinApplicationDAO:  &mysqlImpl.JoinApplicationDAOMySQLImpl{DB: db},
		CheckApplicationDAO: &mysqlImpl.CheckApplicationDAOMySQLImpl{DB: db},
		StatisticsDAO:       &mysqlImpl.StatisticsDAOMySQLImpl{DB: db},

		TaskRecordRedisDAO:  &redisImpl.TaskRecordDAORedisImpl{Client: redisClient},
		TaskRedisDAO:        &redisImpl.TaskRedisDAOImpl{Client: redisClient},
		GroupRedisDAO:       &redisImpl.GroupRedisDAOImpl{Client: redisClient},
		GroupMemberRedisDAO: &redisImpl.GroupMemberRedisDAOImpl{Client: redisClient},
		JoinApplicationRedisDAO: &redisImpl.JoinApplicationRedisDAO{Client: redisClient},
		CheckApplicationRedisDAO: &redisImpl.CheckApplicationRedisDAOImpl{Client: redisClient},
	}
}
