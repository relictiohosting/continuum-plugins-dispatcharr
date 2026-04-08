package model

import "strconv"

type ProgramIdentity struct {
	UpstreamID string
	ChannelID  string
	Title      string
	StartUnix  int64
}

type Program struct {
	ID        string
	ChannelID string
	Title     string
	Summary   string
	StartUnix int64
	EndUnix   int64
}

func StableProgramID(identity ProgramIdentity) string {
	if normalize(identity.UpstreamID) != "" {
		return "program:" + normalize(identity.UpstreamID)
	}

	return "program:" + stableHash(identity.ChannelID+"|"+normalize(identity.Title)+"|"+strconv.FormatInt(identity.StartUnix, 10))
}
