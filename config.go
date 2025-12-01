// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"time"
)

var (
	domainID  = "dfd93a49-cb6c-49ba-b3dc-b4a79249d3ec"
	clientIDs = []string{
		"ec3ed66c-6d19-408e-ab30-6f9c187e51a1",
	}
	channelIDs = []string{
		"fda3eabc-ab61-421d-8a68-b93eb9ae2029",
	}
	// Map of client IDs to their corresponding client secrets
	clientSecretMap = map[string]string{
		"ec3ed66c-6d19-408e-ab30-6f9c187e51a1": "65d73373-4bf9-4e07-b86a-d48518909989",
	}
)

// Default configuration constants
const (
	DefaultUsername         = ""
	DefaultPassword         = ""
	DefaultUsersURL         = "https://cloud.magistrala.absmach.eu/api"
	DefaultClientsURL       = "https://cloud.magistrala.absmach.eu/api"
	DefaultDomainsURL       = "https://cloud.magistrala.absmach.eu/api"
	DefaultChannelsURL      = "https://cloud.magistrala.absmach.eu/api"
	DefaultLogLevel         = "info"
	DefaultChannelCount     = 1
	DefaultClientCount      = 1
	DefaultMQTTBroker       = "tcp://messaging.magistrala.absmach.eu:1883"
	DefaultMQTTMessageCount = 10000
	DefaultMQTTDelay        = 0 * time.Second
	DefaultMQTTTimeout      = 60 * time.Second
	DefaultMQTTMessageSize  = 256
	DefaultMQTTQoS          = byte(0)
	DefaultCOAPHost         = "messaging.magistrala.absmach.eu"
	DefaultCOAPPort         = 5684
	DefaultCOAPMessageCount = 1000000
	DefaultCOAPDelay        = 2 * time.Second
	DefaultCOAPTimeout      = 60 * time.Second
	DefaultCOAPMessageSize  = 256
	DefaultHTTPBaseUrl      = "https://messaging.magistrala.absmach.eu/api/http"
	DefaultHTTPMessageCount = 1000000
	DefaultHTTPDelay        = 3 * time.Second
	DefaultHTTPTimeout      = 60 * time.Second
	DefaultHTTPMessageSize  = 256
	DefaultWSBaseUrl        = "wss://messaging.magistrala.absmach.eu/api/ws"
	DefaultWSMessageCount   = 1000000
	DefaultWSDelay          = 3 * time.Second
	DefaultWSTimeout        = 60 * time.Second
	DefaultWSMessageSize    = 256
	DefaultProvisionFile    = ""
	DefaultSaveToFile       = false
	DefaultCACertPath       = ""
	DefaultTLSVerify        = true
	DefaultRunMQTT          = true
	DefaultRunCOAP          = false
	DefaultRunHTTP          = false
	DefaultRunWS            = false
)

type Config struct {
	Username         string
	Password         string
	UsersURL         string
	ClientsURL       string
	DomainsURL       string
	ChannelsURL      string
	LogLevel         string
	ChannelCount     int
	ClientCount      int
	MQTTBroker       string
	MQTTMessageCount int
	MQTTDelay        time.Duration
	MQTTTimeout      time.Duration
	MQTTMessageSize  int
	MQTTQoS          byte
	COAPHost         string
	COAPPort         int
	COAPMessageCount int
	COAPDelay        time.Duration
	COAPTimeout      time.Duration
	COAPMessageSize  int
	HTTPBaseUrl      string
	HTTPMessageCount int
	HTTPDelay        time.Duration
	HTTPTimeout      time.Duration
	HTTPMessageSize  int
	WSBaseUrl        string
	WSMessageCount   int
	WSDelay          time.Duration
	WSTimeout        time.Duration
	WSMessageSize    int
	ProvisionFile    string
	DomainID         string
	SaveToFile       bool
	CACertPath       string
	TLSVerify        bool
	RunMQTT          bool
	RunCOAP          bool
	RunHTTP          bool
	RunWS            bool
}

func NewConfig() Config {
	return Config{
		Username:         DefaultUsername,
		Password:         DefaultPassword,
		UsersURL:         DefaultUsersURL,
		ClientsURL:       DefaultClientsURL,
		DomainsURL:       DefaultDomainsURL,
		ChannelsURL:      DefaultChannelsURL,
		LogLevel:         DefaultLogLevel,
		ChannelCount:     DefaultChannelCount,
		ClientCount:      DefaultClientCount,
		MQTTBroker:       DefaultMQTTBroker,
		MQTTMessageCount: DefaultMQTTMessageCount,
		MQTTDelay:        DefaultMQTTDelay,
		MQTTTimeout:      DefaultMQTTTimeout,
		MQTTMessageSize:  DefaultMQTTMessageSize,
		MQTTQoS:          DefaultMQTTQoS,
		COAPHost:         DefaultCOAPHost,
		COAPPort:         DefaultCOAPPort,
		COAPMessageCount: DefaultCOAPMessageCount,
		COAPDelay:        DefaultCOAPDelay,
		COAPTimeout:      DefaultCOAPTimeout,
		COAPMessageSize:  DefaultCOAPMessageSize,
		HTTPBaseUrl:      DefaultHTTPBaseUrl,
		HTTPMessageCount: DefaultHTTPMessageCount,
		HTTPDelay:        DefaultHTTPDelay,
		HTTPTimeout:      DefaultHTTPTimeout,
		HTTPMessageSize:  DefaultHTTPMessageSize,
		WSBaseUrl:        DefaultWSBaseUrl,
		WSMessageCount:   DefaultWSMessageCount,
		WSDelay:          DefaultWSDelay,
		WSTimeout:        DefaultWSTimeout,
		WSMessageSize:    DefaultWSMessageSize,
		ProvisionFile:    DefaultProvisionFile,
		DomainID:         domainID,
		SaveToFile:       DefaultSaveToFile,
		CACertPath:       DefaultCACertPath,
		TLSVerify:        DefaultTLSVerify,
		RunMQTT:          DefaultRunMQTT,
		RunCOAP:          DefaultRunCOAP,
		RunHTTP:          DefaultRunHTTP,
		RunWS:            DefaultRunWS,
	}
}
