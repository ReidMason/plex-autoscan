package sonarr

type SonarrWebhookBody struct {
	EventType      string `json:"eventType"`
	InstanceName   string `json:"instanceName"`
	ApplicationURL string `json:"applicationUrl"`
	Episodes       []struct {
		Title         string `json:"title"`
		ID            int    `json:"id"`
		EpisodeNumber int    `json:"episodeNumber"`
		SeasonNumber  int    `json:"seasonNumber"`
		SeriesID      int    `json:"seriesId"`
		TvdbID        int    `json:"tvdbId"`
	} `json:"episodes"`
	Series struct {
		Title    string `json:"title"`
		Path     string `json:"path"`
		Type     string `json:"type"`
		ID       int    `json:"id"`
		TvdbID   int    `json:"tvdbId"`
		TvMazeID int    `json:"tvMazeId"`
		Year     int    `json:"year"`
	} `json:"series"`
}
