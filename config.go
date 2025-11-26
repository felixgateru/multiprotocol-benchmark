// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Username         string        `env:"MG_USERNAME" envDefault:""`
	Password         string        `env:"MG_PASSWORD" envDefault:""`
	UsersURL         string        `env:"MG_USERS_URL" envDefault:"http://localhost:9002"`
	ClientsURL       string        `env:"MG_CLIENTS_URL" envDefault:"http://localhost:9006"`
	DomainsURL       string        `env:"MG_DOMAINS_URL" envDefault:"http://localhost:9003"`
	ChannelsURL      string        `env:"MG_CHANNELS_URL" envDefault:"http://localhost:9005"`
	LogLevel         string        `env:"LOG_LEVEL" envDefault:"info"`
	ChannelCount     int           `env:"CHANNEL_COUNT" envDefault:"1"`
	ClientCount      int           `env:"CLIENT_COUNT" envDefault:"1"`
	MQTTBroker       string        `env:"MQTT_BROKER_URL" envDefault:"tcp://localhost:1883"`
	MQTTMessageCount int           `env:"MQTT_MESSAGE_COUNT" envDefault:"1000"`
	MQTTDelay        time.Duration `env:"MQTT_MESSAGE_DELAY" envDefault:"10ms"`
	MQTTTimeout      time.Duration `env:"MQTT_TIMEOUT" envDefault:"60s"`
	MQTTMessageSize  int           `env:"MQTT_MESSAGE_SIZE" envDefault:"256"`
	MQTTQoS          byte          `env:"MQTT_QOS" envDefault:"1"`
	COAPHost         string        `env:"COAP_SERVER_HOST" envDefault:"localhost"`
	COAPPort         int           `env:"COAP_SERVER_PORT" envDefault:"5683"`
	COAPMessageCount int           `env:"COAP_MESSAGE_COUNT" envDefault:"1000"`
	COAPDelay        time.Duration `env:"COAP_MESSAGE_DELAY" envDefault:"1s"`
	COAPTimeout      time.Duration `env:"COAP_TIMEOUT" envDefault:"60s"`
	COAPMessageSize  int           `env:"COAP_MESSAGE_SIZE" envDefault:"256"`
	HTTPBaseUrl      string        `env:"HTTP_SERVER_BASE_URL" envDefault:"localhost"`
	HTTPMessageCount int           `env:"HTTP_MESSAGE_COUNT" envDefault:"1000"`
	HTTPDelay        time.Duration `env:"HTTP_MESSAGE_DELAY" envDefault:"1s"`
	HTTPTimeout      time.Duration `env:"HTTP_TIMEOUT" envDefault:"60s"`
	HTTPMessageSize  int           `env:"HTTP_MESSAGE_SIZE" envDefault:"256"`
	WSBaseUrl        string        `env:"WS_SERVER_BASE_URL" envDefault:"localhost"`
	WSMessageCount   int           `env:"WS_MESSAGE_COUNT" envDefault:"1000"`
	WSDelay          time.Duration `env:"WS_MESSAGE_DELAY" envDefault:"1s"`
	WSTimeout        time.Duration `env:"WS_TIMEOUT" envDefault:"60s"`
	WSMessageSize    int           `env:"WS_MESSAGE_SIZE" envDefault:"256"`
	ProvisionFile    string        `env:"CONFIG_TOML" envDefault:""`
	DomainID         string        `env:"DOMAIN_ID" envDefault:""`
	SaveToFile       bool          `env:"SAVE_TO_FILE" envDefault:"false"`
	CACertPath       string        `env:"CA_CERT_PATH" envDefault:""`
	TLSVerify        bool          `env:"TLS_VERIFY" envDefault:"true"`
	RunMQTT          bool          `env:"RUN_MQTT" envDefault:"true"`
	RunCOAP          bool          `env:"RUN_COAP" envDefault:"true"`
	RunHTTP          bool          `env:"RUN_HTTP" envDefault:"true"`
	RunWS            bool          `env:"RUN_WS" envDefault:"true"`
}

func NewConfig(opts env.Options) (Config, error) {
	c := Config{}
	if err := env.ParseWithOptions(&c, opts); err != nil {
		return Config{}, err
	}

	return c, nil
}

type ProvisionFile struct {
	Channels []string `toml:"channels"`
	Clients  []struct {
		ID     string `toml:"id"`
		Secret string `toml:"secret"`
	} `toml:"clients"`
}
