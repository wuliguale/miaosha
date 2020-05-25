package common

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
)

var hmacSampleSecret []byte = []byte("miaosha-demo-jwt-secret")

//生成jwt
func JwtSign(claims jwt.MapClaims) (string, error) {
	claims["iss"] = "1dFjvu6WWars9bahttTbWuqFOEGIWB2G"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(hmacSampleSecret)
}


//解析jwt
func JwtParse(tokenString string) (jwt.MapClaims, error){
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return hmacSampleSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("get claims fail")
	}
}
