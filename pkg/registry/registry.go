package registry

import "go-zuul/pkg/model"

type Registry interface {
	GetAppClusters() []model.AppCluster
}
