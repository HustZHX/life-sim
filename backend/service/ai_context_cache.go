package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
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

type eraContextEntry struct {
	Text      string
	CreatedAt time.Time
}

type aiContextCache struct {
	mu          sync.RWMutex
	bundles     map[string]aiContextBundle
	sessions    map[string]*ai.Session
	eraContexts map[string]eraContextEntry
}

func newAIContextCache() *aiContextCache {
	return &aiContextCache{
		bundles:     make(map[string]aiContextBundle),
		sessions:    make(map[string]*ai.Session),
		eraContexts: make(map[string]eraContextEntry),
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

func (c *aiContextCache) eraContextKey(characterID, profileHash, characterMode string, cfg timelineJobConfig) string {
	sum := sha256.Sum256([]byte(cfg.Instructions))
	return fmt.Sprintf("%s:%s:%s:%d:%d:%s", characterID, profileHash, characterMode, cfg.StartYear, cfg.EndYear, hex.EncodeToString(sum[:4]))
}

func (c *aiContextCache) getEraContext(key string) (string, bool) {
	c.mu.RLock()
	entry, ok := c.eraContexts[key]
	c.mu.RUnlock()
	if !ok || time.Since(entry.CreatedAt) > aiContextBundleTTL {
		return "", false
	}
	return entry.Text, true
}

func (c *aiContextCache) putEraContext(key, text string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eraContexts[key] = eraContextEntry{Text: text, CreatedAt: time.Now()}
}

func (c *aiContextCache) invalidate(characterID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.bundles, characterID)
	delete(c.sessions, characterID)
	prefix := characterID + ":"
	for key := range c.eraContexts {
		if strings.HasPrefix(key, prefix) {
			delete(c.eraContexts, key)
		}
	}
}

func (c *aiContextCache) invalidateIfProfileChanged(characterID string, before, after *model.Profile) {
	if before != nil && after != nil && store.ProfileHash(before) == store.ProfileHash(after) {
		return
	}
	c.invalidate(characterID)
}
