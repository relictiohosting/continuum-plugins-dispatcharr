package xtream

type LiveCategory struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
}

type LiveStream struct {
	Num          int    `json:"num"`
	Name         string `json:"name"`
	StreamType   string `json:"stream_type"`
	StreamID     int64  `json:"stream_id"`
	StreamIcon   string `json:"stream_icon"`
	EPGChannelID string `json:"epg_channel_id"`
	CategoryID   string `json:"category_id"`
}

type EPGListing struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	StartTimestamp string `json:"start_timestamp"`
	StopTimestamp  string `json:"stop_timestamp"`
}

type ShortEPGResponse struct {
	EPGListings []EPGListing `json:"epg_listings"`
}
