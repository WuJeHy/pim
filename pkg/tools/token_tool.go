package tools

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"time"
)

const (
	TokenISS          = "iss"
	TokenSignedString = "token"
)

type TokenClaims struct {
	UserID int64  `json:"uid,omitempty"`
	Pf     uint32 `json:"pf,omitempty"`
	Level  int    `json:"le,omitempty"`
	jwt.StandardClaims
}

func CheckHttpToken(ctx *gin.Context) (token *TokenClaims, err error) {
	token, err = CheckHttpTokenNotResp(ctx)
	if err != nil {

		RespData(ctx, 401, err.Error())
		return
	}
	return
}

// CheckHttpTokenNotResp token 解析器
func CheckHttpTokenNotResp(ctx *gin.Context) (tokenInfo *TokenClaims, err error) {

	tokenStrUnCheck := ctx.GetHeader("Authorization")

	headerToken := tokenStrUnCheck[:7]

	if headerToken != "Bearer " {
		return nil, errors.New("get token Bearer fail")
	}
	jwtToken := tokenStrUnCheck[7:]

	// 解析token

	return ParseToken(jwtToken)
}

func ParseToken(jwtToken string) (tokenInfo *TokenClaims, err error) {
	tokenClaims, err := jwt.ParseWithClaims(jwtToken, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(TokenSignedString), nil
	})

	if err != nil {
		return
	}

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*TokenClaims); ok && tokenClaims.Valid {
			// 时间有效检验
			if claims.Issuer != TokenISS {
				return nil, errors.New("not my issuer")
			}

			currentTime := time.Now().Unix()
			// 过期时间校验
			if claims.ExpiresAt < currentTime {

				return nil, errors.New("token expires at ")
			}

			if claims.UserID == 0 {
				// 用户id 校验失败
				return nil, errors.New("user id not found ")
			}

			return claims, nil

		} else {
			return nil, errors.New("token info fail")
		}
	} else {

		return nil, errors.New("not found token info ")
	}

}

func GenToken(Uid int64, pf int, level int) (string, error) {

	nowTime := time.Now()
	expireTime := nowTime.Add(1440 * time.Minute)
	claims := TokenClaims{
		UserID: Uid,
		Pf:     uint32(pf),
		Level:  level,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  nowTime.Unix(),
			ExpiresAt: expireTime.Unix(),
			Issuer:    TokenISS,
		},
	}

	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenObj.SignedString([]byte(TokenSignedString))
	return token, err

}
