package common

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopLapse     time.Duration
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Fatalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClient Sends a bet to the server
func (c *Client) StartClient() {
	sigtermNotifier := make(chan os.Signal, 1)
	signal.Notify(sigtermNotifier, syscall.SIGTERM)

	c.createClientSocket()

	bet := getBetFromEnv(c.config.ID)

	err := c.sendBet(bet)
	if err != nil {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v", c.config.ID, err)
	} else {
		log.Infof("action: send_bet | result: success | dni: %v | number: %v", bet.Document, bet.Number)
	}

	c.conn.Close()

	select {
	case <-sigtermNotifier:
		log.Debugf("action: terminate_client | result: success | client_id: %v", c.config.ID)
		return
	default:
	}

	log.Debugf("action: exit_client | result: success | client_id: %v", c.config.ID)
}

func (c *Client) sendBet(bet Bet) error {

	length, bytes, err := bet.serialize()
	if err != nil {
		return err
	}

	err = binary.Write(c.conn, binary.BigEndian, uint16(length))
	if err != nil {
		return err
	}

	totalSent := 0
	for totalSent < length {
		sent, err := c.conn.Write(bytes[totalSent:])
		if err != nil {
			return err
		}
		totalSent += sent
	}

	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	if err != nil {
		return err
	}
	if msg != "OK\n" {
		return fmt.Errorf("received response from server: %v", msg)
	}

	return nil
}
