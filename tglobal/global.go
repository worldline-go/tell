package tglobal

import (
	"sync"

	"go.opentelemetry.io/otel/sdk/metric"
)

var MetricViews = MetricViewsType{}

type MetricViewsType struct {
	store map[string][]metric.View
	mutex sync.RWMutex
}

func (g *MetricViewsType) Add(name string, v []metric.View) *MetricViewsType {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.store == nil {
		g.store = make(map[string][]metric.View)
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

func (g *MetricViewsType) GetViewGroup(name string) []metric.View {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	return g.store[name]
}

func (g *MetricViewsType) GetViews() []metric.View {
	if g.store == nil {
		return nil
	}

	g.mutex.RLock()
	defer g.mutex.RUnlock()

	views := make([]metric.View, 0, len(g.store))

	for k := range g.store {
		views = append(views, g.store[k]...)
	}

	return views
}
