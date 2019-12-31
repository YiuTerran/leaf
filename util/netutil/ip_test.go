package netutil

import (
	"fmt"
	"testing"
)

func TestGetAllIP(t *testing.T) {
	for _, ip := range GetAllIP() {
		fmt.Println(ip.String())
	}
}

func TestGetOutboundIP(t *testing.T) {
	fmt.Println(GetOutboundIP())
}

func TestIsInternetOK(t *testing.T) {
	fmt.Println(IsInternetOK())
}

func TestGetExternalIP(t *testing.T) {
	fmt.Println(GetExternalIP())
}

func TestGetIPInfo(t *testing.T) {
	fmt.Println(GetIPInfo("64.64.225.85"))
}