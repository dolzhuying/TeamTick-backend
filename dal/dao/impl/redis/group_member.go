package redisImpl

import (
	"TeamTickBackend/dal/models"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	GroupMemberKeyPrefix          = "group_member:"
	GroupMemberKeyPrefixByGroupID = "group_member:group:"
)

type GroupMemberRedisDAOImpl struct {
	Client *redis.Client
}

func buildGroupMemberKey(groupID, userID int) string {
	return fmt.Sprintf("%s%d:%d", GroupMemberKeyPrefix, groupID, userID)
}

func buildGroupMemberKeyByGroupID(groupID int) string {
	return fmt.Sprintf("%s%d", GroupMemberKeyPrefixByGroupID, groupID)
}

// 通过GroupID查询GroupMembers缓存
func (dao *GroupMemberRedisDAOImpl) GetMembersByGroupID(ctx context.Context, groupID int, tx ...*redis.Client) ([]*models.GroupMember, error) {
	client := dao.Client
	if len(tx) > 0 && tx[0] != nil {
		client = tx[0]
	}
	key := buildGroupMemberKeyByGroupID(groupID)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var members []*models.GroupMember
	err = json.Unmarshal(data, &members)
	if err != nil {
		return nil, err
	}
	return members, nil
}

// 设置GroupMembers缓存
func (dao *GroupMemberRedisDAOImpl) SetMembersByGroupID(ctx context.Context,groupID int,members []*models.GroupMember) error{
	client:=dao.Client
	data,err:=json.Marshal(members)
	if err!=nil{
		return err
	}
	key:=buildGroupMemberKeyByGroupID(groupID)
	return client.Set(ctx,key,data,DefaultExpireTime).Err()
}

// 删除GroupMembers缓存
func (dao *GroupMemberRedisDAOImpl) DeleteCacheByGroupID(ctx context.Context,groupID int) error{
	client:=dao.Client
	key:=buildGroupMemberKeyByGroupID(groupID)
	return client.Del(ctx,key).Err()
}

// 通过GroupID和UserID查询GroupMember缓存
func (dao *GroupMemberRedisDAOImpl) GetMemberByGroupIDAndUserID(ctx context.Context,groupID,userID int,tx ...*redis.Client) (*models.GroupMember,error){
	client:=dao.Client
	if len(tx)>0&&tx[0]!=nil{
		client=tx[0]
	}
	key:=buildGroupMemberKey(groupID,userID)
	data,err:=client.Get(ctx,key).Bytes()
	if err!=nil{
		if err==redis.Nil{
			return nil,nil
		}
		return nil,err
	}
	var member models.GroupMember
	err=json.Unmarshal(data,&member)
	if err!=nil{
		return nil,err
	}
	return &member,nil
}

// 设置GroupMember缓存
func (dao *GroupMemberRedisDAOImpl) SetMemberByGroupIDAndUserID(ctx context.Context,member *models.GroupMember) error{
	client:=dao.Client
	data,err:=json.Marshal(member)
	if err!=nil{
		return err
	}
	key:=buildGroupMemberKey(member.GroupID,member.UserID)
	return client.Set(ctx,key,data,DefaultExpireTime).Err()
}

// 删除GroupMember缓存
func (dao *GroupMemberRedisDAOImpl) DeleteCacheByGroupIDAndUserID(ctx context.Context,groupID,userID int) error{
	client:=dao.Client
	key:=buildGroupMemberKey(groupID,userID)
	return client.Del(ctx,key).Err()
	
}