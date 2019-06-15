package main

import (
	"net/http"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

type (
	Socket struct {
		Connection *websocket.Conn
		TypeHandlers map[string]func(SocketContext)SocketResponse
		DisconnectHandler func(SocketContext)
	}

	SocketContext struct {
		Connection *websocket.Conn
		Type string
		Data map[string]interface{}
		Prop map[string]interface{}
	}

	SocketResponse struct {
		Type        string      `json:"type"`
		Debug       bool        `json:"debug"`
		Update      bool        `json:"update"`
		DestroyData DestroyData `json:"destroyData,omitempty"`
		JoinData    JoinData    `json:"joinData,omitempty"`
		LeaveData   LeaveData   `json:"leaveData,omitempty"`
		StartData   StartData   `json:"startData,omitempty"`
		StopData    StopData    `json:"stopData,omitempty"`
	}

	Game struct {
		ID string `json:"id"`
		Code string `json:"code"`
		Location string `json:"location"`
		Players int `json:"players"`
		Active bool `json:"active"`
	}

	Player struct {
		ID string `json:"id"`
		Username string `json:"username"`
		Role string `json:"role"`
		Spy bool `json:"spy"`
	}

	DestroyData struct {
		Sucess bool `json:"sucess"`
		Game Game `json:"game"`
	}

	JoinData struct {
		Sucess bool `json"sucess"`
		Player Player `json:"player"`
		Game Game `json:"game"`
	}

	ResponseUpdate struct {
		GameOver bool `json:"gameOver,omitempty"`
		PlayerLeft bool `json:"playerLeft,omitempty"`
		Username
		Message string `json:"message`
	}

	ResponseError struct {
		Code string `json:"code,omitempty"`
		Desc string `json:"description,omitempty"`
	}
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

)

func newSocketRouter(w http.ResponseWriter, r *http.Request) (Socket, error) {
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return Socket{}, err
	}

	return Socket{
		Connection: connection,
		TypeHandlers: make(map[string]func(SocketContext)SocketResponse),
	}, nil
}

func (s *Socket) handleRoutes() {
	conn := s.Connection

	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			if s.DisconnectHandler != nil {
				s.DisconnectHandler(SocketContext{ Connection: conn, })
			}
			conn.Close()
			return
		}
		
		if msgType == websocket.TextMessage {
			msgType := gjson.Get(string(data), "type").String()
			m, ok := gjson.Parse(string(gjson.Get(string(data), "data").String())).Value().(map[string]interface{})
			if ok {
				s.Connection.WriteJSON(s.TypeHandlers[msgType](SocketContext{
					Connection: conn,
					Type: msgType,
					Data: m,
					Prop: make(map[string]interface{}),
				}))
			}
		}
	}
}

func (s *Socket) addRoute(msgType string, handler func(SocketContext)SocketResponse) {
	s.TypeHandlers[msgType] = handler
}

func (s *Socket) addDisconnect(handler func(SocketContext)) {
	s.DisconnectHandler = handler
}

func (ctx *SocketContext) reply(msg interface{}) {
	ctx.Connection.WriteJSON(msg)
}