package common

import (
	"fmt"
	"strings"
)

type Bet struct {
	Agency    string
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    string
}

func getBetFromCSV(csv string, agency string) Bet {
	fields := strings.Split(csv, ",")
	return Bet{
		Agency:    agency,
		FirstName: fields[0],
		LastName:  fields[1],
		Document:  fields[2],
		Birthdate: fields[3],
		Number:    fields[4],
	}
}

func (b Bet) toCSV() string {
	return fmt.Sprintf(
		"%s,%s,%s,%s,%s,%s\n",
		b.Agency,
		b.FirstName,
		b.LastName,
		b.Document,
		b.Birthdate,
		b.Number,
	)
}
