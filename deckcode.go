package deckcode

import (
	"bytes"
	"encoding/base32"
	"encoding/binary"

	"github.com/pkg/errors"
)

func Decode(code string) (Deck, error) {
	b, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(code)
	if err != nil {
		return nil, errors.Wrap(err, "failed to base32 decode")
	}

	buf := bytes.NewBuffer(b)

	formatAndVersionByte, err := buf.ReadByte()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the format and version")
	}

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
			return nil, errors.Wrap(err, "failed to read the uvarint representing the number of groups")
		}

		var j uint64
		for j = 0; j < groupCount; j++ {
			cardCount, err := binary.ReadUvarint(buf)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read the uvarint representing the number of cards")
			}

			set, err := binary.ReadUvarint(buf)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read the uvarint representing the set")
			}

			faction, err := binary.ReadUvarint(buf)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read the uvarint representing the faction")
			}

			var k uint64
			for k = 0; k < cardCount; k++ {
				cardNumber, err := binary.ReadUvarint(buf)
				if err != nil {
					return nil, errors.Wrap(err, "failed to read the uvarint representing the card number")
				}

				deck = append(deck, CardCodeAndCount{
					CardCode: CardCode{
						Set:        set,
						Faction:    Faction(faction),
						CardNumber: cardNumber,
					},
					Count: i,
				})
			}
		}
	}

	return deck, nil
}
