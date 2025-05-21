package chart

import (
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type FilterFunc func(unstructured.Unstructured) bool

func Uninstall(config *Config, filterFunc ...FilterFunc) error {
	spec, err := config.Cache.Get(config.Ctx, config.CacheKey)
	if err != nil {
		return fmt.Errorf("could not render manifest from chart: %s", err.Error())
	}

	objs, err := parseManifest(spec.Manifest)
	if err != nil {
		return fmt.Errorf("could not parse chart manifest: %s", err.Error())
	}

	err2 := uninstallObjects(config, objs, filterFunc...)
	if err2 != nil {
		return err2
	}

	return config.Cache.Delete(config.Ctx, config.CacheKey)
}

func uninstallObjects(config *Config, objs []unstructured.Unstructured, filterFunc ...FilterFunc) error {
	for i := range objs {
		u := objs[i]
		if !fitToFilters(u, filterFunc...) {
			continue
		}

		config.Log.Debugf("deleting %s %s", u.GetKind(), u.GetName())
		err := config.Cluster.Client.Delete(config.Ctx, &u)
		if k8serrors.IsNotFound(err) {
			config.Log.Debugf("deletion skipped for %s %s", u.GetKind(), u.GetName())
			continue
		}
		if err != nil {
			return fmt.Errorf("could not uninstall object %s/%s: %s", u.GetNamespace(), u.GetName(), err.Error())
		}
	}
	return nil
}

func UninstallResourcesByType(config *Config, resourceType string, filterFunc ...FilterFunc) (error, bool) {
	spec, err := config.Cache.Get(config.Ctx, config.CacheKey)
	if err != nil {
		return fmt.Errorf("could not render manifest from chart: %s", err.Error()), false
	}

	objs, err := parseManifest(spec.Manifest)
	if err != nil {
		return fmt.Errorf("could not parse chart manifest: %s", err.Error()), false
	}

	err2, done := uninstallResourcesByType(config, objs, resourceType, filterFunc...)
	if err2 != nil {
		return err2, false
	}

	return nil, done
}

func uninstallResourcesByType(config *Config, objs []unstructured.Unstructured, resourceType string, filterFunc ...FilterFunc) (error, bool) {
	done := true
	for i := range objs {
		u := objs[i]
		if !fitToFilters(u, filterFunc...) {
			continue
		}
		if u.GetKind() != resourceType {
			continue
		}

		config.Log.Debugf("deleting %s %s", u.GetKind(), u.GetName())
		err := config.Cluster.Client.Delete(config.Ctx, &u)
		if k8serrors.IsNotFound(err) {
			config.Log.Debugf("deletion skipped for %s %s", u.GetKind(), u.GetName())
			continue
		}
		if err != nil {
			return fmt.Errorf("could not uninstall object %s/%s: %s", u.GetNamespace(), u.GetName(), err.Error()), false
		}
		done = false
	}
	return nil, done
}

func WithoutCRDFilter(u unstructured.Unstructured) bool {
	return !isCRD(u)
}

func fitToFilters(u unstructured.Unstructured, filterFunc ...FilterFunc) bool {
	for _, fn := range filterFunc {
		if !fn(u) {
			return false
		}
	}

	return true
}
