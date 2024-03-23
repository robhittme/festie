package main 

import (
  "fmt"
  "bytes"
  "net/http"
  "encoding/json"
)

type SpotifyToken struct {
  AccessToken string `json:"access_token"`
  TokenType string `json:"token_type"`
  ExpiresIn int `json:"expires_in"`
}

type SpotifySearchArtistResponse struct {
    Artists struct {
        Href     string `json:"href"`
        Items    []struct {
            ExternalUrls struct {
                Spotify string `json:"spotify"`
            } `json:"external_urls"`
            Followers struct {
                Href  interface{} `json:"href"`
                Total int64       `json:"total"`
            } `json:"followers"`
            Genres []string `json:"genres"`
            Id     string   `json:"id"`
            Images []struct {
                Height int    `json:"height"`
                Url    string `json:"url"`
                Width  int    `json:"width"`
            } `json:"images"`
            Name       string `json:"name"`
            Popularity int64  `json:"popularity"`
            Type       string `json:"type"`
            Uri        string `json:"uri"`
        } `json:"items"`
        Limit    int64       `json:"limit"`
        Next     interface{} `json:"next"`
        Offset   int64       `json:"offset"`
        Previous interface{} `json:"previous"`
        Total    int64       `json:"total"`
    } `json:"artists"`
}

func getTopArtistTracks(at string, artistId string) {
  // depending on the festival you might need to change the country ?
  url := "https://api.spotify.com/v1/artists/" + artistId + "/top-tracks?country=US"
  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    panic(err)
  }
  req.Header.Set("Authorization", "Bearer " + at)

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()
  var responseBody map[string]interface{}
  err = json.NewDecoder(resp.Body).Decode(&responseBody)
  if err != nil {
    panic(err)
  }
  responseJson, err := json.MarshalIndent(responseBody, "", " ")
  if err != nil {
    panic(err)
  }

  fmt.Println(string(responseJson))
}

func getArtistId(artist SpotifySearchArtistResponse) string {
  return artist.Artists.Items[0].Id 
}

func searchArtists(at string, artist string) SpotifySearchArtistResponse {
  url := "https://api.spotify.com/v1/search?q=" + artist + "&type=artist"
  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    panic(err)
  }
  req.Header.Set("Authorization", "Bearer " + at)

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()
  var responseBody SpotifySearchArtistResponse
  err = json.NewDecoder(resp.Body).Decode(&responseBody)
  if err != nil {
    panic(err)
  }
  return responseBody
}

func main() {
  url := "https://accounts.spotify.com/api/token"
  client_id := "a3c992b0acc349bf9195e18036aa21d9"
  client_secret := "3051aae40b3842818b08a3a80c87e521"
  payload := bytes.NewBuffer([]byte("grant_type=client_credentials&client_id=" + client_id + "&client_secret=" + client_secret))
  req, err := http.NewRequest("POST", url, payload)
  if err != nil {
    panic(err)
  }

  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()
  var responseBody SpotifyToken 
  err = json.NewDecoder(resp.Body).Decode(&responseBody)
  if err != nil {
    panic(err)
  }
  artistSearch := searchArtists(responseBody.AccessToken, "tycho")
  //fmt.Printf("artistSearch %+v", getArtistId(artistSearch))
  getTopArtistTracks(responseBody.AccessToken, getArtistId(artistSearch))
  //responseJson, err := json.MarshalIndent(responseBody, "", " ")
  //if err != nil {
  //  panic(err)
  //}

  //fmt.Println(string(responseJson))
}
