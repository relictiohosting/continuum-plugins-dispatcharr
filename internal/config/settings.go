package config

import (
	"fmt"
	"strings"
)

const (
	DefaultChannelRefreshHours = 24
	DefaultEPGRefreshHours     = 6
)

type SourceMode string

const (
	SourceModeXtream   SourceMode = "xtream"
	SourceModeM3UXMLTV SourceMode = "m3u_xmltv"
)

type Settings struct {
	SourceMode        SourceMode
	XtreamBaseURL     string
	XtreamUsername    string
	XtreamPassword    string
	M3UURL            string
	EPGXMLURL         string
	LiveTVEnabled     bool
	ChannelRefreshH   int
	EPGRefreshH       int
	ModeSwitchWarning string
}

func (s Settings) Validate() error {
	switch s.SourceMode {
	case SourceModeXtream:
		if strings.TrimSpace(s.XtreamBaseURL) == "" {
			return fmt.Errorf("xtream base url is required")
		}
		if strings.TrimSpace(s.XtreamUsername) == "" {
			return fmt.Errorf("xtream username is required")
		}
		if strings.TrimSpace(s.XtreamPassword) == "" {
			return fmt.Errorf("xtream password is required")
		}
	case SourceModeM3UXMLTV:
		if strings.TrimSpace(s.M3UURL) == "" {
			return fmt.Errorf("m3u url is required")
		}
		if strings.TrimSpace(s.EPGXMLURL) == "" {
			return fmt.Errorf("epg xml url is required")
		}
	default:
		return fmt.Errorf("source mode is required")
	}

	if s.ChannelRefreshH == 0 {
		s.ChannelRefreshH = DefaultChannelRefreshHours
	}
	if s.EPGRefreshH == 0 {
		s.EPGRefreshH = DefaultEPGRefreshHours
	}
	if s.ChannelRefreshH <= 0 {
		return fmt.Errorf("channel refresh interval must be positive")
	}
	if s.EPGRefreshH <= 0 {
		return fmt.Errorf("epg refresh interval must be positive")
	}

	return nil
}
