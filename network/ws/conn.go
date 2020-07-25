package ws

import (
	"errors"
	"net"
	"sync"

	"github.com/YiuTerran/leaf/log"
	"github.com/gorilla/websocket"
)

type WebsocketConnSet map[*websocket.Conn]struct{}

type Conn struct {
	sync.Mutex
	conn           *websocket.Conn
	writeChan      chan []byte
	maxMsgLen      uint32
	closeFlag      bool
	remoteOriginIP net.Addr
}

func newWSConn(conn *websocket.Conn, pendingWriteNum int, maxMsgLen uint32, textFormat bool) *Conn {
	wsConn := new(Conn)
	wsConn.conn = conn
	wsConn.writeChan = make(chan []byte, pendingWriteNum)
	wsConn.maxMsgLen = maxMsgLen
	msgType := websocket.BinaryMessage
	if textFormat {
		msgType = websocket.TextMessage
	}
	go func() {
		for b := range wsConn.writeChan {
			if b == nil {
				break
			}

			err := conn.WriteMessage(msgType, b)
			if err != nil {
				break
			}
		}

		_ = conn.Close()
		wsConn.Lock()
		wsConn.closeFlag = true
		wsConn.Unlock()
	}()

	return wsConn
}

func (wsConn *Conn) SetOriginIP(ip net.Addr) {
	wsConn.remoteOriginIP = ip
}

func (wsConn *Conn) doDestroy() {
	_ = wsConn.conn.UnderlyingConn().(*net.TCPConn).SetLinger(0)
	_ = wsConn.conn.Close()

	if !wsConn.closeFlag {
		close(wsConn.writeChan)
		wsConn.closeFlag = true
	}
}

func (wsConn *Conn) Destroy() {
	wsConn.Lock()
	defer wsConn.Unlock()

	wsConn.doDestroy()
}

func (wsConn *Conn) Close() {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return
	}

	wsConn.doWrite(nil)
	wsConn.closeFlag = true
}

func (wsConn *Conn) doWrite(b []byte) {
	if len(wsConn.writeChan) == cap(wsConn.writeChan) {
		log.Debug("close conn: channel full")
		wsConn.doDestroy()
		return
	}

	wsConn.writeChan <- b
}

func (wsConn *Conn) LocalAddr() net.Addr {
	return wsConn.conn.LocalAddr()
}

func (wsConn *Conn) RemoteAddr() net.Addr {
	if wsConn.remoteOriginIP != nil {
		return wsConn.remoteOriginIP
	}
	return wsConn.conn.RemoteAddr()
}

// goroutine not safe
func (wsConn *Conn) ReadMsg() ([]byte, error) {
	_, b, err := wsConn.conn.ReadMessage()
	return b, err
}

// args must not be modified by the others goroutines
func (wsConn *Conn) WriteMsg(args ...[]byte) error {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return nil
	}

	// get len
	var msgLen uint32
	for i := 0; i < len(args); i++ {
		msgLen += uint32(len(args[i]))
	}

	// check len
	if msgLen > wsConn.maxMsgLen {
		return errors.New("message too long")
	} else if msgLen < 1 {
		return errors.New("message too short")
	}

	// don't copy
	if len(args) == 1 {
		wsConn.doWrite(args[0])
		return nil
	}

	// merge the args
	msg := make([]byte, msgLen)
	l := 0
	for i := 0; i < len(args); i++ {
		copy(msg[l:], args[i])
		l += len(args[i])
	}

	wsConn.doWrite(msg)

	return nil
}
