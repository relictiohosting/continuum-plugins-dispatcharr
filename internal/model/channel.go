package model

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type ChannelIdentity struct {
	UpstreamID string
	GuideID    string
	Name       string
	LogoURL    string
	StreamURL  string
}

type Channel struct {
	ID        string
	SourceID  string
	Name      string
	Number    string
	GuideID   string
	LogoURL   string
	StreamURL string
}

func StableChannelID(mode SourceMode, identity ChannelIdentity) string {
	if mode == SourceModeXtream && strings.TrimSpace(identity.UpstreamID) != "" {
		return "xtream:" + strings.TrimSpace(identity.UpstreamID)
	}

	if strings.TrimSpace(identity.GuideID) != "" {
		return "m3u:" + normalize(identity.GuideID)
	}

	parts := []string{
		string(mode),
		normalize(identity.Name),
		normalize(identity.StreamURL),
		normalize(identity.LogoURL),
	}
	return string(mode) + ":" + stableHash(strings.Join(parts, "|"))
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func stableHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}
