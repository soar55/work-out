package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/koron/go-dproxy"
)

type Auth struct {
}

// セッションに保存するユーザ情報
type SessionUser struct {
	Id    uint   `json:"id"`
	Token string `json:"token"`
}

func (Auth) CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		loginUser, err := dproxy.New(session.Get("loginUser")).String()

		var loginInfo SessionUser
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			// Abort()を記述すると、以降の処理は実行されない。
			c.Abort()
		} else {
			err := json.Unmarshal([]byte(loginUser), &loginInfo)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
				c.Abort()
			}
		}
		// TODO: DBからユーザ情報を取得し、セッションのユーザ情報と照合する。

		// Next()前に記述した処理は、ルーティング処理の前に実行
		c.Next()
		// Next()後に記述した処理は、ルーティング処理の後に実行
	}
}
