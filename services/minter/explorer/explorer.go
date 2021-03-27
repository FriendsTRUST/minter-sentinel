package explorer

import (
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

func (svc *Service) ListBlocks() (*ListBlocksResponse, error) {
	res, err := svc.http.Get(fmt.Sprintf("%s/blocks", svc.url))

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

	var resp ListBlocksResponse

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if resp.Code != nil && *resp.Code != 200 {
		return &resp, errors.New("response error")
	}

	return &resp, nil
}

func (svc *Service) GetBlock(height int) (*GetBlockResponse, error) {
	res, err := svc.http.Get(fmt.Sprintf("%s/blocks/%d", svc.url, height))

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

	//svc.logger.Println(string(data))

	var resp GetBlockResponse

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if resp.Error != nil && *resp.Error.Code != 200 {
		if *resp.Error.Code == 404 {
			return &resp, NewBlockNotFoundError(&resp)
		}

		return &resp, errors.New("response error")
	}

	return &resp, nil
}
