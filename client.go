package main

import (
	"time"

	"github.com/gorilla/websocket"
)

// clientはチャットを行っている１人のユーザーを表します。
type client struct {
	// socketはこのクライアントのためのwebSocketです。
	socket *websocket.Conn
	// sendはメッセージが送られるチャネル
	send chan *message
	// roomはこのクライアントが参加しているチャットルームです。
	room *room
	// userDataはユーザに関する情報を保持します。
	userData map[string]interface{}
}

// WebSocketへの読み書きを行うメソッド
func (c *client) read() {
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil {
			msg.When = time.Now()
			msg.Name = c.userData["name"].(string)
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
