package defaults

import "time"

const (
	SyncInterval                   = 5 * time.Second
	RetryInterval                  = 10 * time.Second
	ConflictInterval               = 1 * time.Second
	ConnectorsFinalizerName        = "connectors.cos.bf2.dev/finalizer"
	ConnectorClustersFinalizerName = "connectorclusters.cos.bf2.dev/finalizer"
)
