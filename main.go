package main

import (
	"blockRetreiver/contracts"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"strconv"
	"strings"

)

type LogRecord struct {
	from       common.Address // user address
	token      common.Address
	actionName string // {"withdraw", "supply", "borrow", "payback"}
	amount     uint
}

func strToInt(s string) int64 {
	converted, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return converted
}

func getLastMined(client *ethclient.Client) int64 {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(header.Number.String()) // 5671744
	return strToInt(header.Number.String())
}

func crawl(contractAbi abi.ABI, logs []types.Log) []LogRecord {
	var records []LogRecord
	logDepositSig := []byte("Deposit(address,address,address,uint256,uint16)")
	logDepositSigHash := crypto.Keccak256Hash(logDepositSig)
	logWithdrawSig := []byte("Withdraw(address,address,address,uint256)")
	logWithdrawSigHash := crypto.Keccak256Hash(logWithdrawSig)
	logBorrowSig := []byte("Borrow(address,address,address,uint256,uint256,uint256,uint16)")
	logBorrowSigHash := crypto.Keccak256Hash(logBorrowSig)
	logRepaySig := []byte("Repay(address,address,address,uint256)")
	logRepaySigHash := crypto.Keccak256Hash(logRepaySig)

	for _, vLog := range logs {
		switch vLog.Topics[0].Hex() {
		case logDepositSigHash.Hex():
			data, err := contractAbi.Unpack("Deposit", vLog.Data)
			if err != nil {
				log.Fatal(err)
			}
			var logRecord LogRecord
			a := data[1].(*big.Int)
			logRecord.actionName = "deposit"
			logRecord.from = data[0].(common.Address)
			logRecord.amount = uint(a.Uint64())
			logRecord.token = common.HexToAddress(vLog.Topics[1].Hex())
			records = append(records, logRecord)

		case logBorrowSigHash.Hex():
			var logRecord LogRecord
			data, err := contractAbi.Unpack("Borrow", vLog.Data)
			if err != nil {
				log.Fatal(err)
			}
			a := data[1].(*big.Int)
			logRecord.actionName = "borrow"
			logRecord.from = data[0].(common.Address)
			logRecord.amount = uint(a.Uint64())
			logRecord.token = common.HexToAddress(vLog.Topics[1].Hex())
			records = append(records, logRecord)

		case logWithdrawSigHash.Hex():
			data, err := contractAbi.Unpack("withdraw", vLog.Data)
			if err != nil {
				log.Fatal(err)
			}
			var logRecord LogRecord
			a := data[0].(*big.Int)
			logRecord.actionName = "withdraw"
			logRecord.from = common.HexToAddress(a.String())
			logRecord.amount = uint(a.Uint64())
			logRecord.token = common.HexToAddress(vLog.Topics[1].Hex())
			records = append(records, logRecord)

		case logRepaySigHash.Hex():
			data, err := contractAbi.Unpack("Repay", vLog.Data)
			if err != nil {
				log.Fatal(err)
			}
			var logRecord LogRecord
			b := data[0].(*big.Int)
			logRecord.actionName = "repay"
			logRecord.from = common.HexToAddress(vLog.Topics[2].Hex())
			logRecord.amount = uint(b.Uint64())
			logRecord.token = common.HexToAddress(vLog.Topics[1].Hex())
			records = append(records, logRecord)
		}
	}
	return records
}
func main() {

	client, err := ethclient.Dial("https://mainnet.infura.io/v3/854bb57ac48b4b649b6a653f103e09ab")
	if err != nil {
		log.Fatal(err)
	}

	contractAddress := common.HexToAddress("0x7d2768dE32b0b80b7a3454c06BdAc94A69DDc7A9")

	lastMined := getLastMined(client)
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(lastMined - 5000),
		ToBlock:   big.NewInt(lastMined),
		Addresses: []common.Address{
			contractAddress,
		},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}

	contractAbi, err := abi.JSON(strings.NewReader(string(contracts.AaveABI)))

	if err != nil {
		log.Fatal(err)
	}
	records := crawl(contractAbi, logs)
	fmt.Println(len(records))

}
