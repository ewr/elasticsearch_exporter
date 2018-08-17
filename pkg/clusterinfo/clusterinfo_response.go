package clusterinfo

import (
	"github.com/blang/semver"
	"time"
)

// ClusterInfoResponse is the cluster info retrievable from the / endpoint
type Response struct {
	Name        string      `json:"name"`
	ClusterName string      `json:"cluster_name"`
	ClusterUUID string      `json:"cluster_uuid"`
	Version     VersionInfo `json:"version"`
	Tagline     string      `json:"tagline"`
}

// ClusterVersionInfo is the version info retrievable from the / endpoint, embedded in Response
type VersionInfo struct {
	Number        semver.Version `json:"number"`
	BuildHash     string         `json:"build_hash"`
	BuildDate     time.Time      `json:"build_date"`
	BuildSnapshot bool           `json:"build_snapshot"`
	LuceneVersion semver.Version `json:"lucene_version"`
}
