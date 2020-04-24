package byteutil

import (
	"fmt"
	"testing"
)

func TestBytesToHexString(t *testing.T) {
	heartBeat := []byte{0x00, 0x03, 0x00, 0x77, 0x00, 0x00, 0xf4, 0x01}
	fmt.Println(BytesToHexString(heartBeat, ""))
}