/*
Copyright DNITSCH WTFPL
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	internalinterfaces "github.com/dnitsch/reststrategy/apis/reststrategy/generated/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// RestStrategies returns a RestStrategyInformer.
	RestStrategies() RestStrategyInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// RestStrategies returns a RestStrategyInformer.
func (v *version) RestStrategies() RestStrategyInformer {
	return &restStrategyInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
