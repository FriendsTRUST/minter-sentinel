package minter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/MinterTeam/minter-go-sdk/v2/transaction"
	"github.com/MinterTeam/minter-go-sdk/v2/wallet"
	"github.com/sirupsen/logrus"
)

type Service struct {
	apiUrl             string
	currentApiUrlIndex int
	testnet            bool
	logger             *logrus.Logger
	httpClient         *http.Client
}

func New(apiUrl string, testnet bool, logger *logrus.Logger) (*Service, error) {
	var netTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	s := &Service{
		apiUrl:  apiUrl,
		logger:  logger,
		testnet: testnet,
		httpClient: &http.Client{
			Timeout:   time.Second * 10,
			Transport: netTransport,
		},
	}

	return s, nil
}

func (s *Service) Wallet(seed string) (*wallet.Wallet, error) {
	return wallet.Create(seed, "")
}

func (s *Service) GenerateCandidateOffTransaction(publicKey string, wallet *wallet.Wallet, privateKey ...string) (string, error) {
	var chainID transaction.ChainID
	if s.testnet {
		chainID = transaction.TestNetChainID
	} else {
		chainID = transaction.MainNetChainID
	}

	tx, err := transaction.NewBuilder(chainID).NewTransaction(
		transaction.NewSetCandidateOffData().MustSetPubKey(publicKey),
	)

	if err != nil {
		return "", err
	}

	if len(privateKey) == 1 {
		tx = tx.SetSignatureType(transaction.SignatureTypeSingle)
	} else {
		tx = tx.SetSignatureType(transaction.SignatureTypeMulti)
	}

	getAddress, err := s.getAddress(wallet.Address)

	if err != nil {
		return "", err
	}

	signedTx, err := tx.
		SetNonce(getAddress.TransactionCount + 1).
		SetGasPrice(1).
		SetGasCoin(0).
		Sign(strings.Join(privateKey, ""))

	if err != nil {
		return "", err
	}

	return signedTx.Encode()
}

func (s *Service) getAddress(address string) (*GetAddressResponse, error) {
	res, err := s.httpClient.Get(fmt.Sprintf("%s/address/%s", s.apiUrl, address))

	//goland:noinspection GoDeferInLoop
	defer func() {
		if err := res.Body.Close(); err != nil {
			s.logger.Error(err)
		}
	}()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	var resp GetAddressResponse

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	if resp.Error != nil && resp.Error.Code != 200 {
		return &resp, errors.New("response error")
	}

	return &resp, nil
}
