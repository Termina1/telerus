package main

import "github.com/armon/go-socks5"
import "os"
import "strings"
import "errors"
import "net"
import "strconv"
import "fmt"

type verifier func(string, bool) (string, error)

func getEnvAndVerify(key, fallback string, verify verifier) string {
	value, ok := os.LookupEnv(key)
	result, error := verify(value, ok)

	if error != nil {
		panic(error)
	}
	if ok {
		return result
	}
	return fallback
}

func noverify(value string, ok bool) (string, error) {
	return value, nil
}

func getEnv(key, fallback string) string {
	return getEnvAndVerify(key, fallback, noverify)
}

func getEnvFallbackVerify(value, fallback string, ver verifier) string {
	return getEnvAndVerify(value, fallback, func(value string, ok bool) (string, error) {
		if !ok {
			return value, nil
		}
		return ver(value, ok)
	})
}

func getConfig() *socks5.Config {
	var conf *socks5.Config
	credsLine := getEnv("SOCKS_CREDS", "")

	if len(credsLine) > 0 {
		credsArr := strings.Split(credsLine, ",")
		if len(credsArr) != 2 {
			panic("Credentials should be passed as CREDS='username,password'")
		}
		creds := make(socks5.StaticCredentials)
		creds[credsArr[0]] = credsArr[1]
		conf = &socks5.Config{Credentials: creds}
	} else {
		conf = &socks5.Config{}
	}
	return conf
}

func main() {
	conf := getConfig()
	protocol := getEnvFallbackVerify("SOCKS_PROTOCOL", "tcp", func(value string, ok bool) (string, error) {
		if value != "tcp" && value != "udp" {
			return value, errors.New("Protocol should be tcp or udp")
		}
		return value, nil
	})
	address := getEnvFallbackVerify("SOCKS_BIND_ADDRESS", "0.0.0.0", func(value string, ok bool) (string, error) {
		result := net.ParseIP(value)
		if result == nil {
			return value, errors.New("Not a valid ip")
		}
		return value, nil
	})
	port := getEnvFallbackVerify("SOCKS_BIND_PORT", "8000", func(value string, ok bool) (string, error) {
		port, err := strconv.Atoi(value)
		if err != nil {
			return value, err
		}
		if port < 1 || port > 65535 {
			return value, errors.New("Invalid port number")
		}
		return strconv.Itoa(port), nil
	})

	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening on " + address + ":" + port)
	if err := server.ListenAndServe(protocol, address+":"+port); err != nil {
		panic(err)
	}
}
