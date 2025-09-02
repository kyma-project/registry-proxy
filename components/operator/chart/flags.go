package chart

import (
	"fmt"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/strvals"
)

type ImageReplace func(string) *flagsBuilder

type FlagsBuilder interface {
	Build() (map[string]interface{}, error)
	WithManagedByLabel(string) *flagsBuilder
	WithIstioInstalled(bool) *flagsBuilder
	WithImageRegistryProxy(string) *flagsBuilder
	WithImageConnection(string) *flagsBuilder
}

type flagsBuilder struct {
	flags map[string]interface{}
}

func NewFlagsBuilder() FlagsBuilder {
	return &flagsBuilder{
		flags: map[string]interface{}{},
	}
}

func (fb *flagsBuilder) Build() (map[string]interface{}, error) {
	flags := map[string]interface{}{}
	for key, value := range fb.flags {
		flag := fmt.Sprintf("%s=%v", key, value)
		err := strvals.ParseInto(flag, flags)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s flag", flag)
		}
	}
	return flags, nil
}

func (fb *flagsBuilder) WithManagedByLabel(managedBy string) *flagsBuilder {
	fb.flags["global.commonLabels.app\\.kubernetes\\.io/managed-by"] = managedBy
	return fb
}

func (fb *flagsBuilder) WithIstioInstalled(istioInstalled bool) *flagsBuilder {
	fb.flags["controllerManager.container.env.ISTIO_INSTALLED"] = fmt.Sprintf("\"%s\"", fmt.Sprintf("%t", istioInstalled))
	return fb
}

func (fb *flagsBuilder) WithImageRegistryProxy(image string) *flagsBuilder {
	fb.flags["global.images.registry_proxy"] = image
	return fb
}

func (fb *flagsBuilder) WithImageConnection(image string) *flagsBuilder {
	fb.flags["global.images.connection"] = image
	return fb
}
