package redisImpl

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type EmailRedisDAOImpl struct {
	Client *redis.Client
}

const (
	EmailVerificationCodeKey    = "email:%s"
	EmailVerificationExpireTime = 30 * time.Minute // 邮箱验证码过期时间
)

func buildKeyByEmail(email string) string {
	return fmt.Sprintf(EmailVerificationCodeKey, email)
}

func (dao *EmailRedisDAOImpl) GetVerificationCodeByEmail(ctx context.Context, email string, tx ...*redis.Client) (string, error) {
	client := dao.Client
	if len(tx) > 0 {
		client = tx[0]
	}
	key := buildKeyByEmail(email)
	code, err := client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	return code, nil
}

func (dao *EmailRedisDAOImpl) SetVerificationCodeByEmail(ctx context.Context, email string, code string) error {
	client := dao.Client
	key := buildKeyByEmail(email)
	return client.Set(ctx, key, code, EmailVerificationExpireTime).Err()
}

func (dao *EmailRedisDAOImpl) DeleteCacheByEmail(ctx context.Context, email string) error {
	client := dao.Client
	key := buildKeyByEmail(email)
	return client.Del(ctx, key).Err()
}
