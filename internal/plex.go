package plex

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Plex struct {
	client  HttpClient
	hostUrl *url.URL
	token   string
}

func NewPlex() *Plex {
	return &Plex{}
}

func (p *Plex) Initialize(token, host string, port int, client HttpClient) error {
	hostUrl, err := url.Parse(host)
	if err != nil {
		return err
	}
	hostUrl.Host = hostUrl.Hostname() + ":" + strconv.Itoa(port)

	log.Println("Initializing plex", hostUrl.String(), token)
	p.token = token
	p.hostUrl = hostUrl
	p.client = client

	return nil
}

func buildRequest(method, url, token string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("X-Plex-Token", token)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("accept", "application/json")

	return req, nil
}

func makeRequest[T any](client HttpClient, request *http.Request) (T, error) {
	var result T

	if client == nil {
		return result, errors.New("No client provided for Plex request")
	}

	slog.Info("Making request", slog.String("url", request.URL.String()))
	resp, err := client.Do(request)
	if err != nil {
		log.Println("Failed to make request", err)
		return result, err
	}

	if resp.StatusCode >= 400 {
		log.Printf("Request failed: %s", resp.Status)
		return result, errors.New("Request failed with status: " + resp.Status)
	}

	defer resp.Body.Close()
	return parseResponse[T](resp.Body)
}

func parseResponse[T any](responseBody io.ReadCloser) (T, error) {
	var result T
	body, err := io.ReadAll(responseBody)
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, err
	}

	return result, nil
}

func (p Plex) buildHostUrl(path string) (string, error) {
	log.Println("Building host url", p.hostUrl, path)
	return url.JoinPath(p.hostUrl.String(), path)
}

func (p Plex) RefreshSeason(libraryKey, path string, season int) error {
	url, err := p.buildHostUrl("library/sections/" + libraryKey + "/refresh")
	if err != nil {
		return err
	}

	req, err := buildRequest("GET", url, p.token)
	if err != nil {
		return err
	}

	path = path + "/Season " + strconv.Itoa(season)

	q := req.URL.Query()
	q.Add("path", path)
	req.URL.RawQuery = q.Encode()

	_, err = makeRequest[any](p.client, req)
	return err
}

func (p Plex) GetLibraries() ([]Library, error) {
	url, err := p.buildHostUrl("library/sections")
	if err != nil {
		return nil, err
	}

	req, err := buildRequest("GET", url, p.token)
	if err != nil {
		return nil, err
	}

	response, err := makeRequest[PlexResponse[[]Library]](p.client, req)
	if err != nil {
		return nil, err
	}

	return response.MediaContainer.Directory, nil
}

func (p Plex) GetCurrentUser() (PlexUser, error) {
	var plexUser PlexUser
	req, err := buildRequest("GET", "https://plex.tv/api/v2/user", p.token)
	if err != nil {
		log.Println("Failed to build request for GetCurrentUser", err)
		return plexUser, err
	}

	return makeRequest[PlexUser](p.client, req)
}

type PlexResponse[T any] struct {
	MediaContainer MediaContainer[T] `json:"MediaContainer"`
}

type MediaContainer[T any] struct {
	Directory T      `json:"Directory"`
	Title1    string `json:"title1"`
	Size      int    `json:"size"`
	AllowSync bool   `json:"allowSync"`
}

type Library struct {
	Scanner          string     `json:"scanner"`
	Type             string     `json:"type"`
	Art              string     `json:"art"`
	UUID             string     `json:"uuid"`
	Language         string     `json:"language"`
	Thumb            string     `json:"thumb"`
	Key              string     `json:"key"`
	Composite        string     `json:"composite"`
	Title            string     `json:"title"`
	Agent            string     `json:"agent"`
	Locations        []Location `json:"Location"`
	ContentChangedAt int        `json:"contentChangedAt"`
	Hidden           int        `json:"hidden"`
	UpdatedAt        int        `json:"updatedAt"`
	CreatedAt        int        `json:"createdAt"`
	ScannedAt        int        `json:"scannedAt"`
	Refreshing       bool       `json:"refreshing"`
	Directory        bool       `json:"directory"`
	Content          bool       `json:"content"`
	Filters          bool       `json:"filters"`
	AllowSync        bool       `json:"allowSync"`
}

type Location struct {
	Path string `json:"path"`
	ID   int    `json:"id"`
}

type PlexUser struct {
	Locale            *string      `json:"locale"`
	Thumb             string       `json:"thumb"`
	Title             string       `json:"title"`
	Country           string       `json:"country"`
	ScrobbleTypes     string       `json:"scrobbleTypes"`
	FriendlyName      string       `json:"friendlyName"`
	UUID              string       `json:"uuid"`
	MailingListStatus string       `json:"mailingListStatus"`
	AuthToken         string       `json:"authToken"`
	Email             string       `json:"email"`
	Username          string       `json:"username"`
	Subscription      Subscription `json:"subscription"`
	ID                int          `json:"id"`
	JoinedAt          int          `json:"joinedAt"`
	Confirmed         bool         `json:"confirmed"`
	MailingListActive bool         `json:"mailingListActive"`
	Protected         bool         `json:"protected"`
	HasPassword       bool         `json:"hasPassword"`
	EmailOnlyAuth     bool         `json:"emailOnlyAuth"`
}

type Subscription struct {
	SubscribedAt   string `json:"subscribedAt"`
	Status         string `json:"status"`
	PaymentService string `json:"paymentService"`
	Plan           string `json:"plan"`
	Active         bool   `json:"active"`
}
