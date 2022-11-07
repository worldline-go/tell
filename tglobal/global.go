package tglobal

import (
	"sync"

	"go.opentelemetry.io/otel/sdk/metric/view"
)

var MetricViews = MetricViewsType{}

type MetricViewsType struct {
	store map[string][]view.View
	mutex sync.RWMutex
}

func (g *MetricViewsType) Add(name string, v []view.View) *MetricViewsType {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.store == nil {
		g.store = make(map[string][]view.View)
	}

	g.store[name] = v

	return g
}

func (g *MetricViewsType) Delete(name string) *MetricViewsType {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.store != nil {
		delete(g.store, name)
	}

	return g
}

func (g *MetricViewsType) GetViewGroup(name string) []view.View {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	return g.store[name]
}

func (g *MetricViewsType) GetViews() []view.View {
	if g.store == nil {
		return nil
	}

	g.mutex.RLock()
	defer g.mutex.RUnlock()

	views := make([]view.View, 0, len(g.store))

	for k := range g.store {
		views = append(views, g.store[k]...)
	}

	return views
}
