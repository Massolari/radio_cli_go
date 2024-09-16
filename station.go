package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Station struct {
	Name   string
	stream string
	song   StationSong
}

type StationSong struct {
	endpoint string
	decoder  SongDecoder
}

func (s Station) GetSong() (Song, error) {
	if s.song.endpoint == "" {
		return Song{}, nil
	}
	response, err := http.Get(s.song.endpoint)
	if err != nil {
		return Song{}, err
	}
	defer response.Body.Close()

	song, err := s.song.decoder.DecodeSong(response.Body)
	if err != nil {
		return Song{}, err
	}

	return song, nil
}

var ChristianRock = Station{
	Name:   "Christian Rock",
	stream: "https://listen.christianrock.net/stream/11/",
	song: StationSong{
		endpoint: "https://www.christianrock.net/iphonecrdn.php",
		decoder:  DefaultSongDecoder{},
	},
}
var ChristianHits = Station{
	Name:   "Christian Hits",
	stream: "https://listen.christianrock.net/stream/12/",
	song: StationSong{
		endpoint: "https://www.christianrock.net/iphonechdn.php",
		decoder:  DefaultSongDecoder{},
	},
}
var ChristianLofi = Station{
	Name:   "Christian Lo-fi",
	stream: "https://www.youtube.com/embed/-YJmGR2tD0k",
}
var GospelMix = Station{
	Name:   "Gospel Mix",
	stream: "https://servidor33-3.brlogic.com:8192/live",
	song: StationSong{
		endpoint: "https://d36nr0u3xmc4mm.cloudfront.net/index.php/api/streaming/status/8192/2e1cbe43529055ddda74868d2db9ae98/SV4BR",
		decoder:  GospelMixSongDecoder{},
	},
}

type GospelMixSongDecoder struct{}

func (g GospelMixSongDecoder) DecodeSong(body io.ReadCloser) (Song, error) {
	var track struct {
		CurrentTrack string `json:"currentTrack"`
	}
	err := json.NewDecoder(body).Decode(&track)

	var titleArtist []string
	for _, part := range strings.Split(track.CurrentTrack, " - ") {
		// don't include numbers in the titleArtist slice
		if _, err := strconv.Atoi(part); err == nil {
			continue
		}
		if part == "Ao Vivo" {
			continue
		}
		titleArtist = append(titleArtist, part)
	}

	song := Song{
		Title:  titleArtist[1],
		Artist: titleArtist[0],
	}

	return song, err
}

var Melodia = Station{
	Name:   "Melodia",
	stream: "https://14543.live.streamtheworld.com/MELODIAFMAAC.aac",
	song: StationSong{
		endpoint: "https://np.tritondigital.com/public/nowplaying?mountName=MELODIAFMAAC&numberToFetch=1&eventType=track",
		decoder:  MelodiaSongDecoder{},
	},
}

type MelodiaSongDecoder struct{}

func (m MelodiaSongDecoder) DecodeSong(body io.ReadCloser) (Song, error) {
	var xmlData struct {
		NowPlayingInfo struct {
			Properties []struct {
				Name  string `xml:"name,attr"`
				Value string `xml:",cdata"`
			} `xml:"property"`
		} `xml:"nowplaying-info"`
	}
	bytesBody, err := io.ReadAll(body)
	if err != nil {
		return Song{}, err
	}
	err = xml.Unmarshal(bytesBody, &xmlData)
	if err != nil {
		return Song{}, err
	}

	var song Song

	var artists []string
	for _, property := range xmlData.NowPlayingInfo.Properties {
		if property.Name == "track_artist_name" {
			artists = append(artists, property.Value)
		}
	}

	song.Artist = strings.Join(artists, ", ")

	for _, property := range xmlData.NowPlayingInfo.Properties {
		if property.Name == "cue_title" {
			song.Title = property.Value
			break
		}
	}

	return song, nil
}
