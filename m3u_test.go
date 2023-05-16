package m3u

import (
	"testing"
)

func TestParseURL(t *testing.T) {
	parser, err := NewParser("https://areaengineering.co.uk/playlists/test-playlist.m3u", ResourceTypeURL)
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan Envelop, 1)
	parser.Parse(ch)

	i := 0
	for value := range ch {
		if value.Err != nil {
			t.Fatal(value.Err)
		}

		i++
	}

	if i == 0 {
		t.Fatalf("could not parse m3u")
	}
}

func TestParseFile(t *testing.T) {
	parser, err := NewParser("testdata/test-playlist.m3u", ResourceTypeFile)
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan Envelop, 1)
	parser.Parse(ch)

	i := 0
	for value := range ch {
		if value.Err != nil {
			t.Fatal(value.Err)
		}

		i++
	}

	if i == 0 {
		t.Fatalf("could not parse m3u")
	}
}
