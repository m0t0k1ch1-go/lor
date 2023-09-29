package deckcode_test

import (
	"bufio"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/m0t0k1ch1/lor-deckcode-go/deckcode"
)

type TestCase struct {
	DeckCode string
	Deck     deckcode.Deck
}

func TestEncode(t *testing.T) {
	tcs, err := loadTestCases()
	if err != nil {
		t.Fatalf("failed to load test data: %v", err)
	}

	for _, tc := range tcs {
		t.Run(tc.DeckCode, func(t *testing.T) {
			deckCode, err := deckcode.Encode(tc.Deck)
			if err != nil {
				t.Errorf("failed to encode deck: %v", err)
				return
			}
			if diff := cmp.Diff(tc.DeckCode, deckCode); len(diff) > 0 {
				t.Errorf("mismatch:\n%s", diff)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tcs, err := loadTestCases()
	if err != nil {
		t.Fatalf("failed to load test data: %v", err)
	}

	for _, tc := range tcs {
		t.Run(tc.DeckCode, func(t *testing.T) {
			deck, err := deckcode.Decode(tc.DeckCode)
			if err != nil {
				t.Errorf("failed to decode deck code: %v", err)
				return
			}
			if diff := cmp.Diff(tc.Deck, deck, cmp.Transformer("sort", func(in deckcode.Deck) deckcode.Deck {
				out := append(deckcode.Deck{}, in...)
				sort.Slice(out, func(i, j int) bool {
					return out[i].CardCode < out[j].CardCode
				})
				return out
			})); len(diff) > 0 {
				t.Errorf("mismatch:\n%s", diff)
			}
		})
	}
}

func loadTestCases() ([]TestCase, error) {
	f, err := os.Open("../testdata/DeckCodesTestData.txt")
	if err != nil {
		return nil, errors.Wrap(err, "failed to open test data file")
	}
	defer f.Close()

	tcs := []TestCase{}

	var tc TestCase
	startsNewDeck := true

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		row := scanner.Text()

		if len(row) == 0 {
			tcs = append(tcs, tc)
			tc = TestCase{}
			startsNewDeck = true
			continue
		}

		if startsNewDeck {
			tc.DeckCode = row
			startsNewDeck = false
			continue
		}

		parts := strings.Split(row, ":")
		if len(parts) != 2 {
			return nil, errors.New("malformed row")
		}
		if len(parts[1]) != deckcode.CardCodeLength {
			return nil, errors.New("malformed card code")
		}

		cardCount, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse card count")
		}

		tc.Deck = append(tc.Deck, deckcode.CardCodeAndCount{
			CardCode: parts[1],
			Count:    cardCount,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to scan test data file")
	}

	return tcs, nil
}
