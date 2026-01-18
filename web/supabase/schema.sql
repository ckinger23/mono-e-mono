-- Mono-e-Mono Database Schema for Supabase
-- Run this in the Supabase SQL Editor to set up your database

-- Enable UUID extension (usually already enabled)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- PROFILES (extends Supabase auth.users)
-- ============================================
CREATE TABLE profiles (
    id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    display_name VARCHAR(100) NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Auto-create profile when user signs up
CREATE OR REPLACE FUNCTION handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO public.profiles (id, display_name)
    VALUES (NEW.id, COALESCE(NEW.raw_user_meta_data->>'display_name', NEW.email));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE OR REPLACE TRIGGER on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW EXECUTE FUNCTION handle_new_user();

-- ============================================
-- LEAGUES
-- ============================================
CREATE TABLE leagues (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    invite_code VARCHAR(8) UNIQUE NOT NULL,
    commissioner_id UUID REFERENCES profiles(id) ON DELETE SET NULL,
    max_members INT DEFAULT 10,
    scoring_type VARCHAR(20) DEFAULT 'ppr' CHECK (scoring_type IN ('ppr', 'standard', 'half_ppr')),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Generate random invite code
CREATE OR REPLACE FUNCTION generate_invite_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.invite_code IS NULL THEN
        NEW.invite_code := UPPER(SUBSTRING(MD5(RANDOM()::TEXT) FROM 1 FOR 8));
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_invite_code
    BEFORE INSERT ON leagues
    FOR EACH ROW EXECUTE FUNCTION generate_invite_code();

-- ============================================
-- LEAGUE MEMBERS
-- ============================================
CREATE TABLE league_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    league_id UUID REFERENCES leagues(id) ON DELETE CASCADE,
    user_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    team_name VARCHAR(100) NOT NULL,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(league_id, user_id)
);

-- ============================================
-- SEASONS
-- ============================================
CREATE TABLE seasons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    league_id UUID REFERENCES leagues(id) ON DELETE CASCADE,
    year INT NOT NULL,
    current_week INT DEFAULT 1,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'complete')),
    champion_member_id UUID REFERENCES league_members(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(league_id, year)
);

-- ============================================
-- WEEKLY DRAFTS
-- ============================================
CREATE TABLE weekly_drafts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    season_id UUID REFERENCES seasons(id) ON DELETE CASCADE,
    member_id UUID REFERENCES league_members(id) ON DELETE CASCADE,
    week INT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'complete')),
    current_pick INT DEFAULT 0,
    current_team VARCHAR(10),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    UNIQUE(season_id, member_id, week)
);

-- ============================================
-- DRAFT PICKS
-- ============================================
CREATE TABLE draft_picks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    weekly_draft_id UUID REFERENCES weekly_drafts(id) ON DELETE CASCADE,
    pick_number INT NOT NULL CHECK (pick_number BETWEEN 1 AND 7),
    nfl_player_id VARCHAR(50) NOT NULL,
    player_name VARCHAR(100) NOT NULL,
    position VARCHAR(10) NOT NULL,
    team_drawn VARCHAR(10) NOT NULL,
    picked_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- WEEKLY RESULTS
-- ============================================
CREATE TABLE weekly_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    season_id UUID REFERENCES seasons(id) ON DELETE CASCADE,
    member_id UUID REFERENCES league_members(id) ON DELETE CASCADE,
    week INT NOT NULL,
    total_points DECIMAL(10,2) DEFAULT 0,
    weekly_rank INT,
    is_weekly_winner BOOLEAN DEFAULT FALSE,
    calculated_at TIMESTAMPTZ,
    UNIQUE(season_id, member_id, week)
);

-- ============================================
-- STANDINGS
-- ============================================
CREATE TABLE standings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    season_id UUID REFERENCES seasons(id) ON DELETE CASCADE,
    member_id UUID REFERENCES league_members(id) ON DELETE CASCADE,
    weekly_wins INT DEFAULT 0,
    total_points DECIMAL(10,2) DEFAULT 0,
    best_week DECIMAL(10,2) DEFAULT 0,
    weeks_played INT DEFAULT 0,
    current_rank INT,
    UNIQUE(season_id, member_id)
);

-- ============================================
-- NFL PLAYERS (cached from Sleeper API)
-- ============================================
CREATE TABLE nfl_players (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    position VARCHAR(10) NOT NULL,
    team VARCHAR(10),
    status VARCHAR(20),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- NFL WEEKLY STATS (cached from Sleeper API)
-- ============================================
CREATE TABLE nfl_weekly_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id VARCHAR(50) REFERENCES nfl_players(id),
    year INT NOT NULL,
    week INT NOT NULL,
    points_ppr DECIMAL(10,2),
    points_standard DECIMAL(10,2),
    points_half_ppr DECIMAL(10,2),
    stats JSONB,
    UNIQUE(player_id, year, week)
);

-- ============================================
-- ROW LEVEL SECURITY (RLS) POLICIES
-- ============================================

-- Enable RLS on all tables
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE leagues ENABLE ROW LEVEL SECURITY;
ALTER TABLE league_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE seasons ENABLE ROW LEVEL SECURITY;
ALTER TABLE weekly_drafts ENABLE ROW LEVEL SECURITY;
ALTER TABLE draft_picks ENABLE ROW LEVEL SECURITY;
ALTER TABLE weekly_results ENABLE ROW LEVEL SECURITY;
ALTER TABLE standings ENABLE ROW LEVEL SECURITY;
ALTER TABLE nfl_players ENABLE ROW LEVEL SECURITY;
ALTER TABLE nfl_weekly_stats ENABLE ROW LEVEL SECURITY;

-- Profiles: Users can read all profiles, update their own
CREATE POLICY "Profiles are viewable by everyone" ON profiles
    FOR SELECT USING (true);

CREATE POLICY "Users can update own profile" ON profiles
    FOR UPDATE USING (auth.uid() = id);

-- Leagues: Members can view their leagues, anyone can create
CREATE POLICY "Leagues viewable by members" ON leagues
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM league_members
            WHERE league_members.league_id = leagues.id
            AND league_members.user_id = auth.uid()
        )
        OR commissioner_id = auth.uid()
    );

CREATE POLICY "Anyone can create leagues" ON leagues
    FOR INSERT WITH CHECK (auth.uid() = commissioner_id);

CREATE POLICY "Commissioner can update league" ON leagues
    FOR UPDATE USING (commissioner_id = auth.uid());

-- League Members: Members can view co-members
CREATE POLICY "Members viewable by league members" ON league_members
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM league_members lm
            WHERE lm.league_id = league_members.league_id
            AND lm.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can join leagues" ON league_members
    FOR INSERT WITH CHECK (user_id = auth.uid());

-- Seasons: League members can view
CREATE POLICY "Seasons viewable by league members" ON seasons
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM league_members
            WHERE league_members.league_id = seasons.league_id
            AND league_members.user_id = auth.uid()
        )
    );

CREATE POLICY "Commissioner can manage seasons" ON seasons
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM leagues
            WHERE leagues.id = seasons.league_id
            AND leagues.commissioner_id = auth.uid()
        )
    );

-- Weekly Drafts: Users can view and manage their own drafts
CREATE POLICY "Users can view own drafts" ON weekly_drafts
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM league_members
            WHERE league_members.id = weekly_drafts.member_id
            AND league_members.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can manage own drafts" ON weekly_drafts
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM league_members
            WHERE league_members.id = weekly_drafts.member_id
            AND league_members.user_id = auth.uid()
        )
    );

-- Draft Picks: Users can view and create their own picks
CREATE POLICY "Users can view own picks" ON draft_picks
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM weekly_drafts wd
            JOIN league_members lm ON lm.id = wd.member_id
            WHERE wd.id = draft_picks.weekly_draft_id
            AND lm.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can create own picks" ON draft_picks
    FOR INSERT WITH CHECK (
        EXISTS (
            SELECT 1 FROM weekly_drafts wd
            JOIN league_members lm ON lm.id = wd.member_id
            WHERE wd.id = draft_picks.weekly_draft_id
            AND lm.user_id = auth.uid()
        )
    );

-- Weekly Results: League members can view
CREATE POLICY "Results viewable by league members" ON weekly_results
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM league_members
            WHERE league_members.id = weekly_results.member_id
            AND EXISTS (
                SELECT 1 FROM league_members lm2
                WHERE lm2.league_id = league_members.league_id
                AND lm2.user_id = auth.uid()
            )
        )
    );

-- Standings: League members can view
CREATE POLICY "Standings viewable by league members" ON standings
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM league_members
            WHERE league_members.id = standings.member_id
            AND EXISTS (
                SELECT 1 FROM league_members lm2
                WHERE lm2.league_id = league_members.league_id
                AND lm2.user_id = auth.uid()
            )
        )
    );

-- NFL Players: Public read access
CREATE POLICY "NFL players are public" ON nfl_players
    FOR SELECT USING (true);

-- NFL Stats: Public read access
CREATE POLICY "NFL stats are public" ON nfl_weekly_stats
    FOR SELECT USING (true);

-- ============================================
-- REALTIME SUBSCRIPTIONS
-- ============================================
-- Enable realtime for draft-related tables
ALTER PUBLICATION supabase_realtime ADD TABLE weekly_drafts;
ALTER PUBLICATION supabase_realtime ADD TABLE draft_picks;

-- ============================================
-- INDEXES FOR PERFORMANCE
-- ============================================
CREATE INDEX idx_league_members_league ON league_members(league_id);
CREATE INDEX idx_league_members_user ON league_members(user_id);
CREATE INDEX idx_seasons_league ON seasons(league_id);
CREATE INDEX idx_weekly_drafts_season ON weekly_drafts(season_id);
CREATE INDEX idx_weekly_drafts_member ON weekly_drafts(member_id);
CREATE INDEX idx_draft_picks_draft ON draft_picks(weekly_draft_id);
CREATE INDEX idx_weekly_results_season_week ON weekly_results(season_id, week);
CREATE INDEX idx_standings_season ON standings(season_id);
CREATE INDEX idx_nfl_players_team ON nfl_players(team);
CREATE INDEX idx_nfl_players_position ON nfl_players(position);
CREATE INDEX idx_nfl_weekly_stats_week ON nfl_weekly_stats(year, week);
