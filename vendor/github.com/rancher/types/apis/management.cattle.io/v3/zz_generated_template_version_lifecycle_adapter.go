package v3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type TemplateVersionLifecycle interface {
	Create(obj *TemplateVersion) (*TemplateVersion, error)
	Remove(obj *TemplateVersion) (*TemplateVersion, error)
	Updated(obj *TemplateVersion) (*TemplateVersion, error)
}

type templateVersionLifecycleAdapter struct {
	lifecycle TemplateVersionLifecycle
}

func (w *templateVersionLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*TemplateVersion))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *templateVersionLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*TemplateVersion))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *templateVersionLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*TemplateVersion))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewTemplateVersionLifecycleAdapter(name string, clusterScoped bool, client TemplateVersionInterface, l TemplateVersionLifecycle) TemplateVersionHandlerFunc {
	adapter := &templateVersionLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *TemplateVersion) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
