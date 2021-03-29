package node

import (
	"strings"
	"testing"
)

const publicKey = "Mp61022c1428f17e02e5b3b130564ab3d37d41ad32ba361b5704642f079888c821"
const seed1 = "4518edc842a0edbf1576c69afd04e66649655c166b8805ffca9926eb942c7fc4271f766eac16887a66e302f0daa70df7893bd3fb138eab9042f1ac02d866cf3a"
const seed2 = "db5000a13b73a677d3881543826d527753226dda4f2da50da4808b1f3d7f8f21c41c5407e4a8d969a9600021ea90792c98f08d864da2e023a943260c83fa0cd2"
const address = "Mx4e16a6bfc1bac5f4cf94ef60ab5047510a32abbc"
const multisigAddress = "Mx8b7cd0d453da25954b230de2233605cc35813bd2"

func TestService_Wallet(t *testing.T) {
	svc, _ := New([]string{"https://node-api.testnet.minter.network/v2"}, true, nil)

	wal, err := svc.Wallet("", seed1)

	if err != nil {
		t.Fatalf("failed to create wallet: %s", err)
	}

	if wal.Address != address {
		t.Fatalf("wrong address: expected %s, got %s", address, wal.Address)
	}
}

func TestService_GenerateCandidateOffTransaction_Single(t *testing.T) {
	svc, _ := New([]string{"https://node-api.testnet.minter.network/v2"}, true, nil)

	tx, err := svc.GenerateCandidateOffTransaction(publicKey, address, seed1)

	if err != nil {
		t.Fatalf("failed to generate transaction: %s", err)
	}

	if len(tx) == 0 || !strings.HasPrefix(tx, "0x") {
		t.Fatalf("wrong transaction: %s", tx)
	}
}

func TestService_GenerateCandidateOffTransaction_Multisig(t *testing.T) {
	svc, _ := New([]string{"https://node-api.testnet.minter.network/v2"}, true, nil)

	tx, err := svc.GenerateCandidateOffTransaction(publicKey, multisigAddress, seed1, seed2)

	if err != nil {
		t.Fatalf("failed to generate transaction: %s", err)
	}

	if len(tx) == 0 || !strings.HasPrefix(tx, "0x") {
		t.Fatalf("wrong transaction: %s", tx)
	}
}
