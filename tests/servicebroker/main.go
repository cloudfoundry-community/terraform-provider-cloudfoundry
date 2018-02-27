package main

import (
	"os"

	"net/http"

	"github.com/benlaplanche/cf-basic-auth-route-service/servicebroker/broker"
	"github.com/benlaplanche/cf-basic-auth-route-service/servicebroker/config"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-golang/lager"
)

func main() {
	logger := lager.NewLogger("p-basic-auth-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.ERROR))

	brokerConfigPath := configPath()

	parsedConfig, err := config.ParseConfig(brokerConfigPath)
	if err != nil {
		logger.Fatal("Loading config file", err, lager.Data{
			"broker-config-path": brokerConfigPath,
		})
	}

	brokerCredentials := brokerapi.BrokerCredentials{
		Username: parsedConfig.BrokerConfiguration.BrokerUserName,
		Password: parsedConfig.BrokerConfiguration.BrokerPassword,
	}

	service := &broker.BasicAuthBroker{Config: parsedConfig}
	newBroker := brokerapi.New(service, logger, brokerCredentials)

	http.Handle("/", newBroker)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Fatal("http-listen", http.ListenAndServe("0.0.0.0:"+port, nil))

}

func configPath() string {
	brokerConfigYamlPath := os.Getenv("BROKER_CONFIG_PATH")
	if brokerConfigYamlPath == "" {
		panic("BROKER_CONFIG_PATH not set")
	}
	return brokerConfigYamlPath
}
