package main

import (
  "io/ioutil"
  "encoding/json"
  "log"
  "net/http"
  "net/url"
  "time"
  "math/rand"
  "bytes"
  "encoding/base64"
  "os"
  godotenv "github.com/joho/godotenv"
)

type AuthResponse struct {
  AccessToken  string `json:"access_token"`
  RefreshToken string `json:"refresh_token"`
}

func init() {
  err := godotenv.Load()
  if err != nil {
    log.Fatalf("Error loading .env file: %v", err)
  }
}

var (
  clientID     = os.Getenv("CLIENT_ID")
  clientSecret = os.Getenv("CLIENT_SECRET")
  redirectURI  = "http://localhost:8888/callback"
  stateKey     = "spotify_auth_state"
  authURL        = "https://accounts.spotify.com/authorize"
  tokenURL       = "https://accounts.spotify.com/api/token"
  refreshTokenURL = "https://accounts.spotify.com/api/token"
  meURL          = "https://api.spotify.com/v1/me"
)

func generateRandomString(length int) string {
  var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

  rand.Seed(time.Now().UnixNano())

  b := make([]rune, length)
  for i := range b {
    b[i] = letterRunes[rand.Intn(len(letterRunes))]
  }
  return string(b)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
  state := generateRandomString(16)
  cookie := http.Cookie{
    Name:     stateKey,
    Value:    state,
    Path:     "/",
    MaxAge:   3600,
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteLaxMode,
  }

  http.SetCookie(w, &cookie)

  queryParams := url.Values{}
  queryParams.Add("response_type", "code")
  queryParams.Add("client_id", clientID)
  queryParams.Add("client_secret", clientSecret)
  queryParams.Add("redirect_uri", redirectURI)
  queryParams.Add("state", state)
  queryParams.Add("scope", "user-read-private user-read-email")
  authURL := authURL + "?" + queryParams.Encode()

  // Redirect the user to the authorization URL
  http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
  w.Write([]byte("success"))
}


func cbHandler(w http.ResponseWriter, r *http.Request) {
  queryStrings := r.URL.Query()
  state := queryStrings.Get("state")
  code := queryStrings.Get("code")
  storedState, err := r.Cookie(stateKey)
  if err != nil || state == "" || state != storedState.Value {
    http.Redirect(w, r, "/#?mismatched", http.StatusSeeOther)
    return
  }
  requestBody := url.Values{}
  requestBody.Add("code", code)
  requestBody.Add("redirect_uri", redirectURI)
  requestBody.Add("grant_type", "authorization_code")

  log.Println(requestBody)
  req, err := http.NewRequest(http.MethodPost, tokenURL,  bytes.NewBufferString(requestBody.Encode()))
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret))

  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("Authorization", authHeader)

  client := &http.Client{}
  authResp, err := client.Do(req)
  if err != nil {
    log.Println("error creating", err)
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  defer authResp.Body.Close()
  if authResp.StatusCode != http.StatusOK {
    http.Redirect(w, r, "/#?invalid_token", http.StatusSeeOther)
  }
  var authBody struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
  }
  bodyBytes, err := ioutil.ReadAll(authResp.Body)

  log.Println(string(bodyBytes))
  err = json.NewDecoder(authResp.Body).Decode(&authBody)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  options := url.Values{}
  options.Set("Authorization", "Bearer "+authBody.AccessToken)
  userReq, err := http.NewRequest(http.MethodGet, meURL, nil)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  userResp, err := client.Do(userReq)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  defer userResp.Body.Close()
  if authResp.StatusCode == http.StatusOK {
    var me struct {
      ID string `json:"id"`
    }

    err = json.NewDecoder(userResp.Body).Decode(&me)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    log.Println("welcome", me.ID)
    return
  }
  log.Fatal("something is very wrong")
}

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/login", authHandler)
  mux.HandleFunc("/callback", cbHandler)
  //http.HandleFunc("/refresh_token", handleRefreshToken)


  log.Print("Listening on port 8888")
  log.Fatal(http.ListenAndServe(":8888", mux))

}
