package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	pluginv1 "github.com/ContinuumApp/continuum-plugin-sdk/pkg/pluginproto/continuum/plugin/v1"
	"github.com/relictiohosting/continuum-plugins/dispatcharr/internal/cache"
	"github.com/relictiohosting/continuum-plugins/dispatcharr/internal/config"
	"github.com/relictiohosting/continuum-plugins/dispatcharr/internal/model"
	"github.com/relictiohosting/continuum-plugins/dispatcharr/internal/upstream/xtream"
)

type HTTPRoutesServer struct {
	pluginv1.UnimplementedHttpRoutesServer
	store            *cache.Store
	settingsProvider func() config.Settings
}

func NewHTTPRoutesServer(store *cache.Store) *HTTPRoutesServer {
	return &HTTPRoutesServer{store: store}
}

func NewHTTPRoutesServerWithSettings(store *cache.Store, settingsProvider func() config.Settings) *HTTPRoutesServer {
	return &HTTPRoutesServer{store: store, settingsProvider: settingsProvider}
}

type ChannelsPayload struct {
	SourceName string          `json:"sourceName"`
	Channels   []model.Channel `json:"channels"`
}

type GuidePayload struct {
	Programs []model.Program `json:"programs"`
}

func (s *HTTPRoutesServer) Handle(_ context.Context, request *pluginv1.HandleHTTPRequest) (*pluginv1.HandleHTTPResponse, error) {
	switch request.GetPath() {
	case "/dispatcharr/status":
		return s.respondJSON(http.StatusOK, BuildHealthPayload(s.store.Current()))
	case "/dispatcharr/channels":
		snapshot := s.store.Current()
		return s.respondJSON(http.StatusOK, ChannelsPayload{
			SourceName: snapshot.Catalog.Source.Name,
			Channels:   snapshot.Catalog.Channels,
		})
	case "/dispatcharr/guide":
		channelID := queryValue(request, "channel_id")
		programs := programsForChannel(s.store.Current().Catalog.Programs, channelID)
		sort.Slice(programs, func(i, j int) bool {
			return programs[i].StartUnix < programs[j].StartUnix
		})
		return s.respondJSON(http.StatusOK, GuidePayload{Programs: programs})
	case "/dispatcharr/stream":
		channelID := queryValue(request, "channel_id")
		if strings.TrimSpace(channelID) == "" {
			return textResponse(http.StatusBadRequest, "missing channel_id query parameter"), nil
		}

		streamURL, err := s.resolveStreamURL(channelID)
		if err != nil {
			return textResponse(http.StatusNotFound, err.Error()), nil
		}

		return &pluginv1.HandleHTTPResponse{
			StatusCode: http.StatusFound,
			Headers: map[string]string{
				"location": streamURL,
			},
		}, nil
	case "/dispatcharr/player":
		return &pluginv1.HandleHTTPResponse{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"content-type": "text/html; charset=utf-8",
			},
			Body: []byte(playerPageHTML),
		}, nil
	default:
		return textResponse(http.StatusNotFound, "route not found"), nil
	}
}

func (s *HTTPRoutesServer) respondJSON(status int, value any) (*pluginv1.HandleHTTPResponse, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return &pluginv1.HandleHTTPResponse{
		StatusCode: int32(status),
		Headers: map[string]string{
			"content-type": "application/json",
		},
		Body: payload,
	}, nil
}

func (s *HTTPRoutesServer) resolveStreamURL(channelID string) (string, error) {
	snapshot := s.store.Current()
	for _, channel := range snapshot.Catalog.Channels {
		if channel.ID != channelID {
			continue
		}
		if strings.TrimSpace(channel.StreamURL) != "" {
			return channel.StreamURL, nil
		}
		if strings.HasPrefix(channel.ID, "xtream:") && s.settingsProvider != nil {
			streamID, err := strconv.ParseInt(strings.TrimPrefix(channel.ID, "xtream:"), 10, 64)
			if err != nil {
				return "", fmt.Errorf("invalid xtream channel id")
			}
			settings := s.settingsProvider()
			client := xtream.NewClient(settings.XtreamBaseURL, settings.XtreamUsername, settings.XtreamPassword)
			streamURL := client.ResolveLiveStreamURL(streamID)
			if strings.TrimSpace(streamURL) == "" {
				return "", fmt.Errorf("unable to resolve stream url")
			}
			return streamURL, nil
		}
		return "", fmt.Errorf("stream url unavailable for channel")
	}
	return "", fmt.Errorf("channel not found")
}

func programsForChannel(programs []model.Program, channelID string) []model.Program {
	if strings.TrimSpace(channelID) == "" {
		return append([]model.Program(nil), programs...)
	}
	filtered := make([]model.Program, 0, len(programs))
	for _, program := range programs {
		if program.ChannelID == channelID {
			filtered = append(filtered, program)
		}
	}
	return filtered
}

func queryValue(request *pluginv1.HandleHTTPRequest, key string) string {
	query := request.GetQuery()
	if query == nil {
		return ""
	}
	value := query.AsMap()[key]
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func textResponse(status int, message string) *pluginv1.HandleHTTPResponse {
	return &pluginv1.HandleHTTPResponse{
		StatusCode: int32(status),
		Headers: map[string]string{
			"content-type": "text/plain; charset=utf-8",
		},
		Body: []byte(message),
	}
}

const playerPageHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Dispatcharr Live TV</title>
    <style>
      :root { color-scheme: light; }
      body { margin: 0; font-family: system-ui, sans-serif; background: #0d1117; color: #e6edf3; }
      main { display: grid; grid-template-columns: 320px 1fr; min-height: 100vh; }
      aside { border-right: 1px solid #30363d; padding: 12px; overflow: auto; }
      section { padding: 12px; }
      h1 { margin: 0 0 12px; font-size: 1rem; }
      button.channel { display: block; width: 100%; text-align: left; margin: 0 0 8px; padding: 10px; border: 1px solid #30363d; border-radius: 8px; background: #161b22; color: #e6edf3; cursor: pointer; }
      button.channel:hover { background: #1f2937; }
      video { width: 100%; max-height: 62vh; background: #000; border-radius: 8px; }
      ul { margin: 10px 0 0; padding-left: 18px; }
      @media (max-width: 900px) { main { grid-template-columns: 1fr; } aside { border-right: 0; border-bottom: 1px solid #30363d; max-height: 40vh; } }
    </style>
  </head>
  <body>
    <main>
      <aside>
        <h1>Channels</h1>
        <div id="channels">Loading…</div>
      </aside>
      <section>
        <video id="player" controls autoplay playsinline></video>
        <h2 id="now">Guide</h2>
        <ul id="guide"></ul>
      </section>
    </main>
    <script>
      const path = window.location.pathname;
      const base = path.endsWith("/dispatcharr/player") ? path.slice(0, -"/dispatcharr/player".length) : "";

      function route(url) {
        return base + url;
      }

      async function loadChannels() {
        const response = await fetch(route("/dispatcharr/channels"), { credentials: "include" });
        if (!response.ok) throw new Error("failed to load channels");
        return response.json();
      }

      async function loadGuide(channelID) {
        const response = await fetch(route("/dispatcharr/guide?channel_id=" + encodeURIComponent(channelID)), { credentials: "include" });
        if (!response.ok) return { programs: [] };
        return response.json();
      }

      function setGuide(programs) {
        const guide = document.getElementById("guide");
        guide.innerHTML = "";
        for (const program of programs.slice(0, 12)) {
          const li = document.createElement("li");
          const start = program.startUnix ? new Date(program.startUnix * 1000).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }) : "--:--";
          li.textContent = start + " " + (program.title || "Untitled");
          guide.appendChild(li);
        }
      }

      function playChannel(channel) {
        const player = document.getElementById("player");
        player.src = route("/dispatcharr/stream?channel_id=" + encodeURIComponent(channel.id));
        const now = document.getElementById("now");
        now.textContent = "Guide: " + channel.name;
      }

      async function boot() {
        const channelsRoot = document.getElementById("channels");
        try {
          const data = await loadChannels();
          channelsRoot.innerHTML = "";
          if (!data.channels || data.channels.length === 0) {
            channelsRoot.textContent = "No channels available";
            return;
          }
          for (const channel of data.channels) {
            const btn = document.createElement("button");
            btn.className = "channel";
            btn.textContent = (channel.number ? channel.number + "  " : "") + channel.name;
            btn.onclick = async () => {
              playChannel(channel);
              const guide = await loadGuide(channel.id);
              setGuide(guide.programs || []);
            };
            channelsRoot.appendChild(btn);
          }
          playChannel(data.channels[0]);
          const guide = await loadGuide(data.channels[0].id);
          setGuide(guide.programs || []);
        } catch (error) {
          channelsRoot.textContent = "Unable to load channels";
        }
      }

      boot();
    </script>
  </body>
</html>`
