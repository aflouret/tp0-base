package common

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
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
	BatchSize     int
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

	file, err := os.Open(fmt.Sprintf("agency-%v.csv", c.config.ID))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	c.createClientSocket()
	defer c.conn.Close()

	scanner := bufio.NewScanner(file)

	for {
		batch, err := readBatch(scanner, c.config.BatchSize)
		if err != nil {
			log.Errorf("action: send_bets | result: fail | client_id: %v | error: %v", c.config.ID, err)
		}
		if len(batch) == 0 {
			log.Infof("action: send_bets | result: success | client_id: %v", c.config.ID)
			break
		}

		err = c.sendBatch(batch)
		if err != nil {
			log.Errorf("action: send_bets | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}

		select {
		case <-sigtermNotifier:
			log.Debugf("action: terminate_client | result: success | client_id: %v", c.config.ID)
			return
		default:
		}

	}
	log.Debugf("action: exit_client | result: success | client_id: %v", c.config.ID)
}

func (c *Client) sendBatch(bets []Bet) error {

	length, bytes, err := serializeBatch(bets, c.config.ID)
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

func readBatch(scanner *bufio.Scanner, size int) ([]Bet, error) {
	var batch []Bet

	for i := 0; i < size; i++ {
		if scanner.Scan() {
			bet := getBetFromCSV(scanner.Text())
			batch = append(batch, bet)
		} else {
			break
		}
	}

	return batch, scanner.Err()
}

func serializeBatch(bets []Bet, agency string) (int, []byte, error) {
	batchJson := struct {
		Agency string `json:"agency"`
		Bets   []Bet  `json:"bets"`
	}{
		agency,
		bets,
	}

	b, err := json.Marshal(batchJson)
	if err != nil {
		return 0, []byte{}, err
	}
	length := len(b)
	if length > 8192 {
		return 0, []byte{}, fmt.Errorf("data exceeds maximum length")
	}

	return length, b, nil
}
