package service

// package名はユニークでなくてよい。importする際には、PATHを指定するため、一つのディレクトリに一つのpackageが対応する。
// ディレクトリ名とpackage名は一致する必要はないが、一致してるほうが分かりやすい。
// package名は単数形で一単語が望ましい。複数の単語を組み合わせる場合は、スネークケースを使う。

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Golangにはclassがない。関連する変数やメソッドをまとめるために、structを使う。
type Auth struct{}

// メソッド名の頭文字は大文字にすると、publicになる。
// メソッド名の前にある()内のstructは、レシーバと言い、structにメソッドを追加するために使う。
// レシーバには値レシーバとポインタレシーバがある。値レシーバは、structのコピーを作成してメソッドを呼び出す。ポインタレシーバは、structのポインタを作成してメソッドを呼び出す。
// 値レシーバはメンバ変数を変更できないのでGetter向け。ポインタレシーバはメンバ変数を変更できるのでSetter向け。メンバ変数を参照しない場合、型の指定だけでもコンパイルエラーにならない。
func (Auth) GetConnect() *oauth2.Config {
	// Googleなら、google.ConfigFromJSON()を使うと、Webコンソールから認証設定時に最初にダウンロードできるjsonファイルの内容(とScope)からconfigを生成できる。
	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLECLIENTID"), // 環境変数の取得
		ClientSecret: os.Getenv("GOOGLECLIENTSECRET"),
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		Endpoint:     google.Endpoint,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
	}

	return config
}
