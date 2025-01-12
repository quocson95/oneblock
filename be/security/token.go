package security

import (
	"be/common"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

type UsertClaims struct {
	Id   int
	Role common.RoleUser
	jwt.Claims
}

var SecretJwtAuth = "secretKey"

type TokenResponse struct {
	Token      string `json:"token,omitempty"`
	ExpireUnix int64  `json:"expire_unix,omitempty"`
}

func CreateToken(user *common.User) (*TokenResponse, error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}
	if user.ID < 0 || user.Role <= 0 {
		return nil, errors.New("user invalid")
	}
	now := time.Now()
	expireAfter := time.Duration(7 * 24 * time.Hour)
	expireAt := now.Add(expireAfter)
	claims := &UsertClaims{
		Id:   int(user.ID),
		Role: user.Role,
		Claims: jwt.StandardClaims{
			Issuer:    "system",
			NotBefore: now.Unix(),
			// Id:        models.Sha265Random(),
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: expireAt.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Sign the token
	signedToken, err := token.SignedString([]byte(SecretJwtAuth))
	if err != nil {
		return nil, err
	}
	return &TokenResponse{
		Token:      signedToken,
		ExpireUnix: expireAt.Unix(),
	}, nil
}

func VerifyToken(token string) (*common.User, error) {
	// claim :=var claims *Claims
	// var claims jwt.Claims =
	tokenParse, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretJwtAuth), nil
	})
	if err != nil {
		return nil, err
	}
	if !tokenParse.Valid {
		return nil, errors.New("token not valid")
	}
	user, ok := tokenParse.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("parse to user failed")
	}
	return &common.User{
		Model: gorm.Model{
			ID: uint(user["Id"].(float64)),
		},
	}, nil
}
