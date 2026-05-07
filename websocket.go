package echoserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aliworkshop/gateway/v2"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type echoWebSocket struct {
	conn *websocket.Conn
}

func (ew *echoWebSocket) Read(_ context.Context) (int, []byte, error) {
	return ew.conn.ReadMessage()
}

func (ew *echoWebSocket) Write(_ context.Context, mType int, msg []byte) error {
	return ew.conn.WriteMessage(mType, msg)
}

func (ew *echoWebSocket) WriteJson(_ context.Context, msg any) error {
	return ew.conn.WriteJSON(msg)
}

func (ew *echoWebSocket) Close() {
	if err := ew.conn.Close(); err != nil {
		fmt.Println("in close handler", err)
	}
}

func (ew *echoWebSocket) SetReadDeadLine(deadline time.Duration) error {
	return ew.conn.SetReadDeadline(time.Now().Add(deadline))
}

func (ew *echoWebSocket) SetWriteDeadLine(deadline time.Duration) error {
	return ew.conn.SetWriteDeadline(time.Now().Add(deadline))
}

func (ew *echoWebSocket) Ping(_ context.Context) error {
	return ew.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
}

func upgrade(c echo.Context) (gateway.WebSocketHandler, error) {
	upper := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upper.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Println("connection initialization error", err)
		return nil, err
	}
	return &echoWebSocket{conn: conn}, nil
}
