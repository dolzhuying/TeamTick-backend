package pkg

import (
	"TeamTickBackend/config"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 实例化考虑依赖注入或者全局实例化？
type JwtToken struct {
	jwtConfig *config.JWTConfig
}

//暂时使用全局实例化，
var JwtTokenInstance *JwtToken

func NewJwtToken() error {
	jwtConfig, err := config.GetJWTConfig()
	if err != nil {
		return errors.New(err.Error())
	}
	JwtTokenInstance=&JwtToken{
		jwtConfig,
	}
	return nil
}

type TokenClaims struct {
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
	jwt.RegisteredClaims
}

type jwtPayload struct {
	Username string
	UserID   int
}

// 根据hs256算法以及用户id、用户名生成jwt
func (s *JwtToken) GenerateJWTToken(username string, userID int) (string, error) {
	now := time.Now()
	claims := TokenClaims{
		Username: username,
		UserID:   userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtConfig.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.jwtConfig.Issuer,
			Subject:   fmt.Sprintf("user:%d", userID),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtConfig.SecretKey)
	if err != nil {
		log.Printf("invalid JWT signature: %v", err)
		return "", fmt.Errorf("failed to generate JWT token: %w", err)
	}
	return signedToken, nil
}

// 解析JWT，错误日志处理待完善
func (s *JwtToken) ParseJWTToken(tokenString string) (jwtPayload, error) {
	tokenString = strings.TrimSpace(tokenString)
	if len(tokenString) > 7 && strings.ToUpper(tokenString[:7]) == "BEARER " {
		tokenString = tokenString[7:]
	}

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected token signing method: %v", token.Header["alg"])
		}
		return s.jwtConfig.SecretKey, nil
	})

	//错误解析
	jwtErrPayload := jwtPayload{}
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return jwtErrPayload, jwt.ErrTokenMalformed
		} else if errors.Is(err, jwt.ErrTokenUnverifiable) {
			return jwtErrPayload, jwt.ErrTokenUnverifiable
		} else if errors.Is(err, jwt.ErrTokenInvalidClaims) {
			return jwtErrPayload, jwt.ErrTokenInvalidClaims
		} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return jwtErrPayload, jwt.ErrTokenSignatureInvalid
		}
		return jwtErrPayload, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return jwtErrPayload, errors.New("invalid token claims")
	}
	now := time.Now()
	if claims.ExpiresAt != nil && now.After(claims.ExpiresAt.Time) {
		return jwtErrPayload, jwt.ErrTokenExpired
	}
	if claims.NotBefore != nil && now.Before(claims.NotBefore.Time) {
		return jwtErrPayload, jwt.ErrTokenNotValidYet
	}
	return jwtPayload{
		Username: claims.Username,
		UserID:   claims.UserID,
	}, nil
}
