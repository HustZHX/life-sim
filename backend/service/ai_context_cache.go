package service

import (
	"sync"
	"time"

	"life-sim/backend/ai"
	"life-sim/backend/model"
	"life-sim/backend/store"
)

const aiContextBundleTTL = 30 * time.Minute

type aiContextBundle struct {
	ProfileHash     string
	ProfileTimeline string
	RecommendedAt   time.Time
}

type aiContextCache struct {
	mu       sync.RWMutex
	bundles  map[string]aiContextBundle
	sessions map[string]*ai.Session
}

func newAIContextCache() *aiContextCache {
	return &aiContextCache{
		bundles:  make(map[string]aiContextBundle),
		sessions: make(map[string]*ai.Session),
	}
}

func (c *aiContextCache) putBundle(characterID string, profile *model.Profile) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bundles[characterID] = aiContextBundle{
		ProfileHash:     store.ProfileHash(profile),
		ProfileTimeline: store.ProfileJSONTimeline(profile),
		RecommendedAt:   time.Now(),
	}
}

func (c *aiContextCache) getProfileTimelineJSON(characterID string, profile *model.Profile) string {
	hash := store.ProfileHash(profile)
	c.mu.RLock()
	b, ok := c.bundles[characterID]
	c.mu.RUnlock()
	if ok && b.ProfileHash == hash && time.Since(b.RecommendedAt) < aiContextBundleTTL {
		return b.ProfileTimeline
	}
	return store.ProfileJSONTimeline(profile)
}

func (c *aiContextCache) putSession(characterID string, sess *ai.Session) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessions[characterID] = sess
}

func (c *aiContextCache) getSession(characterID, profileHash string) *ai.Session {
	c.mu.RLock()
	sess := c.sessions[characterID]
	c.mu.RUnlock()
	if sess != nil && sess.Valid(profileHash) {
		return sess
	}
	return nil
}

func (c *aiContextCache) invalidate(characterID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.bundles, characterID)
	delete(c.sessions, characterID)
}
