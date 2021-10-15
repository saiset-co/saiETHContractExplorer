package main

import (
	"fmt"
	"github.com/onrik/ethrpc"
)

func geth() (*ethrpc.EthRPC, error) {
	if len(config.Geth) < 0 {
		fmt.Println("Geth configuration missed!! ")
		panic("Geth configuration missed!!")
	}

	client := ethrpc.New(config.Geth[0])

	_, clientErr := client.Web3ClientVersion()

	if clientErr != nil {
		var iclientErr error

		for i := 0; i < len(config.Geth); i++ {
			client = ethrpc.New(config.Geth[i])
			_, iclientErr = client.Web3ClientVersion()

			if iclientErr == nil {
				return client, nil
			}
		}

		if iclientErr != nil {
			fmt.Println("Geth client problem!! ")
			return client, clientErr
		}
	}

	return client, nil
}
