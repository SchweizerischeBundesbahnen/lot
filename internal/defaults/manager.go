package defaults

import (
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var InitManager = func(options *manager.Options) (manager.Manager, error) {
	if options == nil {
		options = &manager.Options{}
	}

	m, err := manager.New(config.GetConfigOrDie(), *options)
	if err != nil {
		return nil, err
	}

	return m, nil
}
