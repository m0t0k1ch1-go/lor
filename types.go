package deckcode

import (
	"github.com/pkg/errors"
)

const (
	Format          uint8 = 1
	InitialVersion  uint8 = 1
	MaxKnownVersion uint8 = 5

	MaxKnownSet   uint64 = 6
	MaxCardNumber uint64 = 999

	CardCodeLength int = 7
)

var (
	ErrUnknownFormat  = errors.New("unknown format")
	ErrUnknownVersion = errors.New("unknown version")
	ErrUnknownSet     = errors.New("unknown set")
	ErrUnknownFaction = errors.New("unknown faction")

	ErrUnexpectedCardCount      = errors.New("unexpected card count")
	ErrUnexpectedCardNumber     = errors.New("unexpected card number")
	ErrUnexpectedCardCodeLength = errors.New("unexpected card code length")
)

var (
	factionIdentifierToVersion = map[string]uint8{
		"DE": 1,
		"FR": 1,
		"IO": 1,
		"NX": 1,
		"PZ": 1,
		"SI": 1,
		"BW": 2,
		"MT": 2,
		"SH": 3,
		"BC": 4,
		"RU": 5,
	}

	uint64ToFactionIdentifier = map[uint64]string{
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

	factionIdentifierToUint64 = map[string]uint64{
		"DE": 0,
		"FR": 1,
		"IO": 2,
		"NX": 3,
		"PZ": 4,
		"SI": 5,
		"BW": 6,
		"SH": 7,
		"MT": 9,
		"BC": 10,
		"RU": 12,
	}
)

type CardCodeAndCount struct {
	CardCode string `json:"cardCode"`
	Count    uint64 `json:"count"`
}

type Deck []CardCodeAndCount
