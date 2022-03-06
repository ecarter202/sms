package sms

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"strings"

	"github.com/BurntSushi/toml"
)

type Client struct {
	OutboundServer string
	Port           int
	EmailAddress   string
	Password       string
	config         *Config
	smtpClient     *smtp.Client
}

type Config struct {
	FromAddress string            `toml:"from_address"`
	Carriers    map[string]string `toml:"carriers"`
}

func New(outboundServer, email, password string, port int) (*Client, error) {
	client := &Client{
		OutboundServer: outboundServer,
		EmailAddress:   email,
		Password:       password,
		Port:           port,
	}
	config := Config{Carriers: make(map[string]string)}

	buffer, err := ioutil.ReadFile("sms.toml")
	if err != nil {
		return nil, fmt.Errorf("reading toml file: %v", err)
	}

	if err := toml.Unmarshal(buffer, &config); err != nil {
		return nil, fmt.Errorf("unmarshaling toml file: %v", err)
	}

	client.config = &config

	return client, nil
}

func (client *Client) Deliver(number, carrier, message string) error {
	c := client.config.Carriers[strings.ToLower(carrier)]

	if c == "" {
		return errors.New("Unsupported carrier. Please check sms.toml for supported carriers.")
	}

	to := []string{fmt.Sprintf("%s%s", number, c)}
	auth := smtp.PlainAuth("", client.EmailAddress, client.Password, client.OutboundServer)
	if err := smtp.SendMail(fmt.Sprintf("%s:%d", client.OutboundServer, client.Port), auth, client.config.FromAddress,
		to, []byte(message)); err != nil {
		return fmt.Errorf("sending message: %v", err)
	}

	return nil
}
