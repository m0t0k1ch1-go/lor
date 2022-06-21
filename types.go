package deckcode

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	MaxKnownVersion uint8 = 5
)

var (
	ErrUnknownVersion = errors.New("unknown version of code")
)

var (
	FactionsMap = map[uint64]string{
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

type Faction uint64

func (faction Faction) Uint64() uint64 {
	return uint64(faction)
}

func (faction Faction) Identifier() string {
	return FactionsMap[faction.Uint64()]
}

type CardCode struct {
	Set        uint64  `json:"set"`
	Faction    Faction `json:"faction"`
	CardNumber uint64  `json:"cardNumber"`
}

func (code CardCode) String() string {
	return fmt.Sprintf("%02d%s%03d", code.Set, code.Faction.Identifier(), code.CardNumber)
}

type CardCodeAndCount struct {
	CardCode CardCode `json:"cardCode"`
	Count    uint64   `json:"count"`
}

type Deck []CardCodeAndCount
