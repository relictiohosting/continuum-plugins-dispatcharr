package cache

import "github.com/relictiohosting/continuum-plugins/dispatcharr/internal/model"

func SnapshotFromCatalog(catalog model.CatalogState) Snapshot {
	return Snapshot{Catalog: catalog, Health: catalog.Health}
}
