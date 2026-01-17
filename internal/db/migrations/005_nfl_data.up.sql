-- NFL Players (cached from Sleeper API)
CREATE TABLE nfl_players (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    position VARCHAR(10) NOT NULL,
    team VARCHAR(10),
    status VARCHAR(20),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for players by team
CREATE INDEX idx_nfl_players_team ON nfl_players(team);

-- Index for players by position
CREATE INDEX idx_nfl_players_position ON nfl_players(position);

-- NFL Weekly Stats (cached from Sleeper API)
CREATE TABLE nfl_weekly_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id VARCHAR(50) NOT NULL REFERENCES nfl_players(id),
    year INT NOT NULL,
    week INT NOT NULL,
    points_ppr DECIMAL(10,2),
    points_standard DECIMAL(10,2),
    points_half_ppr DECIMAL(10,2),
    stats JSONB,
    UNIQUE(player_id, year, week)
);

-- Index for stats by year and week
CREATE INDEX idx_nfl_weekly_stats_year_week ON nfl_weekly_stats(year, week);

-- Index for stats by player
CREATE INDEX idx_nfl_weekly_stats_player ON nfl_weekly_stats(player_id);
