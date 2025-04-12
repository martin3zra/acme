package foundation

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
)

func GetIpAddress(r *http.Request) (string, error) {
	ips := r.Header.Get("X-Forwarded-For")
	splitIps := strings.Split(ips, ",")
	if len(splitIps) > 0 {
		// get last IP in list since ELB prepends other user defined IPs, meaning the last one is the actual client IP.
		netIP := net.ParseIP(splitIps[len(splitIps)-1])
		if netIP != nil {
			return netIP.String(), nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	netIP := net.ParseIP(ip)
	if netIP != nil {
		ip := netIP.String()
		if ip == "::1" {
			return "127.0.0.1", nil
		}
		return ip, nil
	}

	return "", errors.New("IP not found")
}

func AsMap(obj any) map[string]any {
	var result = make(map[string]any)
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		log.Fatal(err)
	}

	return result
}
