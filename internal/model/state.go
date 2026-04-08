package model

type SyncHealth struct {
	LastSuccessUnix int64
	LastFailureUnix int64
	LastError       string
}

type CatalogState struct {
	Source   Source
	Channels []Channel
	Programs []Program
	Health   SyncHealth
}
