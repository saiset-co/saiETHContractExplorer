package main

import (
	"fmt"
	"github.com/prometheus/common/log"
	"github.com/tendermint/tendermint/libs/json"
	"github.com/tkanos/gonfig"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type configuration struct {
	HttpServer struct {
		Host string
		Port string
	}
	Address struct {
		Url string
	}
	Storage struct {
		Url  string
		Auth struct {
			Email    string
			Password string
		}
	}
	Operations []struct {
		Code 	string
		Name 	string
	}
	StartBlock int
	WebSocket  struct {
		Token 	string
		Url 	string
	}
	Contract string
	Geth     []string
	Sleep    int
}
type LastBlockType []struct {
	LastBlockNumber string `json:"last_block"`
}

var lastBlock LastBlockType
var startBlock int
var counter int
var config configuration

func getOperation(code string) (string, bool) {
	for _, v := range config.Operations {
		if v.Code == code {
			return v.Name, true
		}
	}
	return "", false
}

func process() {
	for {
		counter++
		fmt.Println("Iteration: ", counter)

		client, clientError := geth()

		if clientError != nil {
			fmt.Println("Geth problem!! ")
			panic(clientError)
		}

		blockId, blockErr := client.EthBlockNumber()

		if blockErr != nil {
			fmt.Println("Can't get last block!! ")
			panic(blockErr)
		}

		log.Info("Geth last block: %d", blockId)

		request := StorageRequest{collection: "last_block"}
		resultJsonString := CallSaiStrorage("get", request)
		_ = json.Unmarshal([]byte(resultJsonString), &lastBlock)

		if len(lastBlock) > 0 {
			lb, _ := strconv.Atoi(lastBlock[0].LastBlockNumber)
			startBlock = lb + 1
		} else if config.StartBlock > 0 {
			startBlock = config.StartBlock
		} else {
			startBlock = blockId
		}

		log.Infof("Start block: %d", startBlock)

		for i := startBlock; i <= blockId; i++ {
			trs, blockInfoErr := client.EthGetBlockByNumber(i, true)

			if blockInfoErr != nil {
				fmt.Println("Can't get block data!! ")
				panic(blockInfoErr)
			}

			if len(trs.Transactions) > 0 {
				fmt.Printf("Block %d from %d: %d transactions found.\n", i, blockId, len(trs.Transactions))
				for j := 0; j < len(trs.Transactions); j++ {
					if strings.ToLower(trs.Transactions[j].From) == strings.ToLower(config.Contract) || strings.ToLower(trs.Transactions[j].To) == strings.ToLower(config.Contract) {
						raw, _ := json.Marshal(trs.Transactions[j])
						operationId := trs.Transactions[j].Input[:10]

						data := bson.M{
							"From":      trs.Transactions[j].From,
							"To":        trs.Transactions[j].To,
							"Amount":    trs.Transactions[j].Value,
							"Raw":       raw,
							"Operation": "",
						}

						operation, ok := getOperation(operationId)

						if ok {
							data["Operation"] = operation
							WebSocketMessage(string(raw), config.WebSocket.Token)
						}

						transactionRequest := StorageRequest{collection: "transactions", data: data}
						CallSaiStrorage("save", transactionRequest)
						fmt.Printf("%d transaction has been updated.\n", trs.Transactions[j].TransactionIndex)
					}
				}
			} else {
				fmt.Printf("Block %d from %d: No transactions found.\n", i, blockId)
			}
		}

		if len(lastBlock) > 0 {
			lastBlock[0].LastBlockNumber = strconv.Itoa(blockId)
			request := StorageRequest{collection: "last_block", selectString: bson.M{"last_block": lastBlock[0].LastBlockNumber}, data: bson.M{"$set": lastBlock[0]}, options: "set"}
			_ = CallSaiStrorage("update", request)
		} else {
			request := StorageRequest{collection: "last_block", data: bson.M{"last_block": strconv.Itoa(blockId)}}
			_ = CallSaiStrorage("save", request)
		}

		time.Sleep(time.Duration(config.Sleep * 1000000000))
	}
}

func main() {
	configErr := gonfig.GetConf("config.json", &config)

	if configErr != nil {
		fmt.Println("Config missed!! ")
		panic(configErr)
	}

	go process()

	fmt.Println("Server start: http://" + config.HttpServer.Host + ":" + config.HttpServer.Port)

	http.HandleFunc("/", api)

	serverErr := http.ListenAndServe(config.HttpServer.Host+":"+config.HttpServer.Port, nil)

	if serverErr != nil {
		fmt.Println("Server error: ", serverErr)
	}
}
