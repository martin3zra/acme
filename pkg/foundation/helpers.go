package foundation

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
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
	result := make(map[string]any)
	val := reflect.ValueOf(obj)
	typ := reflect.TypeOf(obj)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	for i := range val.NumField() {
		field := typ.Field(i)
		if field.IsExported() {
			result[field.Name] = val.Field(i).Interface()
		}
	}

	return result
}

func GeneratePrefixedNumber(prefix string, length, value int) string {
	return fmt.Sprintf(fmt.Sprintf("%s%%0%dv", prefix, length), value)
}

func ToJSON(m any) string {
	js, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}

	return strings.ReplaceAll(string(js), ",", ", ")
}

// ResolvePath returns the absolute path to an asset file based on the mode.
func ResolvePath(relativePath string) string {
	// Get the path of the running binary
	execPath, err := os.Executable()
	if err != nil {
		panic(err) // or handle more gracefully
	}
	baseDir := filepath.Dir(execPath)

	// Join with the relative path to the static file
	return filepath.Join(baseDir, relativePath)
}

func MapSlice[T, U any](s []T, f func(T) U) []U {
	result := make([]U, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}
