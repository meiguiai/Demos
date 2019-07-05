package main

import (
	"./impl"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

/*
 //HTTP协议

func wsHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("HelloWorld"))
}
func main() {
	http.HandleFunc("/ws",wsHandler)
	http.ListenAndServe("0.0.0.0:7777",nil)
}
*/

/*
//Socket简单实现
var (
  upgrade = websocket.Upgrader {
  	//允许跨域
  	CheckOrigin: func(r *http.Request) bool {
		return  true
	},
  }	
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var(
		conn *websocket.Conn
		err error
		data []byte
	)
	// Upgrade: websocket
	if conn, err = upgrade.Upgrade(w,r,nil); err != nil {
		return
	}
	
	for {
		if _,data, err = conn.ReadMessage(); err != nil {
			goto ERR
		}
		
		if err = conn.WriteMessage(websocket.TextMessage,data); err != nil {
			goto ERR
		}
}
	
ERR:
	//TODO：关闭连接操作
	conn.Close()
}

func main() {
	http.HandleFunc("/ws",wsHandler)
	http.ListenAndServe("0.0.0.0:7777",nil)
}
*/

//封装Connection实现
var (
	upgrader = websocket.Upgrader {
		//允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return  true
		},
	}
	err error
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var(
		wsConn *websocket.Conn
		data []byte
		conn * impl.Connection
	)
	// Upgrade: websocket
	if wsConn, err = upgrader.Upgrade(w,r,nil); err != nil {
		return
	}

	if conn, err = impl.InitConnection(wsConn); err != nil {
		goto ERR
	}
	
	//为了演示线程安全
	go func() {
		var (
			err error
		)
		for  {
			if err = conn.WriteMessage([]byte("heartbeat")); err != nil {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()
	
	for {
		if data, err = conn.ReadMessage(); err != nil {
			goto ERR
		}
		if err = conn.WriteMessage(data); err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close()
}

func main() {
	http.HandleFunc("/ws",wsHandler)
	if err = http.ListenAndServe("0.0.0.0:7777",nil); err != nil {
		fmt.Println(err)
	}
}
