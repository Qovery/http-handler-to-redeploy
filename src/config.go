package src

import (
	"os"
	"strconv"
)

type Config struct {
	HttpPort                   int
	QoveryApiToken             string
	QoveryProdApplicationId    string
	QoveryStagingApplicationId string
}

func LoadConfig() Config {
	httpPortStr, set := os.LookupEnv("HTTP_PORT")
	httpPort, err := strconv.Atoi(httpPortStr)

	if set && err != nil {
		panic("HTTP_PORT is not an integer")
	}

	if !set {
		httpPort = 8080
	}

	qoveryApiToken, set := os.LookupEnv("Q_API_TOKEN")
	if !set {
		panic("Q_API_TOKEN is not set")
	}

	qoveryProdApplicationId, set := os.LookupEnv("Q_PROD_APPLICATION_ID")
	if !set {
		panic("Q_PROD_APPLICATION_ID is not set")
	}

	qoveryStagingApplicationId, set := os.LookupEnv("Q_STAGING_APPLICATION_ID")
	if !set {
		panic("Q_STAGING_APPLICATION_ID is not set")
	}

	return Config{
		HttpPort:                   httpPort,
		QoveryApiToken:             qoveryApiToken,
		QoveryProdApplicationId:    qoveryProdApplicationId,
		QoveryStagingApplicationId: qoveryStagingApplicationId,
	}
}
