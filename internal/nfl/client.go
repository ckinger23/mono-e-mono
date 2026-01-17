package nfl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL        = "https://api.sleeper.app/v1"
	DefaultTimeout = 30 * time.Second
)

// Client is a Sleeper API client
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Sleeper API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL: BaseURL,
	}
}

// NewClientWithTimeout creates a new client with custom timeout
func NewClientWithTimeout(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: BaseURL,
	}
}

// get performs a GET request and decodes the JSON response
func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	return nil
}

// SleeperPlayer represents a player from the Sleeper API
type SleeperPlayer struct {
	PlayerID       string   `json:"player_id"`
	FirstName      string   `json:"first_name"`
	LastName       string   `json:"last_name"`
	FullName       string   `json:"full_name"`
	Position       string   `json:"position"`
	Team           string   `json:"team"`
	Status         string   `json:"status"`
	InjuryStatus   string   `json:"injury_status"`
	Age            int      `json:"age"`
	YearsExp       int      `json:"years_exp"`
	College        string   `json:"college"`
	Height         string   `json:"height"`
	Weight         string   `json:"weight"`
	FantasyPositions []string `json:"fantasy_positions"`
	Active         bool     `json:"active"`
	SearchRank     int      `json:"search_rank"`
}

// SleeperStats represents weekly stats from the Sleeper API
type SleeperStats struct {
	PlayerID string `json:"player_id"`

	// Passing
	PassYds   float64 `json:"pass_yd"`
	PassTD    float64 `json:"pass_td"`
	PassInt   float64 `json:"pass_int"`
	Pass2Pt   float64 `json:"pass_2pt"`

	// Rushing
	RushYds   float64 `json:"rush_yd"`
	RushTD    float64 `json:"rush_td"`
	Rush2Pt   float64 `json:"rush_2pt"`

	// Receiving
	Receptions float64 `json:"rec"`
	RecYds     float64 `json:"rec_yd"`
	RecTD      float64 `json:"rec_td"`
	Rec2Pt     float64 `json:"rec_2pt"`

	// Fumbles
	FumblesLost float64 `json:"fum_lost"`

	// Defense/Special Teams
	DefTD         float64 `json:"def_td"`
	DefInt        float64 `json:"def_int"`
	DefSack       float64 `json:"sack"`
	DefFumbRec    float64 `json:"fum_rec"`
	DefSafety     float64 `json:"safe"`
	DefPtsAllowed float64 `json:"pts_allow"`
	DefBlkKick    float64 `json:"blk_kick"`

	// Kicking
	FGMade   float64 `json:"fgm"`
	FGMissed float64 `json:"fgmiss"`
	XPMade   float64 `json:"xpm"`
	XPMissed float64 `json:"xpmiss"`
}

// SleeperProjections represents projected stats
type SleeperProjections struct {
	SleeperStats
	GamesPlayed float64 `json:"gp"`
}

// GetAllPlayers fetches all NFL players from Sleeper
func (c *Client) GetAllPlayers(ctx context.Context) (map[string]SleeperPlayer, error) {
	var players map[string]SleeperPlayer
	if err := c.get(ctx, "/players/nfl", &players); err != nil {
		return nil, fmt.Errorf("fetching players: %w", err)
	}
	return players, nil
}

// GetWeeklyStats fetches stats for a specific week
func (c *Client) GetWeeklyStats(ctx context.Context, season int, week int) (map[string]SleeperStats, error) {
	path := fmt.Sprintf("/stats/nfl/regular/%d/%d", season, week)
	var stats map[string]SleeperStats
	if err := c.get(ctx, path, &stats); err != nil {
		return nil, fmt.Errorf("fetching stats for week %d: %w", week, err)
	}
	return stats, nil
}

// GetWeeklyProjections fetches projections for a specific week
func (c *Client) GetWeeklyProjections(ctx context.Context, season int, week int) (map[string]SleeperProjections, error) {
	path := fmt.Sprintf("/projections/nfl/regular/%d/%d", season, week)
	var projections map[string]SleeperProjections
	if err := c.get(ctx, path, &projections); err != nil {
		return nil, fmt.Errorf("fetching projections for week %d: %w", week, err)
	}
	return projections, nil
}

// GetNFLState fetches the current NFL state (current week, season, etc.)
type NFLState struct {
	Week               int    `json:"week"`
	Season             string `json:"season"`
	SeasonType         string `json:"season_type"`
	DisplayWeek        int    `json:"display_week"`
	SeasonStartDate    string `json:"season_start_date"`
	LeagueCreateSeason string `json:"league_create_season"`
}

func (c *Client) GetNFLState(ctx context.Context) (*NFLState, error) {
	var state NFLState
	if err := c.get(ctx, "/state/nfl", &state); err != nil {
		return nil, fmt.Errorf("fetching NFL state: %w", err)
	}
	return &state, nil
}
