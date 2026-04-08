package model

type SourceMode string

const (
	SourceModeXtream   SourceMode = "xtream"
	SourceModeM3UXMLTV SourceMode = "m3u_xmltv"
	LiveTVSourceID     string     = "source:live-tv"
)

type Source struct {
	ID   string
	Name string
	Mode SourceMode
}

func LiveTVSource(mode SourceMode) Source {
	return Source{ID: LiveTVSourceID, Name: "Live TV", Mode: mode}
}
