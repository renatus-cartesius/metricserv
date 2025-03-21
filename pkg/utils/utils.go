// Package utils providing some useful functions
package utils

import (
	"io"
	"math"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/renatus-cartesius/metricserv/pkg/logger"
	"go.uber.org/zap"
)

func SafeClose(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		logger.Log.Fatal(
			"error on closing",
			zap.String("closer", reflect.ValueOf(closer).String()),
		)
	}
}

func ParseExpNotation(value string) (int64, error) {
	if strings.Contains(value, "e") {
		notation := strings.Split(value, "e+")
		val, err := strconv.ParseFloat(notation[0], 64)
		if err != nil {
			return 0, err
		}
		man, err := strconv.ParseFloat(notation[1], 64)
		if err != nil {
			return 0, err
		}
		return int64(val * math.Pow(10, man)), nil
	}
	res, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func TagHelper(tag string) string {
	if tag == "" {
		return "N/A"
	} else {
		return tag
	}
}

func GetOutgoingIPByURL(rawURL string) (net.IP, error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", url.Host)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.TCPAddr)

	return localAddr.IP, nil
}
