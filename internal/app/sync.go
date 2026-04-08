package app

import (
	"context"
	"fmt"

	"github.com/relictiohosting/continuum-plugins/dispatcharr/internal/cache"
	"github.com/relictiohosting/continuum-plugins/dispatcharr/internal/config"
	"github.com/relictiohosting/continuum-plugins/dispatcharr/internal/mapping"
	"github.com/relictiohosting/continuum-plugins/dispatcharr/internal/matching"
	"github.com/relictiohosting/continuum-plugins/dispatcharr/internal/model"
	"github.com/relictiohosting/continuum-plugins/dispatcharr/internal/upstream/m3u"
	"github.com/relictiohosting/continuum-plugins/dispatcharr/internal/upstream/xmltv"
)

func (s *Service) SyncNow(ctx context.Context, settings config.Settings, nowUnix int64) error {
	if err := settings.Validate(); err != nil {
		return err
	}

	switch settings.SourceMode {
	case config.SourceModeXtream:
		client := s.xtreamFactory(settings.XtreamBaseURL, settings.XtreamUsername, settings.XtreamPassword)
		streams, err := client.LiveStreams(ctx)
		if err != nil {
			s.store.RecordFailure(nowUnix, err.Error())
			return err
		}

		channels := make([]model.Channel, 0, len(streams))
		programs := make([]model.Program, 0)
		for _, stream := range streams {
			channel := mapping.MapXtreamChannel(stream)
			channels = append(channels, channel)

			epg, err := client.ShortEPG(ctx, stream.StreamID)
			if err != nil {
				s.store.RecordFailure(nowUnix, err.Error())
				return err
			}
			for _, listing := range epg.EPGListings {
				programs = append(programs, mapping.MapXtreamProgram(channel.ID, listing))
			}
		}

		catalog := model.CatalogState{
			Source:   model.LiveTVSource(model.SourceModeXtream),
			Channels: channels,
			Programs: programs,
			Health:   model.SyncHealth{LastSuccessUnix: nowUnix},
		}
		state := cache.SnapshotFromCatalog(catalog)
		state.Health.LastSuccessUnix = nowUnix
		s.store.Replace(state)
		return nil
	case config.SourceModeM3UXMLTV:
		playlistData, err := s.fetchURL(ctx, settings.M3UURL)
		if err != nil {
			s.store.RecordFailure(nowUnix, err.Error())
			return err
		}
		xmltvData, err := s.fetchURL(ctx, settings.EPGXMLURL)
		if err != nil {
			s.store.RecordFailure(nowUnix, err.Error())
			return err
		}
		entries, err := m3u.Parse(playlistData)
		if err != nil {
			s.store.RecordFailure(nowUnix, err.Error())
			return err
		}
		doc, err := xmltv.Parse(xmltvData)
		if err != nil {
			s.store.RecordFailure(nowUnix, err.Error())
			return err
		}
		channels := make([]model.Channel, 0, len(entries))
		programs := make([]model.Program, 0)
		for _, entry := range entries {
			channel := mapping.MapM3UChannel(entry)
			channels = append(channels, channel)
			matchedChannel, ok := matching.Match(entry, doc)
			if !ok {
				continue
			}
			for _, programme := range doc.Programmes {
				if programme.Channel == matchedChannel.ID {
					programs = append(programs, mapping.MapXMLTVProgramme(channel.ID, programme))
				}
			}
		}
		catalog := model.CatalogState{Source: model.LiveTVSource(model.SourceModeM3UXMLTV), Channels: channels, Programs: programs, Health: model.SyncHealth{LastSuccessUnix: nowUnix}}
		state := cache.SnapshotFromCatalog(catalog)
		state.Health.LastSuccessUnix = nowUnix
		s.store.Replace(state)
		return nil
	default:
		return fmt.Errorf("source mode %q not implemented", settings.SourceMode)
	}
}
