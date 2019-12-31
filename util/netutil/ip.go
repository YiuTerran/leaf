package netutil

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
)

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func IsInternetOK() (ok bool) {
	_, err := http.Get("http://www.google.cn/generate_204")
	if err != nil {
		return false
	}
	return true
}

func GetAllIP() []net.IP {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	result := make([]net.IP, 0, 1)
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip, ok := addr.(*net.IPNet)
			if ok && !ip.IP.IsLoopback() && ip.IP.To4() != nil {
				result = append(result, ip.IP)
			}
		}
	}
	return result
}

func IsPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}

func InetAton(ip net.IP) int64 {
	ipv4Int := big.NewInt(0)
	ipv4Int.SetBytes(ip.To4())
	return ipv4Int.Int64()
}

func InetNtoa(ipnr int64) net.IP {
	var bs [4]byte
	bs[0] = byte(ipnr & 0xFF)
	bs[1] = byte((ipnr >> 8) & 0xFF)
	bs[2] = byte((ipnr >> 16) & 0xFF)
	bs[3] = byte((ipnr >> 24) & 0xFF)
	return net.IPv4(bs[3], bs[2], bs[1], bs[0])
}

func IpBetween(from net.IP, to net.IP, test net.IP) bool {
	if from == nil || to == nil || test == nil {
		return false
	}

	from16 := from.To16()
	to16 := to.To16()
	test16 := test.To16()
	if from16 == nil || to16 == nil || test16 == nil {
		return false
	}

	if bytes.Compare(test16, from16) >= 0 && bytes.Compare(test16, to16) <= 0 {
		return true
	}
	return false
}

type IPInfo struct {
	Code int `json:"code"`
	Data IP  `json:"data"`
}

type IP struct {
	Country   string `json:"country"`
	CountryId string `json:"country_id"`
	Area      string `json:"area"`
	AreaId    string `json:"area_id"`
	Region    string `json:"region"`
	RegionId  string `json:"region_id"`
	City      string `json:"city"`
	CityId    string `json:"city_id"`
	Isp       string `json:"isp"`
}

//获取公网IP
func GetExternalIP() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return string(content)
}

func GetIPInfo(ip string) *IPInfo {
	url := "http://ip.taobao.com/service/getIpInfo.php?ip="
	url += ip
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var result IPInfo
	if err := json.Unmarshal(out, &result); err != nil {
		return nil
	}
	return &result
}
