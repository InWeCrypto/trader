package main

import (
	"net/http"
	"net/url"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"

	cli "gopkg.in/urfave/cli.v2"
)

type Config struct {
	listen           string
	geth             string
	insight          string
	gasUpdate        time.Duration
	coinMarketCapURL string
	postgres         string
	proxy            string
	remote           *url.URL
}

var globalConfig Config

// default servr ip 120.77.208.222:8545
func main() {
	app := &cli.App{
		Name:    "trader",
		Usage:   "trader for the wallet",
		Version: "2.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "listen",
				Value: ":8888",
				Usage: "listening address:port",
			},
			&cli.StringFlag{
				Name:  "geth",
				Value: "http://127.0.0.1:8545",
				Usage: "geth nodes",
			},
			&cli.StringFlag{
				Name:  "insight",
				Value: "http://127.0.0.1:8545",
				Usage: "insight nodes",
			},
			&cli.StringFlag{
				Name:  "coinmarketcapurl",
				Value: "https://api.coinmarketcap.com/v1/ticker/?convert=CNY",
				Usage: "query market price",
			},
			&cli.DurationFlag{
				Name:  "gas_update",
				Value: 10 * time.Second,
				Usage: "set gas update period",
			},
			&cli.StringFlag{
				Name:  "account",
				Value: "0x6bd25eb2e60f5cc47c86abf6ba1b3d03fc74ee27",
				Usage: "address for executing constant queries for tokens",
			},
			&cli.StringFlag{
				Name:  "postgres",
				Value: "host=localhost port=5432 user=postgres dbname=trader sslmode=disable password=qwer1234",
				Usage: "postgres connection string",
			},
			&cli.StringFlag{
				Name:  "proxy",
				Value: "/",
				Usage: "proxy URI",
			},
			&cli.StringFlag{
				Name:  "remote",
				Value: "http://127.0.0.1:8545",
				Usage: "remote URL",
			},
		},
		Action: func(c *cli.Context) error {
			globalConfig.listen = c.String("listen")
			globalConfig.geth = c.String("geth")
			globalConfig.gasUpdate = c.Duration("gas_update")
			globalConfig.coinMarketCapURL = c.String("coinmarketcapurl")
			globalConfig.postgres = c.String("postgres")
			globalConfig.insight = c.String("insight")
			globalConfig.proxy = c.String("proxy")

			remote, err := url.Parse(c.String("remote"))

			if err != nil {
				return err
			}

			globalConfig.remote = remote

			log.Println("listen:", globalConfig.listen)
			log.Println("geth:", globalConfig.geth)
			log.Println("gas_update:", globalConfig.gasUpdate)
			log.Println("coinmarketcapurl:", globalConfig.coinMarketCapURL)
			log.Println("postgres:", globalConfig.postgres)
			log.Println("insight:", globalConfig.insight)
			log.Println("proxy:", globalConfig.proxy)
			log.Println("remote:", globalConfig.remote.String())
			// init
			go updateGasTask()
			defaultBlockTimeEstimator.init()

			// webapi
			router := httprouter.New()
			router.GET("/eth/getGasPrice", getGasPriceHandler)
			//router.GET("/market/priceList", priceListHandler)
			router.GET("/eth/blockNumber", blockNumberHandler)
			router.GET("/eth/blockPerSecond", blockPerSecondHandler)
			router.POST("/eth/getBalance", getBalanceHandler)
			router.POST("/eth/getTransactionCount", getTransactionCountHandler)
			router.POST("/eth/getEstimateGas", getEstimateGas)
			router.POST("/eth/getTransaction", getTransactionHandler)
			router.POST("/eth/sendRawTransaction", sendRawTransactionHandler)
			router.POST("/eth/tokens/balanceOf", tokenBalanceOfHandler)
			router.POST("/eth/tokens/totalSupply", tokenTotalSupplyHandler)
			router.POST("/eth/tokens/transferABI", transferABIHandler)
			// btc
			router.POST("/btc/estimatefee", estimatefee)
			router.POST("/btc/getTransactions", getBtcTransactions)
			router.POST("/btc/getTransactionById", getBtcTransactionById)
			router.POST("/btc/getUtxo", getUtxo)
			router.POST("/btc/send", send)
			router.POST("/btc/address", getAddress)

			// proxy
			router.POST(globalConfig.proxy, ReverseProxy)

			log.Fatal(http.ListenAndServe(globalConfig.listen, router))
			select {}
		},
	}
	app.Run(os.Args)
}
