package config

import (
	"sync"

	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/logging"
)

const (
	ConfigVarsName = "sample-source-config-vars"
)

type SampleConfigVars struct {
	configVars map[string]string
	m          sync.RWMutex
}

func NewSampleConfigVars(cm *corev1.ConfigMap) (*SampleConfigVars, error) {
	return &SampleConfigVars{
		configVars: cm.DeepCopy().Data,
	}, nil
}

func (scv *SampleConfigVars) Reset(new *SampleConfigVars) {
	scv.m.Lock()
	defer scv.m.Unlock()
	scv.configVars = new.configVars
}

func (scv *SampleConfigVars) GetConfigVars() map[string]string {
	scv.m.RLock()
	defer scv.m.RUnlock()
	return scv.configVars
}

// +k8s:deepcopy-gen=false
type Store struct {
	*configmap.UntypedStore
}

func NewStore(ctx context.Context, onAfterStore ...func(name string, value *SampleConfigVars)) *Store {
	store := &Store{
		UntypedStore: configmap.NewUntypedStore(
			"sample-source-config-vars",
			logging.FromContext(ctx).Named(ConfigVarsName),
			configmap.Constructors{
				ConfigVarsName: NewSampleConfigVars,
			},
			func(name string, value interface{}) {
				for _, f := range onAfterStore {
					f(name, value.(*SampleConfigVars))
				}
			},
		),
	}

	return store
}
