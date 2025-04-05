
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

// 实例化考虑依赖注入？
type JwtToken struct {
	jwtConfig *config.JWTConfig
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
	if err != nil {
		jwtErrPayload := jwtPayload{}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return jwtErrPayload, errors.New("invalid token format")
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return jwtErrPayload, errors.New("token expired")
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return jwtErrPayload, errors.New("token not valid yet")
		} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return jwtErrPayload, errors.New("invalid token signature")
		}
		return jwtErrPayload, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return jwtPayload{}, errors.New("invalid token")
	}
	return jwtPayload{
		Username: claims.Username,
		UserID:   claims.UserID,
	}, nil
}
