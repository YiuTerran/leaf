package tcp

//parser的接口，开放给外部自定义

type IParser interface {
	Read(conn *Conn) ([]byte, error)
	Write(conn *Conn, args ...[]byte) error
}
