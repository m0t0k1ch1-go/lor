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

func Encode(deck Deck) (string, error) {
	buf := new(bytes.Buffer)

	version, err := getMinSupportedVersion(deck)
	if err != nil {
		return "", errors.Wrap(err, "failed to get the min supported version")
	}

	formatAndVersionByte := Format<<4 | version

	if err := buf.WriteByte(formatAndVersionByte); err != nil {
		return "", errors.Wrap(err, "failed to write the format and version")
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
		return "", errors.Wrap(err, "failed to encode the groups 3")
	}
	if err := encodeGroups(buf, groups2); err != nil {
		return "", errors.Wrap(err, "failed to encode the groups 2")
	}
	if err := encodeGroups(buf, groups1); err != nil {
		return "", errors.Wrap(err, "failed to encode the groups 1")
	}

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf.Bytes()), nil
}

func Decode(deckCode string) (Deck, error) {
	b, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(deckCode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to base32 decode")
	}

	buf := bytes.NewBuffer(b)

	formatAndVersionByte, err := buf.ReadByte()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the format and version")
	}

	format := formatAndVersionByte >> 4
	if !isSupportedFormat(format) {
		return nil, ErrUnknownFormat
	}

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
			cardNumberCount, err := binary.ReadUvarint(buf)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read the uvarint representing the number of card numbers")
			}

			set, err := binary.ReadUvarint(buf)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read the uvarint representing the set")
			}
			if set > MaxKnownSet {
				return nil, ErrUnknownSet
			}

			faction, err := binary.ReadUvarint(buf)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read the uvarint representing the faction")
			}

			factionIdentifier, ok := uint64ToFactionIdentifier[faction]
			if !ok {
				return nil, ErrUnknownFaction
			}

			var k uint64
			for k = 0; k < cardNumberCount; k++ {
				cardNumber, err := binary.ReadUvarint(buf)
				if err != nil {
					return nil, errors.Wrap(err, "failed to read the uvarint representing the card number")
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

func isSupportedFormat(format uint8) bool {
	return format == Format
}

func getMinSupportedVersion(deck Deck) (uint8, error) {
	if len(deck) == 0 {
		return InitialVersion, nil
	}

	minSupportedVersion := InitialVersion

	for _, cardCodeAndCount := range deck {
		version, ok := factionIdentifierToVersion[cardCodeAndCount.CardCode[2:4]]
		if !ok {
			return 0, ErrUnknownFaction
		}

		if version > minSupportedVersion {
			minSupportedVersion = version
		}
	}

	return minSupportedVersion, nil
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
		return errors.Wrap(err, "failed to write the uvarint representing the number of groups")
	}

	for _, group := range groups {
		if err := writeUvarint(w, uint64(len(group))); err != nil {
			return errors.Wrap(err, "failed to write the uvarint representing the number of card numbers")
		}

		set, err := strconv.ParseUint(group[0].CardCode[:2], 10, 64)
		if err != nil {
			return errors.Wrap(err, "failed to parse the set")
		}
		if set > MaxKnownSet {
			return ErrUnknownSet
		}

		faction, ok := factionIdentifierToUint64[group[0].CardCode[2:4]]
		if !ok {
			return ErrUnknownFaction
		}

		if err := writeUvarint(w, set); err != nil {
			return errors.Wrap(err, "failed to write the uvarint representing the set")
		}
		if err := writeUvarint(w, faction); err != nil {
			return errors.Wrap(err, "failed to write the uvarint representing the faction")
		}

		for _, cardCodeAndCount := range group {
			if len(cardCodeAndCount.CardCode) != CardCodeLength {
				return ErrUnexpectedCardCodeLength
			}

			cardNumber, err := strconv.ParseUint(cardCodeAndCount.CardCode[4:], 10, 64)
			if err != nil {
				return errors.Wrap(err, "failed to parse the card number")
			}

			if err := writeUvarint(w, cardNumber); err != nil {
				return errors.Wrap(err, "failed to write the uvarint representing the card number")
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
