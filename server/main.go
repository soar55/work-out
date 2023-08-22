package main

// すべてのコードはpackageに属している。
// mainは特別なpackageで、main package内のmain()がプログラムのエントリーポイントとなる。

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"server/middleware"
	"server/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// gormのモデル定義
type User struct {
	gorm.Model
	Name     string `gorm:"size:32"`
	Mail     string `gorm:"size:256"`
	GoogleID string `gorm:"size:64"`
	Picture  string `gorm:"size:256"`
}

// Googleユーザ情報
type ProfileData struct {
	Name     string `json:"name"`
	Mail     string `json:"email"`
	GoogleID string `json:"sub"`
	Picture  string `json:"picture"`
}

// ginというwebフレームワークを使うが、PHPのLaravelと比べると大変シンプルなフレームワークなので、コード記述量が減ったり、便利な機能がまとまって提供されている程度。
// アーキテクチャ(デザインパターン)は、自身で設計する必要がある。
func main() {
	// ginの初期化
	// New()だと、素の状態でインスタンスが生成される。
	// Default()だと、loggerとrecoveryのミドルウェアが設定されたインスタンスが生成される。
	router := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	// ルート(APIサーバの予定なので、一旦生存確認のPing代わり)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World!",
		})
	})

	// GOOGLE AUTH
	router.GET("/auth", func(c *gin.Context) {
		authService := service.Auth{}
		authGoogle := authService.GetConnect()
		url := authGoogle.AuthCodeURL("statedayo-")
		c.Redirect(http.StatusTemporaryRedirect, url)
	})

	// GOOGLE AUTH CALLBACK
	router.GET("/auth/google/callback", func(c *gin.Context) {
		code := c.Query("code")
		authService := service.Auth{}
		authGoogle := authService.GetConnect()
		token, err := authGoogle.Exchange(c, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error",
			})
		}

		client := authGoogle.Client(c, token)
		profile, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error",
			})
		}
		defer profile.Body.Close()
		response, err := ioutil.ReadAll(profile.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error",
			})
		}

		// Googleユーザ情報
		var profileData ProfileData
		if err := json.Unmarshal(response, &profileData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error",
			})
		}

		fmt.Println(profileData)

		// TODO: DBにメールアドレスが登録されているか確認。
		// されていれば、ログイン。されていなければ、新規登録。

		// DB接続
		dsn := os.Getenv("DB_USERNAME") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(127.0.0.1:3306)/" + os.Getenv("DB_DATABASE") + "?charset=utf8mb4&parseTime=True&loc=Local"
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

		db.AutoMigrate(&User{})

		// ユーザ情報をDBに登録
		// TODO: subをハッシュ化
		insertUser := &User{Name: profileData.Name, GoogleID: profileData.GoogleID, Picture: profileData.Picture, Mail: profileData.Mail}

		db.Create(insertUser)
		fmt.Printf("insert ID: %d, Name: %s, Picture: %s\n",
			insertUser.ID, insertUser.Name, insertUser.Picture)

		sessionUser := &middleware.SessionUser{Id: insertUser.ID, Token: insertUser.GoogleID}
		loginUser, err := json.Marshal(sessionUser)
		// TODO:セッションにユーザ情報を保存
		session := sessions.Default(c)
		session.Set("loginUser", string(loginUser))
		session.Save()

		c.JSON(http.StatusOK, gin.H{
			"message":  "Success",
			"response": string(response),
		})
	})

	// 認証済みでしかつかえない機能
	// TODO: セッションチェックをミドルウェアにし、処理の前に差し込むようにする。
	// ginのグループ機能を使って、認証済みのグループを作成する。
	authMiddleware := middleware.Auth{}
	mypageGroup := router.Group("/mypage")
	mypageGroup.Use(authMiddleware.CheckAuth())
	{
		mypageGroup.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Loged in",
			})
		})
	}

	router.Run(":8080")
}

// あとでつかう。ハッシュ化関数
func HashStr(str string) string {
	hashStr := ""

	// 対象文字列をバイト型スライスに変換
	byteStr := []byte(str)

	// バイト型スライスをハッシュ化
	sha256 := sha256.Sum256(byteStr)
	hashStr = hex.EncodeToString(sha256[:])

	return hashStr
}
