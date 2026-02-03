package flags

import (
	"fmt"

	"github.com/kyma-project/manager-toolkit/installation/chart"
)

type ImageReplace func(string) *Builder

type Builder struct {
	chart.FlagsBuilder
}

func NewBuilder() *Builder {
	return &Builder{
		FlagsBuilder: chart.NewFlagsBuilder(),
	}
}

func (fb *Builder) WithManagedByLabel(managedBy string) *Builder {
	fb.With("global.commonLabels.managedBy", managedBy)
	return fb
}

func (fb *Builder) WithIstioInstalled(istioInstalled bool) *Builder {
	fb.With("controllerManager.container.env.ISTIO_INSTALLED", fmt.Sprintf("\"%s\"", fmt.Sprintf("%t", istioInstalled)))
	return fb
}

func (fb *Builder) WithProxyURL(proxyURL string) *Builder {
	fb.With("global.proxy.url", proxyURL)
	return fb
}

func (fb *Builder) WithProxyLocationID(locationID string) *Builder {
	fb.With("global.proxy.locationID", locationID)
	return fb
}

func (fb *Builder) WithImageRegistryProxy(image string) *Builder {
	fb.With("global.images.registry_proxy", image)
	return fb
}

func (fb *Builder) WithImageConnection(image string) *Builder {
	fb.With("global.images.connection", image)
	return fb
}
