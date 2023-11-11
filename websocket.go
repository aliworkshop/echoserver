package echoserver

import (
	"fmt"
	"github.com/aliworkshop/gateway/v2"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type echoWebSocket struct {
	conn *websocket.Conn
}

func (ew *echoWebSocket) Read() (int, []byte, error) {
	return ew.conn.ReadMessage()
}

func (ew *echoWebSocket) Write(mType int, msg []byte) error {
	return ew.conn.WriteMessage(mType, msg)
}

func (ew *echoWebSocket) WriteJson(msg interface{}) error {
	return ew.conn.WriteJSON(msg)
}

func (ew *echoWebSocket) Close() {
	err := ew.conn.Close()
	if err != nil {
		fmt.Println("in close handler", err)
	}
}
func (ew *echoWebSocket) SetReadDeadLine(deadline time.Duration) {
	ew.conn.SetReadDeadline(time.Now().Add(deadline))
}
func (ew *echoWebSocket) SetWriteDeadLine(deadline time.Duration) {
	ew.conn.SetWriteDeadline(time.Now().Add(deadline))
}
func (ew *echoWebSocket) WriteControl(mType int, msg []byte, deadline time.Time) error {
	return ew.conn.WriteControl(mType, msg, deadline)
}
func (ew *echoWebSocket) SetPingHandler(f func(string) error) error {
	ew.conn.SetPingHandler(f)
	return nil
}
func (ew *echoWebSocket) SetCloseHandler(f func(code int, text string) error) error {
	ew.conn.SetCloseHandler(f)
	return nil
}

func upgrade(c echo.Context) (gateway.WebSocketHandler, error) {
	gs := new(echoWebSocket)
	upper := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upper.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Println("connection initialization error", err)
		return nil, err
	}
	gs.conn = conn
	return gs, nil
}
