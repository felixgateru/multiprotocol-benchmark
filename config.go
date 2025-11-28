// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"time"
)

var (
	clientIDs = []string{
		"1d633a90-5a68-45f1-8295-d47e8bb099a3",
	}
	channelIDs = []string{
		"6ff4a7ea-4061-44da-9e73-594a65eebc43",
	}
	clientSecretMap = map[string]string{
		"1d633a90-5a68-45f1-8295-d47e8bb099a3": "244ba026-40d9-45da-b6e5-6cce4ae21a2a",
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
	DefaultLogLevel         = "debug"
	DefaultChannelCount     = 1
	DefaultClientCount      = 1
	DefaultMQTTBroker       = "tcp://messaging.magistrala.absmach.eu:1883"
	DefaultMQTTMessageCount = 1000000
	DefaultMQTTDelay        = 2 * time.Second
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
	DefaultWSMessageCount   = 1000
	DefaultWSDelay          = 3 * time.Second
	DefaultWSTimeout        = 60 * time.Second
	DefaultWSMessageSize    = 256
	DefaultProvisionFile    = ""
	DefaultDomainID         = "7b73fb42-fa56-48d6-8eb0-0955becb462c"
	DefaultSaveToFile       = false
	DefaultCACertPath       = ""
	DefaultTLSVerify        = true
	DefaultRunMQTT          = true
	DefaultRunCOAP          = true
	DefaultRunHTTP          = true
	DefaultRunWS            = true
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
		DomainID:         DefaultDomainID,
		SaveToFile:       DefaultSaveToFile,
		CACertPath:       DefaultCACertPath,
		TLSVerify:        DefaultTLSVerify,
		RunMQTT:          DefaultRunMQTT,
		RunCOAP:          DefaultRunCOAP,
		RunHTTP:          DefaultRunHTTP,
		RunWS:            DefaultRunWS,
	}
}
