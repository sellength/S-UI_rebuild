package service

import "s-ui/database/model"

const (
	ClientAccessLocal   = "local"
	ClientAccessCluster = "cluster"
	ClientAccessBoth    = "both"
)

func clientAllowsCluster(client model.Client) bool {
	return ClientAllowsCluster(client)
}

func ClientAllowsCluster(client model.Client) bool {
	return client.AccessScope == ClientAccessCluster
}

func ClientAllowsLocal(client model.Client) bool {
	return client.AccessScope == "" || client.AccessScope == ClientAccessLocal
}
