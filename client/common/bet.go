package common

import (
	"encoding/json"
	"fmt"
	"os"
)

type Bet struct {
	Agency    string `json:"agency"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Document  string `json:"document"`
	Birthdate string `json:"birthdate"`
	Number    string `json:"number"`
}

func getBetFromEnv(agencyID string) Bet {
	firstName := os.Getenv("NOMBRE")
	lastName := os.Getenv("APELLIDO")
	document := os.Getenv("DOCUMENTO")
	birthdate := os.Getenv("NACIMIENTO")
	number := os.Getenv("NUMERO")

	return Bet{
		Agency:    agencyID,
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: birthdate,
		Number:    number,
	}
}

func (bet Bet) serialize() (int, []byte, error) {
	b, err := json.Marshal(bet)
	if err != nil {
		return 0, []byte{}, err
	}
	length := len(b)
	if length > 8192 {
		return 0, []byte{}, fmt.Errorf("data exceeds maximum length")
	}

	return length, b, nil
}
