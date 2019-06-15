package main

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

type (
	Socket struct {
		Connection        *websocket.Conn
		TypeHandlers      map[string]func(SocketContext) SocketResponse
		DisconnectHandler func(SocketContext)
	}

	SocketContext struct {
		Connection *websocket.Conn
		Type       string
		Data       map[string]interface{}
		Prop       map[string]interface{}
	}

	SocketResponse struct {
		Type          string         `json:"type"`
		BroadcastData *BroadcastData `json:"data,omitempty"`
		ResponseData  *ResponseData  `json:"data,omitempty"`
		Error         *ResponseError `json:"error,omitempty"`
	}

	BroadcastData struct {
		Code      string    `json:"code,omitempty"`
		Players   []string  `json:"players,omitempty"`
		Locations []string  `json:"locations,omitempty"`
		Time      time.Time `json:"startTime,omitempty"`
	}

	ResponseData struct {
		Code     string `json:"code,omitempty"`
		Username string `json:"username,omitempty"`
		Spy      bool   `json:"spy,omitempty"`
		Location string `json:"location,omitempty"`
		Role     string `json:"role,omitempty"`
	}

	ResponseError struct {
		Code string `json:"code"`
		Desc string `json:"description"`
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
		Connection:   connection,
		TypeHandlers: make(map[string]func(SocketContext) SocketResponse),
	}, nil
}

func (s *Socket) handleRoutes() {
	conn := s.Connection

	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			if s.DisconnectHandler != nil {
				s.DisconnectHandler(SocketContext{Connection: conn})
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
					Type:       msgType,
					Data:       m,
					Prop:       make(map[string]interface{}),
				}))
			}
		}
	}
}

func (s *Socket) addRoute(msgType string, handler func(SocketContext) SocketResponse) {
	s.TypeHandlers[msgType] = handler
}

func (s *Socket) addDisconnect(handler func(SocketContext)) {
	s.DisconnectHandler = handler
}

func (ctx *SocketContext) reply(msg interface{}) {
	ctx.Connection.WriteJSON(msg)
}
