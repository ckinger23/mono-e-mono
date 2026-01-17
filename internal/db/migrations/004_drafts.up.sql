-- Weekly Drafts (one per member per week)
CREATE TABLE weekly_drafts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    season_id UUID NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    member_id UUID NOT NULL REFERENCES league_members(id),
    week INT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    UNIQUE(season_id, member_id, week)
);

-- Index for finding drafts by season and week
CREATE INDEX idx_weekly_drafts_season_week ON weekly_drafts(season_id, week);

-- Index for finding drafts by member
CREATE INDEX idx_weekly_drafts_member ON weekly_drafts(member_id);

-- Draft Picks (each pick in a weekly draft)
CREATE TABLE draft_picks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    weekly_draft_id UUID NOT NULL REFERENCES weekly_drafts(id) ON DELETE CASCADE,
    pick_number INT NOT NULL,
    nfl_player_id VARCHAR(50) NOT NULL,
    player_name VARCHAR(100) NOT NULL,
    position VARCHAR(10) NOT NULL,
    team_drawn VARCHAR(50) NOT NULL,
    picked_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for picks by draft
CREATE INDEX idx_draft_picks_draft ON draft_picks(weekly_draft_id);

-- Weekly Results (scores after NFL games)
CREATE TABLE weekly_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    season_id UUID NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    member_id UUID NOT NULL REFERENCES league_members(id),
    week INT NOT NULL,
    total_points DECIMAL(10,2) DEFAULT 0,
    weekly_rank INT,
    is_weekly_winner BOOLEAN DEFAULT FALSE,
    calculated_at TIMESTAMPTZ,
    UNIQUE(season_id, member_id, week)
);

-- Index for results by season and week
CREATE INDEX idx_weekly_results_season_week ON weekly_results(season_id, week);

-- Index for results by member
CREATE INDEX idx_weekly_results_member ON weekly_results(member_id);
