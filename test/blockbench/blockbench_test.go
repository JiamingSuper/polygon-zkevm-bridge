package blockbench

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/0xPolygonHermez/zkevm-bridge-service/bridgectrl"
	"github.com/0xPolygonHermez/zkevm-bridge-service/bridgectrl/pb"
	"github.com/0xPolygonHermez/zkevm-bridge-service/db"
	"github.com/0xPolygonHermez/zkevm-bridge-service/server"
	"github.com/0xPolygonHermez/zkevm-bridge-service/test/operations"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

type Timemap struct {
	mu          sync.Mutex
	TotalSend   int
	SuccessSend int
	SendTime    map[string]time.Time
	ClaimTime   map[string]time.Time
}

func depositEthFromL1(ctx context.Context, opsman *operations.Manager, t *testing.T, destAddress string, wg *sync.WaitGroup, timeMap *Timemap, privateKey string) {
	// 在 goroutine 开始时添加 WaitGroup 计数器
	wg.Add(1)
	go func() {
		defer wg.Done() // 在函数结束时调用 wg.Done()
		// 使用 WithTimeout 创建带有超时的上下文
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		// 在 select 中检查上下文的错误
		select {
		case <-ctx.Done():
			// 超时或上下文被取消
			fmt.Println("eth Deposit operation timed out")
			return
		default:
			amount := new(big.Int).SetUint64(10000000000000000)
			tokenAddr := common.Address{} // Eth
			destAddr := common.HexToAddress(destAddress)
			var destNetwork uint32 = 1
			// L1 to L2
			var err error
			send_time := time.Now()
			for i := 0; i < 100; i++ {
				err = opsman.SendL1Deposit(ctx, tokenAddr, amount, destNetwork, &destAddr, privateKey) // =》 修改能否返回交易hash，通过交易hash查看
				if err == nil {
					break
				}
				time.Sleep(50 * time.Millisecond)
				send_time = time.Now()
			}
			timeMap.mu.Lock()
			defer timeMap.mu.Unlock()
			timeMap.TotalSend++
			if err != nil {
				return
			}
			// 等待可以claim的时间  => ws的方式进行确认
			done := false
			for i := 0; i < 100; i++ {
				if done {
					break
				}
				var deposits []*pb.Deposit
				for a := 0; a < 10000; a++ {
					deposits, err = opsman.GetBridgeInfoByDestAddr(ctx, &destAddr)
					if err == nil {
						break
					}
					time.Sleep(50 * time.Millisecond)
				}
				// deposits, err := opsman.GetBridgeInfoByDestAddr(ctx, &destAddr)
				fmt.Printf("\n")
				fmt.Println("=====================================================")
				fmt.Println("查询账号： ", &destAddr)
				fmt.Println(deposits)
				fmt.Println("=====================================================")
				fmt.Printf("\n")
				for _, deposit := range deposits {
					// 判断 ReadyForClaim 是否为 true
					if deposit.ReadyForClaim {
						// claim
						// smtProof, globaExitRoot, err := opsman.GetClaimData(ctx, uint(deposits[0].NetworkId), uint(deposits[0].DepositCnt))
						// require.NoError(t, err)
						// err = opsman.SendL1Claim(ctx, deposits[0], smtProof, globaExitRoot)
						// require.NoError(t, err)
						claim_time := time.Now()
						timeMap.ClaimTime[destAddress] = claim_time
						timeMap.SuccessSend++
						timeMap.SendTime[destAddress] = send_time
						done = true
						break
					}
				}
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
}

func depositErc20FromL1(ctx context.Context, opsman *operations.Manager, t *testing.T, destAddress string, tokenAddress string, wg *sync.WaitGroup, timeMap *Timemap, privateKey string) {
	// 在 goroutine 开始时添加 WaitGroup 计数器
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		select {
		case <-ctx.Done():
			fmt.Println("erc20 Deposit operation timed out")
			return
		default:
			amount := new(big.Int).SetUint64(1000000000000000000)
			tokenAddr := common.HexToAddress(tokenAddress)
			destAddr := common.HexToAddress(destAddress)
			var destNetwork uint32 = 1
			var err error
			send_time := time.Now()
			for i := 0; i < 100; i++ {
				err = opsman.SendL1Deposit(ctx, tokenAddr, amount, destNetwork, &destAddr, privateKey)
				if err == nil {
					break
				}
				time.Sleep(50 * time.Millisecond)
				send_time = time.Now()
			}
			timeMap.mu.Lock()
			defer timeMap.mu.Unlock()
			timeMap.TotalSend++
			if err != nil {
				return
			}
			// 等待可以claim的时间
			done := false
			for i := 0; i < 10000; i++ {
				if done {
					break
				}
				var deposits []*pb.Deposit
				for a := 0; a < 100; a++ {
					deposits, err = opsman.GetBridgeInfoByDestAddr(ctx, &destAddr)
					if err == nil {
						break
					}
					time.Sleep(50 * time.Millisecond)
				}
				fmt.Printf("\n")
				fmt.Println("=====================================================")
				fmt.Println("查询账号： ", &destAddr)
				fmt.Println(deposits)
				fmt.Println("=====================================================")
				fmt.Printf("\n")
				for _, deposit := range deposits {
					// 判断 ReadyForClaim 是否为 true
					if deposit.ReadyForClaim {
						// claim
						// smtProof, globaExitRoot, err := opsman.GetClaimData(ctx, uint(deposits[0].NetworkId), uint(deposits[0].DepositCnt))
						// require.NoError(t, err)
						// err = opsman.SendL1Claim(ctx, deposits[0], smtProof, globaExitRoot)
						// require.NoError(t, err)
						claim_time := time.Now()
						timeMap.ClaimTime[destAddress] = claim_time
						timeMap.SuccessSend++
						timeMap.SendTime[destAddress] = send_time
						done = true
						break
					}
				}
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
}

func deployL1Erc20AndMint(ctx context.Context, opsman *operations.Manager, t *testing.T, transfer_num int) string {
	// var destNetwork uint32 = 1
	// amount1 := new(big.Int).SetUint64(1000000000000000000)
	// totAmount := new(big.Int).SetUint64(3500000000000000000)
	tokenAddr, _, err := opsman.DeployERC20(ctx, "A COIN", "ACO", "l1")
	require.NoError(t, err)
	t.Log("ERC20 token address: ", tokenAddr)
	done := make(chan bool)
	go func() {
		for i := 0; i < transfer_num; i++ {
			receiverAddress := common.HexToAddress(Addresses[i])
			err = opsman.MintERC20(ctx, tokenAddr, new(big.Int).SetUint64(5000000000000000000), "l1", PrivateKeys[i])
			require.NoError(t, err)
			origAddr := common.HexToAddress("0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC")
			balance, err := opsman.CheckAccountTokenBalance(ctx, "l1", tokenAddr, &origAddr, PrivateKeys[i])
			require.NoError(t, err)
			fmt.Printf("\n")
			fmt.Println(receiverAddress)
			t.Log("Init account balance l1: ", balance)
			fmt.Printf("\n")
		}
		done <- true
	}()
	<-done
	return tokenAddr.String()
}

func aggregator(ctx context.Context, timeMap *Timemap) {
	timeMap.mu.Lock()
	defer timeMap.mu.Unlock()

	var maxClaimTimeValue time.Time
	var minSendTimeValue time.Time
	var startTime time.Time
	var endTime time.Time
	totalLatency := 0

	firstIteration := true

	for key, claimTime := range timeMap.ClaimTime {
		// 判断是否存在相同的 key 在 SendTime 中
		if sendTime, ok := timeMap.SendTime[key]; ok {
			// 初始化最小值和最大值
			if firstIteration {
				minSendTimeValue = sendTime
				maxClaimTimeValue = claimTime
				firstIteration = false
			}

			// 判断最小值
			if sendTime.Before(minSendTimeValue) {
				minSendTimeValue = sendTime
				startTime = sendTime
			}

			// 判断最大值
			if claimTime.After(maxClaimTimeValue) {
				maxClaimTimeValue = claimTime
				endTime = claimTime
			}

			// 计算差值并相加
			totalLatency += int(claimTime.Sub(sendTime).Seconds())
		}
	}
	fmt.Println("============================================")
	fmt.Println("L1 -> L2 Total Deposit: ", timeMap.TotalSend)
	fmt.Println("L1 -> L2 Success Deposit: ", timeMap.SuccessSend)
	fmt.Println("L1 -> L2 Deposit TPS: ", float64(timeMap.SuccessSend)/endTime.Sub(startTime).Seconds())
	fmt.Println("============================================")
	fmt.Println("L1 -> L2 Total Latency: ", totalLatency)
	fmt.Println("L1 -> L2 Avg Latency: ", float64(totalLatency)/float64(timeMap.SuccessSend))
	fmt.Println("============================================")

}

func TestSend(t *testing.T) {
	fmt.Println("start Block Bench test")
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()
	opsCfg := &operations.Config{
		Storage: db.Config{
			Database: "postgres",
			Name:     "test_db",
			User:     "test_user",
			Password: "test_password",
			Host:     "localhost",
			Port:     "5435",
			MaxConns: 10,
		},
		BT: bridgectrl.Config{
			Store:  "postgres",
			Height: uint8(32),
		},
		BS: server.Config{
			GRPCPort:         "9090",
			HTTPPort:         "8080",
			CacheSize:        100000,
			DefaultPageLimit: 25,
			MaxPageLimit:     100,
			BridgeVersion:    "v1",
			DB: db.Config{
				Database: "postgres",
				Name:     "test_db",
				User:     "test_user",
				Password: "test_password",
				Host:     "localhost",
				Port:     "5435",
				MaxConns: 10,
			},
		},
	}
	opsman, err := operations.NewManager(ctx, opsCfg)
	require.NoError(t, err)

	// 设置要测试的账户数
	numGoroutines := 10
	transferEth(numGoroutines)
	t.Run("L1-L2 eth bridge", func(t *testing.T) {
		time_map_eth := Timemap{
			TotalSend:   0,
			SuccessSend: 0,
			SendTime:    make(map[string]time.Time),
			ClaimTime:   make(map[string]time.Time),
		}
		// 对涉及的账户进行准备

		var wg sync.WaitGroup
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			destAddress := Addresses[i]
			go func(destAddress string, i int) {
				// 在匿名 goroutine 中调用
				depositEthFromL1(ctx, opsman, t, destAddress, &wg, &time_map_eth, PrivateKeys[i])
				// 通知 WaitGroup 这个 goroutine 已经完成
				wg.Done()
			}(destAddress, i)
		}
		// 等待所有 goroutine 完成
		wg.Wait()
		aggregator(ctx, &time_map_eth)
	})

	t.Run("L1-L2 erc20 bridge", func(t *testing.T) {
		time_map_erc20 := Timemap{
			TotalSend:   0,
			SuccessSend: 0,
			SendTime:    make(map[string]time.Time),
			ClaimTime:   make(map[string]time.Time),
		}
		// 部署ERC20代币合约并给涉及的用户进行准备
		erc20Contract := deployL1Erc20AndMint(ctx, opsman, t, numGoroutines)
		// 开始发送
		var wg sync.WaitGroup
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			destAddress := Addresses[i]
			go func(destAddress string, i int) {
				depositErc20FromL1(ctx, opsman, t, destAddress, erc20Contract, &wg, &time_map_erc20, PrivateKeys[i])
				wg.Done()
			}(destAddress, i)
		}
		wg.Wait()
		aggregator(ctx, &time_map_erc20)
	})
}
