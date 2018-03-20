package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

type wsResponse struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

// resolveWebsocketURL accepts the action link for a websocket connection and a
// payload and returns the full ws URL
func resolveWebsocketURL(
	ctx *cli.Context,
	URL string,
	payload []byte,
) (string, error) {
	config, err := lookupConfig(ctx)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(payload))
	if nil != err {
		return "", fmt.Errorf("request error:%v", err)
	}

	req.SetBasicAuth(config.AccessKey, config.SecretKey)

	res, err := http.DefaultClient.Do(req)
	if nil != err {
		return "", err
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if nil != err {
		return "", err
	}

	res.Body.Close()

	logrus.Debugf("websocket response: %s", string(bodyBytes))

	var ws wsResponse
	err = json.Unmarshal(bodyBytes, &ws)
	if nil != err {
		return "", err
	}

	wsURL := ws.URL + "?token=" + ws.Token
	return wsURL, nil
}
