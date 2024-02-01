package blockbench

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func transferEth(transfer_num int) {
	// 连接到节点
	client, err := ethclient.Dial("http://192.168.10.21:8545")
	if err != nil {
		log.Fatal(err)
	}
	// for i := 0; i < len(utils.Addresses); i++ {
	done := make(chan bool)
	go func() {
		for i := 0; i < transfer_num; i++ {
			privateKey, err := crypto.HexToECDSA("5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a")
			if err != nil {
				log.Fatal(err)
			}
			publicKey := privateKey.Public()
			publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
			if !ok {
				log.Fatal("error casting public key to ECDSA")
			}
			fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
			receiverAddress := common.HexToAddress(Addresses[i])
			nonce, err := client.NonceAt(context.Background(), fromAddress, nil) // =》读取一次，异步发送，自增nonce
			if err != nil {
				log.Fatal(err)
			}
			value := new(big.Int)
			value.SetString("300000000000000000", 10) // 3 ETH in Wei
			// value.SetString("10000000000000000000", 10) // 10 ETH in Wei
			// value.SetString("100000000000000000000", 10) // 100 ETH in Wei
			gasLimit := uint64(25000) // 默认gas limit
			gasPrice, err := client.SuggestGasPrice(context.Background())
			if err != nil {
				log.Fatal(err)
			}
			tx := types.NewTransaction(nonce, receiverAddress, value, gasLimit, gasPrice, nil)
			// 签署交易
			chainID, err := client.NetworkID(context.Background())
			if err != nil {
				log.Fatal(err)
			}
			signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
			if err != nil {
				log.Fatal(err)
			}
			// 发送交易
			err = client.SendTransaction(context.Background(), signedTx)
			if err != nil {
				fmt.Println("失败")
				fmt.Println(err)
				log.Fatal(err)
			}
			fmt.Println("账户序号：", i+1)
			fmt.Println("账户地址：", receiverAddress)
			fmt.Printf("交易已发送: %s\n", signedTx.Hash().Hex())
			time.Sleep(2 * time.Second)
		}
		done <- true
	}()
	// 等待转账结束
	<-done

}
