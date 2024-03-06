package plex

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ReidMason/plex-autoscan/internal/logger"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Plex struct {
	client  HttpClient
	log     logger.Logger
	hostUrl *url.URL
	token   string
}

func NewPlex(log logger.Logger) *Plex {
	return &Plex{log: log}
}

func (p *Plex) Initialize(token, host string, port int, client HttpClient) error {
	p.log.Info("Initializing Plex", slog.String("host", host), slog.Int("port", port), slog.String("token", token))
	hostUrl, err := url.Parse(host)
	if err != nil {
		p.log.Error("Failed to parse Plex host url", slog.Any("error", err))
		return err
	}
	hostUrl.Host = hostUrl.Hostname() + ":" + strconv.Itoa(port)

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

func makeRequest(p Plex, request *http.Request) (io.ReadCloser, error) {
	if p.client == nil {
		p.log.Error("No client provided for Plex request")
		return nil, errors.New("No client provided for Plex request")
	}

	slog.Debug("Making request", slog.String("url", request.URL.String()))
	resp, err := p.client.Do(request)
	if err != nil {
		p.log.Error("Failed to make request", slog.Any("error", err))
		return nil, err
	}

	if resp.StatusCode >= 400 {
		p.log.Error("Request failed", slog.String("status", resp.Status))
		return nil, errors.New("Request failed with status: " + resp.Status)
	}

	return resp.Body, nil
}

func parseResponse[T any](p Plex, responseBody io.ReadCloser) (T, error) {
	var result T
	body, err := io.ReadAll(responseBody)
	if err != nil {
		p.log.Error("Failed to read response body", slog.Any("error", err))
		return result, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		p.log.Error("Failed to unmarshal response body", slog.Any("error", err))
		return result, err
	}

	return result, nil
}

func (p Plex) buildRequestUrl(path string) (string, error) {
	return url.JoinPath(p.hostUrl.String(), path)
}

func (p Plex) RescanPath(libraryKey, path string, season *int) error {
	p.log.Info("Refreshing season", slog.String("libraryKey", libraryKey), slog.String("path", path), slog.Any("season", season))
	url, err := p.buildRequestUrl("library/sections/" + libraryKey + "/refresh")
	if err != nil {
		p.log.Error("Failed to build Plex request url", slog.Any("error", err))
		return err
	}

	req, err := buildRequest("GET", url, p.token)
	if err != nil {
		p.log.Error("Failed to build Plex request", slog.Any("error", err))
		return err
	}

	if season != nil {
		path = path + "/Season " + strconv.Itoa(*season)
	}

	q := req.URL.Query()
	q.Add("path", path)
	req.URL.RawQuery = q.Encode()

	_, err = makeRequest(p, req)
	if err != nil {
		p.log.Error("Failed to make Plex request", slog.Any("error", err))
	}

	return err
}

func (p Plex) GetLibraries() ([]Library, error) {
	url, err := p.buildRequestUrl("library/sections")
	if err != nil {
		return nil, err
	}

	req, err := buildRequest("GET", url, p.token)
	if err != nil {
		return nil, err
	}

	body, err := makeRequest(p, req)
	if err != nil {
		return nil, err
	}
	response, err := parseResponse[PlexResponse[[]Library]](p, body)

	return response.MediaContainer.Directory, nil
}

func (p Plex) GetCurrentUser() (PlexUser, error) {
	var plexUser PlexUser
	req, err := buildRequest("GET", "https://plex.tv/api/v2/user", p.token)
	if err != nil {
		p.log.Error("Failed to build request for GetCurrentUser", slog.Any("error", err))
		return plexUser, err
	}

	body, err := makeRequest(p, req)
	if err != nil {
		p.log.Error("Failed to make request for GetCurrentUser", slog.Any("error", err))
		return plexUser, err
	}

	return parseResponse[PlexUser](p, body)
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
