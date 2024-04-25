package deckcode

import (
	"bytes"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/pkg/errors"
)

const (
	Format          uint8 = 1
	InitialVersion  uint8 = 1
	MaxKnownVersion uint8 = 5

	MaxKnownSet   uint64 = 9
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
	base32Encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

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

// CardCodeAndCount represents a card code and count.
type CardCodeAndCount struct {
	CardCode string `json:"cardCode"`
	Count    uint64 `json:"count"`
}

func (cardCodeAndCount CardCodeAndCount) version() uint8 {
	version, ok := factionIdentifierToVersion[cardCodeAndCount.CardCode[2:4]]
	if !ok {
		return 0
	}

	return version
}

// Deck represents a deck.
type Deck []CardCodeAndCount

func (deck Deck) maxVersion() uint8 {
	if len(deck) == 0 {
		return InitialVersion
	}

	maxVersion := InitialVersion

	for _, cardCodeAndCount := range deck {
		maxVersion = max(cardCodeAndCount.version(), maxVersion)
	}

	return maxVersion
}

// Encode encodes a deck to a deck code.
func Encode(deck Deck) (string, error) {
	buf := new(bytes.Buffer)

	if err := buf.WriteByte(Format<<4 | deck.maxVersion()); err != nil {
		return "", errors.Wrap(err, "failed to write format and version")
	}

	of3 := []CardCodeAndCount{}
	of2 := []CardCodeAndCount{}
	of1 := []CardCodeAndCount{}

	for _, cardCodeAndCount := range deck {
		switch cardCodeAndCount.Count {
		case 3:
			of3 = append(of3, cardCodeAndCount)
		case 2:
			of2 = append(of2, cardCodeAndCount)
		case 1:
			of1 = append(of1, cardCodeAndCount)
		default:
			return "", ErrUnexpectedCardCount
		}
	}

	groups3 := newSortedGroups(of3)
	groups2 := newSortedGroups(of2)
	groups1 := newSortedGroups(of1)

	if err := encodeGroups(buf, groups3); err != nil {
		return "", errors.Wrap(err, "failed to encode groups 3")
	}
	if err := encodeGroups(buf, groups2); err != nil {
		return "", errors.Wrap(err, "failed to encode groups 2")
	}
	if err := encodeGroups(buf, groups1); err != nil {
		return "", errors.Wrap(err, "failed to encode groups 1")
	}

	return base32Encoding.EncodeToString(buf.Bytes()), nil
}

// Decode decodes a deck code to a deck.
func Decode(deckCode string) (Deck, error) {
	b, err := base32Encoding.DecodeString(deckCode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to base32 decode")
	}

	buf := bytes.NewBuffer(b)

	formatAndVersionByte, err := buf.ReadByte()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read format and version")
	}

	// the original implementation does not validate the format
	// format := formatAndVersionByte >> 4

	version := formatAndVersionByte & 0xf
	if version > MaxKnownVersion {
		return nil, ErrUnknownVersion
	}

	deck := []CardCodeAndCount{}

	var i uint64
	for i = 3; i > 0; i-- {
		groupCount, err := binary.ReadUvarint(buf)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read uvarint representing number of groups")
		}

		var j uint64
		for j = 0; j < groupCount; j++ {
			cardNumberCount, err := binary.ReadUvarint(buf)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read uvarint representing number of card numbers")
			}

			set, err := binary.ReadUvarint(buf)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read uvarint representing set")
			}
			if set > MaxKnownSet {
				return nil, ErrUnknownSet
			}

			faction, err := binary.ReadUvarint(buf)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read uvarint representing faction")
			}

			factionIdentifier, ok := uint64ToFactionIdentifier[faction]
			if !ok {
				return nil, ErrUnknownFaction
			}

			var k uint64
			for k = 0; k < cardNumberCount; k++ {
				cardNumber, err := binary.ReadUvarint(buf)
				if err != nil {
					return nil, errors.Wrap(err, "failed to read uvarint representing card number")
				}
				if cardNumber > MaxCardNumber {
					return nil, ErrUnexpectedCardNumber
				}

				deck = append(deck, CardCodeAndCount{
					CardCode: fmt.Sprintf("%02d%s%03d", set, factionIdentifier, cardNumber),
					Count:    i,
				})
			}
		}
	}

	return deck, nil
}

func newSortedGroups(ofX []CardCodeAndCount) [][]CardCodeAndCount {
	sort.Slice(ofX, func(i, j int) bool {
		return ofX[i].CardCode < ofX[j].CardCode
	})

	groups := [][]CardCodeAndCount{}

	for len(ofX) > 0 {
		firstCardCodeAndCount := ofX[0]
		group := []CardCodeAndCount{
			firstCardCodeAndCount,
		}

		if len(ofX) == 1 {
			groups = append(groups, group)
			break
		}

		restOfX := ofX[1:]
		ofX = nil

		groupCode := firstCardCodeAndCount.CardCode[:4]

		for _, cardCodeAndCount := range restOfX {
			if cardCodeAndCount.CardCode[:4] == groupCode {
				group = append(group, cardCodeAndCount)
			} else {
				ofX = append(ofX, cardCodeAndCount)
			}
		}

		groups = append(groups, group)
	}

	sort.Slice(groups, func(i, j int) bool {
		return len(groups[i]) < len(groups[j])
	})

	return groups
}

func encodeGroups(w io.Writer, groups [][]CardCodeAndCount) error {
	if err := writeUvarint(w, uint64(len(groups))); err != nil {
		return errors.Wrap(err, "failed to write uvarint representing number of groups")
	}

	for _, group := range groups {
		if err := writeUvarint(w, uint64(len(group))); err != nil {
			return errors.Wrap(err, "failed to write uvarint representing number of card numbers")
		}

		set, err := strconv.ParseUint(group[0].CardCode[:2], 10, 64)
		if err != nil {
			return errors.Wrap(err, "failed to parse set")
		}
		if set > MaxKnownSet {
			return ErrUnknownSet
		}

		faction, ok := factionIdentifierToUint64[group[0].CardCode[2:4]]
		if !ok {
			return ErrUnknownFaction
		}

		if err := writeUvarint(w, set); err != nil {
			return errors.Wrap(err, "failed to write uvarint representing set")
		}
		if err := writeUvarint(w, faction); err != nil {
			return errors.Wrap(err, "failed to write uvarint representing faction")
		}

		for _, cardCodeAndCount := range group {
			if len(cardCodeAndCount.CardCode) != CardCodeLength {
				return ErrUnexpectedCardCodeLength
			}

			cardNumber, err := strconv.ParseUint(cardCodeAndCount.CardCode[4:], 10, 64)
			if err != nil {
				return errors.Wrap(err, "failed to parse card number")
			}

			if err := writeUvarint(w, cardNumber); err != nil {
				return errors.Wrap(err, "failed to write uvarint representing card number")
			}
		}
	}

	return nil
}

func writeUvarint(w io.Writer, x uint64) (err error) {
	b := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(b, x)
	_, err = w.Write(b[:n])
	return
}
