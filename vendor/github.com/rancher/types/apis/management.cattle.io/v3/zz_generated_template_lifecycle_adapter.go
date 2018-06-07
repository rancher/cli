package v3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type TemplateLifecycle interface {
	Create(obj *Template) (*Template, error)
	Remove(obj *Template) (*Template, error)
	Updated(obj *Template) (*Template, error)
}

type templateLifecycleAdapter struct {
	lifecycle TemplateLifecycle
}

func (w *templateLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*Template))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *templateLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*Template))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *templateLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*Template))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewTemplateLifecycleAdapter(name string, clusterScoped bool, client TemplateInterface, l TemplateLifecycle) TemplateHandlerFunc {
	adapter := &templateLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *Template) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
