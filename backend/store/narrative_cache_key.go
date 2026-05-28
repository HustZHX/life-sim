package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"life-sim/backend/model"
)

// NarrativeContentKey 由叙事类型与节点实质内容推导，用于 version_id 变更时的缓存回退。
func NarrativeContentKey(kind string, nodes []model.LifeNode, person string) string {
	h := sha256.New()
	_, _ = io.WriteString(h, kind)
	_, _ = io.WriteString(h, "\x00")
	_, _ = io.WriteString(h, model.NormalizeLightNovelPerson(person))
	for _, n := range nodes {
		_, _ = fmt.Fprintf(h, "\n%d|%s|%s|%s|%s", n.Sequence, n.Title, n.Events, n.Thoughts, n.PersonalitySnapshot)
	}
	return hex.EncodeToString(h.Sum(nil)[:16])
}
