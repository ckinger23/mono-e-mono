-- Seasons table
CREATE TABLE seasons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    league_id UUID NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    year INT NOT NULL,
    current_week INT DEFAULT 1,
    status VARCHAR(20) DEFAULT 'active',
    champion_member_id UUID REFERENCES league_members(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(league_id, year)
);

-- Index for active seasons
CREATE INDEX idx_seasons_league_status ON seasons(league_id, status);

-- Standings table (aggregated from weekly results)
CREATE TABLE standings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    season_id UUID NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    member_id UUID NOT NULL REFERENCES league_members(id),
    weekly_wins INT DEFAULT 0,
    total_points DECIMAL(10,2) DEFAULT 0,
    best_week DECIMAL(10,2) DEFAULT 0,
    weeks_played INT DEFAULT 0,
    current_rank INT,
    UNIQUE(season_id, member_id)
);

-- Index for standings by season
CREATE INDEX idx_standings_season ON standings(season_id);

-- Index for standings by member
CREATE INDEX idx_standings_member ON standings(member_id);
