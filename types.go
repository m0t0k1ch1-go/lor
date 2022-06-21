package deckcode

import (
	"github.com/pkg/errors"
)

const (
	MaxKnownVersion uint8  = 5
	MaxKnownSet     uint64 = 6
	MaxCardNumber   uint64 = 999
)

var (
	ErrUnknownVersion       = errors.New("unknown version")
	ErrUnknownSet           = errors.New("unknown set")
	ErrUnknownFaction       = errors.New("unknown faction")
	ErrUnexpectedCardNumber = errors.New("unexpected card number")
)

var (
	factionsMap = map[uint64]string{
		0:  "DE",
		1:  "FR",
		2:  "IO",
		3:  "NX",
		4:  "PZ",
		5:  "SI",
		6:  "BW",
		7:  "SH",
		9:  "MT",
		10: "BC",
		12: "RU",
	}
)

type CardCodeAndCount struct {
	CardCode string `json:"cardCode"`
	Count    uint64 `json:"count"`
}

type Deck []CardCodeAndCount
