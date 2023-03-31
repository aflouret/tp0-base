package common

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	betsRequest       = 1
	winnersRequest    = 2
	winnersOKResponse = 1
	maxPacketLength   = 8192
)

var ErrClientTerminated = errors.New("client terminated")

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

	err := c.sendBetsToServer()
	if err != nil {
		log.Errorf("action: send_bets | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	log.Infof("action: send_bets | result: success | client_id: %v", c.config.ID)

	winners, err := c.requestWinnersFromServer()
	if err != nil {
		log.Errorf("action: get_winners | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	log.Infof("action: get_winners | result: success | number_of_winners: %v | winners: %v", len(winners), winners)

	log.Debugf("action: exit_client | result: success | client_id: %v", c.config.ID)
}

func (c *Client) sendBetsToServer() error {
	sigtermNotifier := make(chan os.Signal, 1)
	signal.Notify(sigtermNotifier, syscall.SIGTERM)

	file, err := os.Open(fmt.Sprintf("agency-%v.csv", c.config.ID))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = c.setupConnection(betsRequest)
	if err != nil {
		return err
	}
	defer c.conn.Close()

	scanner := bufio.NewScanner(file)

	for {
		batch, err := c.readBatch(scanner)
		if err != nil {
			return err
		}
		if len(batch) == 0 {
			break
		}

		err = c.sendBatch(batch)
		if err != nil {
			return err
		}

		select {
		case <-sigtermNotifier:
			log.Debugf("action: terminate_client | result: success | client_id: %v", c.config.ID)
			return ErrClientTerminated
		default:
		}
	}
	return nil
}

func (c *Client) requestWinnersFromServer() ([]string, error) {
	sigtermNotifier := make(chan os.Signal, 1)
	signal.Notify(sigtermNotifier, syscall.SIGTERM)
	for {
		err := c.setupConnection(winnersRequest)
		if err != nil {
			return nil, err
		}

		// await response
		reader := bufio.NewReader(c.conn)
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		response := int(b)
		if response == winnersOKResponse {
			winners, err := c.getWinners()
			c.conn.Close()
			return winners, err
		}

		c.conn.Close()

		sleep := time.After(c.config.LoopPeriod)
		select {
		case <-sigtermNotifier:
			log.Debugf("action: terminate_client | result: success | client_id: %v", c.config.ID)
			return nil, ErrClientTerminated
		case <-sleep:
		}
	}
}

func (c *Client) getWinners() ([]string, error) {
	agencyID, err := strconv.Atoi(c.config.ID)
	if err != nil {
		return nil, err
	}

	err = binary.Write(c.conn, binary.BigEndian, uint16(agencyID))
	if err != nil {
		return nil, err
	}

	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	if err != nil {
		return nil, err
	}
	winners := strings.Split(strings.TrimSpace(msg), ",")
	if len(winners) == 1 && winners[0] == "" {
		winners = []string{}
	}

	return winners, nil
}

func (c *Client) setupConnection(requestType int) error {
	c.createClientSocket()
	err := binary.Write(c.conn, binary.BigEndian, uint8(requestType))
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) sendBatch(bets []Bet) error {

	length, bytes, err := serializeBatch(bets)
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

func (c *Client) readBatch(scanner *bufio.Scanner) ([]Bet, error) {
	var batch []Bet

	for i := 0; i < c.config.BatchSize; i++ {
		if scanner.Scan() {
			bet := getBetFromCSV(scanner.Text(), c.config.ID)
			batch = append(batch, bet)
		} else {
			break
		}
	}

	return batch, scanner.Err()
}

func serializeBatch(bets []Bet) (int, []byte, error) {
	var batchCSV string
	for _, bet := range bets {
		batchCSV += bet.toCSV()
	}

	bytes := []byte(batchCSV)
	length := len(bytes)
	if length > maxPacketLength {
		return 0, []byte{}, fmt.Errorf("data exceeds maximum length")
	}

	return length, bytes, nil
}
