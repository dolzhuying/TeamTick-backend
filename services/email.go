package service

import (
	"TeamTickBackend/dal/dao"
	redisImpl "TeamTickBackend/dal/dao/impl/redis"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	appErrors "TeamTickBackend/pkg/errors"

	"github.com/redis/go-redis/v9"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	EmailRedisDAO dao.EmailRedisDAO

	dialer *gomail.Dialer
}

func NewEmailService(emailRedisDAO dao.EmailRedisDAO, smtpHost string, smtpPort int, smtpUsername, smtpPassword string) *EmailService {
	d := gomail.NewDialer(smtpHost, smtpPort, smtpUsername, smtpPassword)
	return &EmailService{
		EmailRedisDAO: emailRedisDAO,
		dialer:        d,
	}
}

// GenerateVerificationCode 生成随机验证码
func (s *EmailService) GenerateVerificationCode(length int) (string, error) {
	const letters = "0123456789"
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		result[i] = letters[num.Int64()]
	}
	log.Println("生成验证码：", string(result))
	return string(result), nil
}

// SendVerificationEmail 发送验证码邮件
func (s *EmailService) SendVerificationEmail(ctx context.Context, email, code string) error {
	// 检查发送频率
	rateKey := fmt.Sprintf("email:rate:%s", email)
	count, err := s.EmailRedisDAO.(*redisImpl.EmailRedisDAOImpl).Client.Get(ctx, rateKey).Int()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("检查发送频率失败: %w", err)
	}
	if count >= 5 { // 5分钟内最多发送5次
		return appErrors.ErrTooManyRequests
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.dialer.Username) //  需要配置发件人邮箱
	m.SetHeader("To", email)
	m.SetHeader("Subject", "您的 TeamTick 验证码")

	imagePath := "src/logo.png"
	imageFilename := "logo.png" // 将用作 cid

	m.Embed(imagePath)
	cidForHTML := imageFilename

	// 构建邮件正文 HTML
	bodyHTML := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>TeamTick 验证码</title>
<style>
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol";
    margin: 0;
    padding: 0;
    background-color: #f0f2f5; /* 类似图片中的浅灰色背景 */
  }
  .container {
    max-width: 500px;
    margin: 40px auto;
    background-color: #ffffff;
    border: 1px solid #e0e0e0; /* 轻微边框 */
    border-radius: 8px; /* 圆角 */
    padding: 30px;
    text-align: center;
    box-shadow: 0 4px 12px rgba(0,0,0,0.05);
  }
  .logo {
    max-width: 60px; /* 根据您的logo调整 */
    margin-bottom: 25px;
  }
  h1 {
    font-size: 24px;
    color: #1c1e21; /* 深灰色字体 */
    margin-bottom: 15px;
  }
  .code-intro {
    font-size: 16px;
    color: #505050; /* 中灰色字体 */
    margin-bottom: 10px;
    line-height: 1.5;
  }
  .verification-code {
    font-size: 38px;
    font-weight: bold;
    color: #1c1e21;
    margin: 25px 0;
    letter-spacing: 2px; /* 字符间距 */
  }
  hr {
    border: none;
    border-top: 1px solid #e0e0e0;
    margin: 30px 0;
  }
  .footer-text {
    font-size: 13px;
    color: #808080; /* 浅灰色字体 */
    line-height: 1.6;
  }
  .footer-text a {
    color: #007bff;
    text-decoration: none;
  }
</style>
</head>
<body>
  <div class="container">
    <img src="cid:%s" alt="TeamTick Logo" class="logo">
    <h1>注册 TeamTick</h1>
    <p class="code-intro">您正在请求注册 TeamTick。<br>您的动态验证码是：</p>
    <div class="verification-code">%s</div>
    <hr>
    <p class="footer-text">此验证码将在 30 分钟后失效。</p>
    <p class="footer-text">如果您未请求此操作，或并非您本人操作，请忽略本邮件。可能是其他人误填了您的邮箱地址。</p>
  </div>
</body>
</html>
	`, cidForHTML, code)

	m.SetBody("text/html", bodyHTML)

	sendErr := s.dialer.DialAndSend(m)
	if sendErr != nil {
		// 检查错误是否与文件有关，例如文件未找到
		// os.IsNotExist 需要一个 error 类型的参数
		if os.IsNotExist(sendErr) { // 使用 sendErr 进行检查
			log.Printf("发送邮件失败，可能是因为图片文件 %s 未找到: %v\n", imagePath, sendErr)
			// 尝试发送不带图片的邮件作为回退
			mWithoutImage := gomail.NewMessage()
			mWithoutImage.SetHeader("From", s.dialer.Username)
			mWithoutImage.SetHeader("To", email)
			mWithoutImage.SetHeader("Subject", "您的 TeamTick 验证码")
			mWithoutImage.SetBody("text/html", fmt.Sprintf("您的验证码是: <b>%s</b>，有效期5分钟。 (图片无法加载)", code))
			if errFallback := s.dialer.DialAndSend(mWithoutImage); errFallback != nil {
				log.Printf("回退邮件发送失败: %v\n", errFallback)
				return fmt.Errorf("发送邮件失败（回退邮件也失败 Original Err: %v, Fallback Err: %v）", sendErr, errFallback)
			}

			log.Println("已发送不含图片的回退邮件")

			return fmt.Errorf("原始邮件发送失败（图片问题），但已发送回退邮件: %w", sendErr)
		} else {
			// 其他类型的发送错误
			return fmt.Errorf("发送邮件失败: %w", sendErr)
		}
	} else {
		fmt.Println("发送邮件成功")
	}

	// 更新发送频率计数
	err = s.EmailRedisDAO.(*redisImpl.EmailRedisDAOImpl).Client.Set(ctx, rateKey, count+1, 5*time.Minute).Err()
	if err != nil {
		log.Printf("更新发送频率计数失败: %v", err)
	}

	// 将验证码存入 Redis
	if err := s.EmailRedisDAO.SetVerificationCodeByEmail(ctx, email, code); err != nil {
		return fmt.Errorf("设置验证码到 Redis 失败: %w", err)
	}
	return nil
}

// VerifyEmailCode 校验邮箱验证码
func (s *EmailService) VerifyEmailCode(ctx context.Context, email, code string) (bool, error) {
	cachedCode, err := s.EmailRedisDAO.GetVerificationCodeByEmail(ctx, email)
	if err != nil {
		if err == redis.Nil {
			return false, appErrors.ErrVerificationCodeExpiredOrNotFound
		}
		return false, fmt.Errorf("从 Redis 获取验证码失败: %w", err)
	}
	if cachedCode == "" {
		return false, appErrors.ErrVerificationCodeExpiredOrNotFound
	}
	if cachedCode != code {
		return false, appErrors.ErrInvalidVerificationCode
	}
	return true, nil
}
