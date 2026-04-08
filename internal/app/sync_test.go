package app

import (
	"context"
	"testing"

	"github.com/relictiohosting/continuum-plugins/dispatcharr/internal/cache"
	"github.com/relictiohosting/continuum-plugins/dispatcharr/internal/config"
	"github.com/relictiohosting/continuum-plugins/dispatcharr/internal/upstream/xtream"
)

func TestSyncStoresChannelsAndPrograms(t *testing.T) {
	t.Parallel()

	store := cache.NewStore()
	service := NewService(Dependencies{
		Store: store,
		XtreamFactory: func(string, string, string) XtreamClient {
			return &stubXtreamClient{
				streams: []xtream.LiveStream{{Num: 1, Name: "News HD", StreamID: 1001, EPGChannelID: "news.hd"}},
				epg:     xtream.ShortEPGResponse{EPGListings: []xtream.EPGListing{{ID: "epg-1", Title: "Morning News", StartTimestamp: "1700000000", StopTimestamp: "1700003600"}}},
			}
		},
	})

	err := service.SyncNow(context.Background(), config.Settings{
		SourceMode:      config.SourceModeXtream,
		XtreamBaseURL:   "https://dispatcharr.example.com",
		XtreamUsername:  "demo",
		XtreamPassword:  "secret",
		ChannelRefreshH: 24,
		EPGRefreshH:     6,
	}, 200)
	if err != nil {
		t.Fatalf("expected sync success, got %v", err)
	}

	snapshot := store.Current()
	if len(snapshot.Catalog.Channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(snapshot.Catalog.Channels))
	}
	if len(snapshot.Catalog.Programs) != 1 {
		t.Fatalf("expected 1 program, got %d", len(snapshot.Catalog.Programs))
	}
	if snapshot.Health.LastSuccessUnix != 200 {
		t.Fatalf("expected sync success timestamp, got %d", snapshot.Health.LastSuccessUnix)
	}
}

func TestSyncKeepsStaleSnapshotOnFailure(t *testing.T) {
	t.Parallel()

	store := cache.NewStore()
	store.Replace(cache.Snapshot{})

	service := NewService(Dependencies{
		Store: store,
		XtreamFactory: func(string, string, string) XtreamClient {
			return &stubXtreamClient{streamsErr: context.DeadlineExceeded}
		},
	})

	err := service.SyncNow(context.Background(), config.Settings{
		SourceMode:      config.SourceModeXtream,
		XtreamBaseURL:   "https://dispatcharr.example.com",
		XtreamUsername:  "demo",
		XtreamPassword:  "secret",
		ChannelRefreshH: 24,
		EPGRefreshH:     6,
	}, 300)
	if err == nil {
		t.Fatal("expected sync error")
	}

	snapshot := store.Current()
	if snapshot.Health.LastFailureUnix != 300 {
		t.Fatalf("expected failure timestamp, got %d", snapshot.Health.LastFailureUnix)
	}
}

func TestSyncM3UXMLTVBuildsFallbackCatalog(t *testing.T) {
	t.Parallel()

	store := cache.NewStore()
	service := NewService(Dependencies{Store: store, FetchURL: func(_ context.Context, rawURL string) ([]byte, error) {
		switch rawURL {
		case "https://dispatcharr.example.com/playlist.m3u":
			return []byte("#EXTM3U\n#EXTINF:-1 tvg-id=\"news.hd\",News HD\nhttps://dispatcharr.example.com/live/news-hd.m3u8\n"), nil
		case "https://dispatcharr.example.com/guide.xml":
			return []byte("<?xml version=\"1.0\"?><tv><channel id=\"news.hd\"><display-name>News HD</display-name></channel><programme start=\"20231114221320 +0000\" stop=\"20231114231320 +0000\" channel=\"news.hd\"><title>Morning News</title><desc>Top headlines.</desc></programme></tv>"), nil
		default:
			return nil, context.DeadlineExceeded
		}
	}})

	err := service.SyncNow(context.Background(), config.Settings{SourceMode: config.SourceModeM3UXMLTV, M3UURL: "https://dispatcharr.example.com/playlist.m3u", EPGXMLURL: "https://dispatcharr.example.com/guide.xml", ChannelRefreshH: 24, EPGRefreshH: 6}, 400)
	if err != nil {
		t.Fatalf("expected fallback sync success, got %v", err)
	}

	snapshot := store.Current()
	if len(snapshot.Catalog.Channels) != 1 || len(snapshot.Catalog.Programs) != 1 {
		t.Fatalf("unexpected fallback snapshot: %+v", snapshot)
	}
}
