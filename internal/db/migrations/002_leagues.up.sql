-- Leagues table
CREATE TABLE leagues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    invite_code VARCHAR(8) UNIQUE NOT NULL,
    commissioner_id UUID NOT NULL REFERENCES users(id),
    max_members INT DEFAULT 10,
    scoring_type VARCHAR(20) DEFAULT 'ppr',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for invite code lookups
CREATE INDEX idx_leagues_invite_code ON leagues(invite_code);

-- Index for commissioner lookups
CREATE INDEX idx_leagues_commissioner ON leagues(commissioner_id);

-- League Members table
CREATE TABLE league_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    league_id UUID NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_name VARCHAR(100) NOT NULL,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(league_id, user_id)
);

-- Index for user's leagues
CREATE INDEX idx_league_members_user ON league_members(user_id);

-- Index for league's members
CREATE INDEX idx_league_members_league ON league_members(league_id);
