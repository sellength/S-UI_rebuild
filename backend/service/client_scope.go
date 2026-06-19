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
	// 本地代理账号功能已去除，所有账号均作为集群（分布式）代理账号处理。
	return true
}

func ClientAllowsLocal(client model.Client) bool {
	// 本地代理账号功能已去除，不再支持本地代理
	return false
}
