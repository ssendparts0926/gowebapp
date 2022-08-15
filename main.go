package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/stretchr/objx"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
)

type templateHandler struct {
	once     sync.Once
	filename string
	tampl    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// sync.onceは常に同じ値を使用するため、レシーバー(t *templateHandler)はポインタ型
	t.once.Do(func() {
		// 解析されたテンプレートへの参照を保持
		// Must関数は解析したときにエラー発生した場合panicする
		t.tampl = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	// valueがinteface型なので、あらゆる型をvalueにできる
	data := map[string]interface{}{
		"Host":   r.Host,
		"String": "string",
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		// objxはマップ, スライス, JSON およびその他のデータを扱うための Go パッケージ
		// UserDataの値はJSON文字列
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	// 解析されたテンプレートをdataに適用し、結果をwに書き込み
	t.tampl.Execute(w, data)
}

func main() {
	var addr = flag.String("addr", "8082", "アプリケーションのアドレス")
	flag.Parse() // フラグを解釈します

	// Gomniauthのセットアップ
	gomniauth.SetSecurityKey("webapp")
	gomniauth.WithProviders(facebook.New("", "", "http://localhost:8082/auth/callback/facebook"), github.New("", "", "http://localhost:8082/auth/callback/github"), google.New("113585110266-kukiaqpkfjj88h1q23701kc80govkdkq.apps.googleusercontent.com", "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", "http://localhost:8082/auth/callback/google"))

	r := newRoom()

	// ルート
	// 第一引数のパスにアクセスすると、第二引数の関数が呼ばれる
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)

	// チャットルームを開始します（バックグラウンド）
	go r.run()

	// webサーバーを起動します（メインスレッド）
	log.Println("webサーバーを起動します。ポート：", *addr)
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
