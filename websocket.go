package echoserver

import (
	"fmt"
	"github.com/aliworkshop/handlerlib"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type echoWebSocket struct {
	conn *websocket.Conn
}

func (gws *echoWebSocket) Read() (int, []byte, error) {
	return gws.conn.ReadMessage()
}

func (gws *echoWebSocket) Write(mType int, msg []byte) error {
	return gws.conn.WriteMessage(mType, msg)
}

func (gws *echoWebSocket) WriteJson(msg interface{}) error {
	return gws.conn.WriteJSON(msg)
}

func (gws *echoWebSocket) Close() {
	err := gws.conn.Close()
	if err != nil {
		fmt.Println("in close handler", err)
	}
}
func (gws *echoWebSocket) SetReadDeadLine(deadline time.Duration) {
	gws.conn.SetReadDeadline(time.Now().Add(deadline))
}
func (gws *echoWebSocket) SetWriteDeadLine(deadline time.Duration) {
	gws.conn.SetWriteDeadline(time.Now().Add(deadline))
}
func (gws *echoWebSocket) WriteControl(mType int, msg []byte, deadline time.Time) error {
	return gws.conn.WriteControl(mType, msg, deadline)
}
func (gws *echoWebSocket) SetPingHandler(f func(string) error) error {
	gws.conn.SetPingHandler(f)
	return nil
}
func (gws *echoWebSocket) SetCloseHandler(f func(code int, text string) error) error {
	gws.conn.SetCloseHandler(f)
	return nil
}

func upgrade(c echo.Context) (handlerlib.WebSocketModel, error) {
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
