package gate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Service struct {
	url    string
	http   *http.Client
	logger *logrus.Logger
}

func New(url string, logger *logrus.Logger) (*Service, error) {
	var netTransport = &http.Transport{
		DialContext: (&net.Dialer{
			KeepAlive: 5 * time.Second,
			Timeout:   5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	return &Service{
		url: url,
		http: &http.Client{
			Timeout:   time.Second * 10,
			Transport: netTransport,
		},
		logger: logger,
	}, nil
}

func (svc *Service) SendTransaction(tx string) (*SendTransactionResponse, error) {
	body, err := json.Marshal(SendTransactionRequest{Tx: tx})

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/send_transaction", svc.url), bytes.NewBuffer(body))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := svc.http.Do(req)

	defer func() {
		if err := res.Body.Close(); err != nil {
			svc.logger.Error(err)
		}
	}()

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	var resp SendTransactionResponse

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if resp.Error != nil && resp.Error.Code != 200 {
		return &resp, errors.New("response error")
	}

	return &resp, nil
}
