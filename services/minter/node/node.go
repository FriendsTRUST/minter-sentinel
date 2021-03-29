package node

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/MinterTeam/minter-go-sdk/v2/transaction"
	"github.com/MinterTeam/minter-go-sdk/v2/wallet"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

const (
	status          = "/status"
	candidate       = "/candidate/{candidate}"
	getBlock        = "/block/{height}"
	getAddress      = "/address/{address}"
	sendTransaction = "/send_transaction"
)

type Service struct {
	nodeApis            []string
	currentNodeApiIndex int
	testnet             bool
	logger              *logrus.Logger
	http                *resty.Client
}

func New(nodeApis []string, testnet bool, logger *logrus.Logger) (*Service, error) {
	http := resty.New().
		SetRetryCount(1).
		SetHostURL(nodeApis[0])

	s := &Service{
		nodeApis:            nodeApis,
		currentNodeApiIndex: 0,
		logger:              logger,
		testnet:             testnet,
		http:                http,
	}

	return s, nil
}

func (svc *Service) Ping() error {
	for _, url := range svc.nodeApis {
		svc.http.SetHostURL(url)
		if v, err := svc.Status(); err != nil {
			return err
		} else if v.CatchingUp {
			return errors.New(fmt.Sprintf("node %s is catching up", url))
		}
	}

	return nil
}

func (svc *Service) Status() (*StatusResponse, error) {
	var res StatusResponse

	_, err := svc.http.R().
		SetResult(&res).
		Get(status)

	return &res, err
}

func (svc *Service) GetCandidate(publicKey string) (*CandidateResponse, error) {
	var res CandidateResponse

	r, err := svc.http.R().
		SetPathParam("candidate", publicKey).
		SetResult(&res).
		Get(candidate)

	if err != nil {
		return nil, err
	}

	if r.StatusCode() != 200 {
		return &res, &CandidateNotFound{}
	}

	return &res, err
}

func (svc *Service) GetBlock(height int) (*GetBlockResponse, error) {
	var resp *resty.Response
	var err error
	var res GetBlockResponse

	svc.try(func() error {
		resp, err = svc.http.R().
			SetPathParam("height", strconv.Itoa(height)).
			SetResult(&res).
			SetError(&res).
			Get(getBlock)

		if err != nil {
			return err
		}

		if resp.StatusCode() == 404 {
			return nil
		}

		return err
	})

	if resp.StatusCode() == 404 {
		return &res, NewBlockNotFoundError(&res)
	}

	return &res, err
}

func (svc *Service) Wallet(mnemonic string, seed string) (*wallet.Wallet, error) {
	return wallet.Create(mnemonic, seed)
}

func (svc *Service) GenerateCandidateOffTransaction(publicKey string, walletAddress string, seeds ...string) (string, error) {
	var chainID transaction.ChainID
	if svc.testnet {
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

	getAddress, err := svc.getAddress(walletAddress)

	if err != nil {
		return "", err
	}

	tx = tx.
		SetNonce(getAddress.TransactionCount + 1).
		SetGasPrice(1).
		SetGasCoin(0)

	var signed transaction.Signed

	if len(seeds) == 1 {
		wal, err := svc.Wallet("", seeds[0])

		if err != nil {
			return "", err
		}

		s, err := tx.SetSignatureType(transaction.SignatureTypeSingle).Sign(wal.PrivateKey)

		if err != nil {
			return "", err
		}

		signed = s
	} else {
		var privateKeys []string

		for _, seed := range seeds {
			wal, err := svc.Wallet("", seed)

			if err != nil {
				return "", err
			}

			privateKeys = append(privateKeys, wal.PrivateKey)
		}

		s, err := tx.SetSignatureType(transaction.SignatureTypeMulti).Sign(walletAddress, privateKeys...)

		if err != nil {
			return "", err
		}

		signed = s
	}

	return signed.Encode()
}

func (svc *Service) SendTransaction(tx string) (*SendTransactionResponse, error) {
	var res SendTransactionResponse
	var err error

	svc.try(func() error {
		_, err := svc.http.R().
			SetBody(&SendTransactionRequest{Tx: tx}).
			SetResult(&res).
			SetError(&res).
			Post(sendTransaction)

		return err
	})

	return &res, err
}

func (svc *Service) getAddress(address string) (*GetAddressResponse, error) {
	var res GetAddressResponse

	_, err := svc.http.R().
		SetPathParam("address", address).
		SetResult(&res).
		Get(getAddress)

	if err != nil {
		return &res, err
	}

	return &res, nil
}

func (svc *Service) try(callback func() error) {
	lastIndex := len(svc.nodeApis) - 1

	for i, url := range svc.nodeApis {
		svc.http.SetHostURL(url)

		err := callback()

		if err != nil {
			if i == lastIndex {
				return
			}

			continue
		}

		return
	}

	return
}
