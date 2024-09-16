package main

import (
	"encoding/json"
	"io"
)

type Song struct {
	Title  string
	Artist string
}

type SongDecoder interface {
	DecodeSong(body io.ReadCloser) (Song, error)
}

type DefaultSongDecoder struct{}

func (d DefaultSongDecoder) DecodeSong(body io.ReadCloser) (Song, error) {
	var song Song
	err := json.NewDecoder(body).Decode(&song)
	return song, err
}
