package console

import (
	"bufio"
	"math"
	"strconv"
	"strings"

	"github.com/YiuTerran/leaf/network"
	"github.com/YiuTerran/leaf/network/tcp"
)

const (
	consolePrompt = "leaf#"
)

var server *tcp.Server

func Init(consolePort int) {
	if consolePort == 0 {
		return
	}
	server = new(tcp.Server)
	server.Addr = "localhost:" + strconv.Itoa(consolePort)
	server.MaxConnNum = math.MaxInt32
	server.PendingWriteNum = 100
	server.NewAgent = newAgent

	server.Start()
}

func Destroy() {
	if server != nil {
		server.Close()
		server = nil
	}
}

type Agent struct {
	conn   *tcp.Conn
	reader *bufio.Reader
}

func newAgent(conn *tcp.Conn) network.Agent {
	a := new(Agent)
	a.conn = conn
	a.reader = bufio.NewReader(conn)
	return a
}

func (a *Agent) Run() {
	for {
		a.conn.Write([]byte(consolePrompt))
		line, err := a.reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSuffix(line[:len(line)-1], "\r")

		args := strings.Fields(line)
		if len(args) == 0 {
			continue
		}
		if args[0] == "quit" {
			break
		}
		var c Command
		for _, _c := range commands {
			if _c.name() == args[0] {
				c = _c
				break
			}
		}
		if c == nil {
			a.conn.Write([]byte("command not found, try `help` for help\r\n"))
			continue
		}
		output := c.run(args[1:])
		if output != "" {
			a.conn.Write([]byte(output + "\r\n"))
		}
	}
}

func (a *Agent) OnClose() {}
