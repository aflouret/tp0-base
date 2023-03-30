package common

import (
	"strings"
)

type Bet struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Document  string `json:"document"`
	Birthdate string `json:"birthdate"`
	Number    string `json:"number"`
}

func getBetFromCSV(csv string) Bet {
	fields := strings.Split(csv, ",")
	return Bet{
		FirstName: fields[0],
		LastName:  fields[1],
		Document:  fields[2],
		Birthdate: fields[3],
		Number:    fields[4],
	}
}
