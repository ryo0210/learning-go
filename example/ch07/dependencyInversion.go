package main

import (
	"errors"
	"fmt"
	"net/http"
)

// LogOutputはログを記録する関数
func LogOutput(message string) {
	fmt.Println(message)
}

// SimpleDataStoreは簡単なデータの保存場所
type SimpleDataStore struct {
	userData map[string]string
}

func (sds SimpleDataStore) UserNameForID(userID string) (string, bool) {
	name, ok := sds.userData[userID]
	return name, ok
}

// NewSimpleDataStoreは、SimpleDataStoreのインスタンスを生成するファクトリ関数
func NewSimpleDataStore() SimpleDataStore {
	return SimpleDataStore{
		userData: map[string]string{
			"1": "Fred",
			"2": "Mary",
			"3": "Pat",
		},
	}
}

// DataStoreは、ビジネスロジックが何に依存するかを説明したインターフェイス
type DataStore interface {
	UserNameForID(userID string) (string, bool)
}

// Loggerは、ビジネスロジックが何に依存するかを説明したインターフェイス
type Logger interface {
	Log(message string)
}

// LoggerAdapterは、LogOutputが適合するメソッドを持った関数型
type LoggerAdapter func(message string)

// 関数型にメソッドを定義
func (lg LoggerAdapter) Log(message string) {
	lg(message)
}

// SimpleLogicは、LoggerとDataStoreのフィールドを持った構造体。
// 具象型には触れていないので依存はなく、後になって違うとこらから新たな実装を持ってきて入れ替えても問題ない。
type SimpleLogic struct {
	l  Logger
	ds DataStore
}

func (sl SimpleLogic) SayHello(userID string) (string, error) {
	sl.l.Log("SayHello(" + userID + ")")
	name, ok := sl.ds.UserNameForID(userID)
	if !ok {
		return "", errors.New("不明なユーザー")
	}
	return name + "さん　こんにちは。", nil
}

func (sl SimpleLogic) SayGoodbye(userID string) (string, error) {
	sl.l.Log("SayGoodbye(" + userID + ")")
	name, ok := sl.ds.UserNameForID(userID)
	if !ok {
		return "", errors.New("不明なユーザー")
	}
	return name + "さん　さようなら", nil
}

// NewSimpleLogicは、SimpleLogicのインスタンスを作成するファクトリ関数。インターフェイスを渡すと構造体を返す。
func NewSimpleLogic(l Logger, ds DataStore) SimpleLogic {
	return SimpleLogic{
		l:  l,
		ds: ds,
	}
}

// Logicは、Controllerで「こんにちは」を言うためのインターフェイス
type Logic interface {
	SayHello(userID string) (string, error)
}

// Controllerは、
type Controller struct {
	l     Logger
	logic Logic
}

func (c Controller) SayHello(w http.ResponseWriter, r *http.Request) {
	c.l.Log("SayHello内: ")
	userID := r.URL.Query().Get("user_id")
	message, err := c.logic.SayHello(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte(message))
}

// NewSimpleLogicは、Controllerのインスタンスを作成するファクトリ関数。インターフェイスを渡すと構造体を返す。
func NewController(l Logger, logic Logic) Controller {
	return Controller{
		l:     l,
		logic: logic,
	}
}

// 全てのコンポーネントを結びつけ、サーバーを起動する。
func main() {
	l := LoggerAdapter(LogOutput)
	ds := NewSimpleDataStore()
	logic := NewSimpleLogic(l, ds)
	c := NewController(l, logic)
	http.HandleFunc("/hello", c.SayHello)
	http.ListenAndServe(":8080", nil)
}
