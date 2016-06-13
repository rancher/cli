package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rancher/go-rancher/client"
)

func FormatEndpoint(data interface{}) string {
	dataSlice, ok := data.([]interface{})
	if !ok {
		return ""
	}

	buf := &bytes.Buffer{}
	for _, value := range dataSlice {
		dataMap, ok := value.(map[string]interface{})
		if !ok {
			return ""
		}

		s := fmt.Sprintf("%v:%v", dataMap["ipAddress"], dataMap["port"])
		if buf.Len() == 0 {
			buf.WriteString(s)
		} else {
			buf.WriteString(", ")
			buf.WriteString(s)
		}
	}

	return buf.String()
}

func FormatIPAddresses(data interface{}) string {
	ips, ok := data.([]client.IpAddress)
	if !ok {
		return ""
	}

	ipStrings := []string{}
	for _, ip := range ips {
		if ip.Address != "" {
			ipStrings = append(ipStrings, ip.Address)
		}
	}

	return strings.Join(ipStrings, ", ")
}

func FormatJson(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "    ")
	return string(bytes) + "\n", err
}
