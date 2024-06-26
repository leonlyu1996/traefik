package metrics

import (
	"errors"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/multi"
)

const defaultMetricsPrefix = "traefik"

// Registry has to implemented by any system that wants to monitor and expose metrics.
type Registry interface {
	// IsEpEnabled shows whether metrics instrumentation is enabled on entry points.
	IsEpEnabled() bool
	// IsRouterEnabled shows whether metrics instrumentation is enabled on routers.
	IsRouterEnabled() bool
	// IsSvcEnabled shows whether metrics instrumentation is enabled on services.
	IsSvcEnabled() bool

	// server metrics

	ConfigReloadsCounter() metrics.Counter
	LastConfigReloadSuccessGauge() metrics.Gauge
	OpenConnectionsGauge() metrics.Gauge

	// TLS

	TLSCertsNotAfterTimestampGauge() metrics.Gauge

	// entry point metrics

	EntryPointReqsCounter() CounterWithHeaders
	EntryPointReqsTLSCounter() CounterWithHeaders
	EntryPointReqDurationHistogram() ScalableHistogramWithHeaders
	EntryPointReqsBytesCounter() CounterWithHeaders
	EntryPointRespsBytesCounter() CounterWithHeaders

	// router metrics

	RouterReqsCounter() CounterWithHeaders
	RouterReqsTLSCounter() CounterWithHeaders
	RouterReqDurationHistogram() ScalableHistogramWithHeaders
	RouterReqsBytesCounter() CounterWithHeaders
	RouterRespsBytesCounter() CounterWithHeaders

	// service metrics

	ServiceReqsCounter() CounterWithHeaders
	ServiceReqsTLSCounter() CounterWithHeaders
	ServiceReqDurationHistogram() ScalableHistogramWithHeaders
	ServiceRetriesCounter() metrics.Counter
	ServiceServerUpGauge() metrics.Gauge
	ServiceReqsBytesCounter() CounterWithHeaders
	ServiceRespsBytesCounter() CounterWithHeaders
}

// NewVoidRegistry is a noop implementation of metrics.Registry.
// It is used to avoid nil checking in components that do metric collections.
func NewVoidRegistry() Registry {
	return NewMultiRegistry([]Registry{})
}

// NewMultiRegistry is an implementation of metrics.Registry that wraps multiple registries.
// It handles the case when a registry hasn't registered some metric and returns nil.
// This allows for feature disparity between the different metric implementations.
func NewMultiRegistry(registries []Registry) Registry {
	var configReloadsCounter []metrics.Counter
	var lastConfigReloadSuccessGauge []metrics.Gauge
	var openConnectionsGauge []metrics.Gauge
	var tlsCertsNotAfterTimestampGauge []metrics.Gauge
	var entryPointReqsCounter []CounterWithHeaders
	var entryPointReqsTLSCounter []CounterWithHeaders
	var entryPointReqDurationHistogram []ScalableHistogramWithHeaders
	var entryPointReqsBytesCounter []CounterWithHeaders
	var entryPointRespsBytesCounter []CounterWithHeaders
	var routerReqsCounter []CounterWithHeaders
	var routerReqsTLSCounter []CounterWithHeaders
	var routerReqDurationHistogram []ScalableHistogramWithHeaders
	var routerReqsBytesCounter []CounterWithHeaders
	var routerRespsBytesCounter []CounterWithHeaders
	var serviceReqsCounter []CounterWithHeaders
	var serviceReqsTLSCounter []CounterWithHeaders
	var serviceReqDurationHistogram []ScalableHistogramWithHeaders
	var serviceRetriesCounter []metrics.Counter
	var serviceServerUpGauge []metrics.Gauge
	var serviceReqsBytesCounter []CounterWithHeaders
	var serviceRespsBytesCounter []CounterWithHeaders

	for _, r := range registries {
		if r.ConfigReloadsCounter() != nil {
			configReloadsCounter = append(configReloadsCounter, r.ConfigReloadsCounter())
		}
		if r.LastConfigReloadSuccessGauge() != nil {
			lastConfigReloadSuccessGauge = append(lastConfigReloadSuccessGauge, r.LastConfigReloadSuccessGauge())
		}
		if r.OpenConnectionsGauge() != nil {
			openConnectionsGauge = append(openConnectionsGauge, r.OpenConnectionsGauge())
		}
		if r.TLSCertsNotAfterTimestampGauge() != nil {
			tlsCertsNotAfterTimestampGauge = append(tlsCertsNotAfterTimestampGauge, r.TLSCertsNotAfterTimestampGauge())
		}
		if r.EntryPointReqsCounter() != nil {
			entryPointReqsCounter = append(entryPointReqsCounter, r.EntryPointReqsCounter())
		}
		if r.EntryPointReqsTLSCounter() != nil {
			entryPointReqsTLSCounter = append(entryPointReqsTLSCounter, r.EntryPointReqsTLSCounter())
		}
		if r.EntryPointReqDurationHistogram() != nil {
			entryPointReqDurationHistogram = append(entryPointReqDurationHistogram, r.EntryPointReqDurationHistogram())
		}
		if r.EntryPointReqsBytesCounter() != nil {
			entryPointReqsBytesCounter = append(entryPointReqsBytesCounter, r.EntryPointReqsBytesCounter())
		}
		if r.EntryPointRespsBytesCounter() != nil {
			entryPointRespsBytesCounter = append(entryPointRespsBytesCounter, r.EntryPointRespsBytesCounter())
		}
		if r.RouterReqsCounter() != nil {
			routerReqsCounter = append(routerReqsCounter, r.RouterReqsCounter())
		}
		if r.RouterReqsTLSCounter() != nil {
			routerReqsTLSCounter = append(routerReqsTLSCounter, r.RouterReqsTLSCounter())
		}
		if r.RouterReqDurationHistogram() != nil {
			routerReqDurationHistogram = append(routerReqDurationHistogram, r.RouterReqDurationHistogram())
		}
		if r.RouterReqsBytesCounter() != nil {
			routerReqsBytesCounter = append(routerReqsBytesCounter, r.RouterReqsBytesCounter())
		}
		if r.RouterRespsBytesCounter() != nil {
			routerRespsBytesCounter = append(routerRespsBytesCounter, r.RouterRespsBytesCounter())
		}
		if r.ServiceReqsCounter() != nil {
			serviceReqsCounter = append(serviceReqsCounter, r.ServiceReqsCounter())
		}
		if r.ServiceReqsTLSCounter() != nil {
			serviceReqsTLSCounter = append(serviceReqsTLSCounter, r.ServiceReqsTLSCounter())
		}
		if r.ServiceReqDurationHistogram() != nil {
			serviceReqDurationHistogram = append(serviceReqDurationHistogram, r.ServiceReqDurationHistogram())
		}
		if r.ServiceRetriesCounter() != nil {
			serviceRetriesCounter = append(serviceRetriesCounter, r.ServiceRetriesCounter())
		}
		if r.ServiceServerUpGauge() != nil {
			serviceServerUpGauge = append(serviceServerUpGauge, r.ServiceServerUpGauge())
		}
		if r.ServiceReqsBytesCounter() != nil {
			serviceReqsBytesCounter = append(serviceReqsBytesCounter, r.ServiceReqsBytesCounter())
		}
		if r.ServiceRespsBytesCounter() != nil {
			serviceRespsBytesCounter = append(serviceRespsBytesCounter, r.ServiceRespsBytesCounter())
		}
	}

	return &standardRegistry{
		epEnabled:                      len(entryPointReqsCounter) > 0 || len(entryPointReqDurationHistogram) > 0,
		svcEnabled:                     len(serviceReqsCounter) > 0 || len(serviceReqDurationHistogram) > 0 || len(serviceRetriesCounter) > 0 || len(serviceServerUpGauge) > 0,
		routerEnabled:                  len(routerReqsCounter) > 0 || len(routerReqDurationHistogram) > 0,
		configReloadsCounter:           multi.NewCounter(configReloadsCounter...),
		lastConfigReloadSuccessGauge:   multi.NewGauge(lastConfigReloadSuccessGauge...),
		openConnectionsGauge:           multi.NewGauge(openConnectionsGauge...),
		tlsCertsNotAfterTimestampGauge: multi.NewGauge(tlsCertsNotAfterTimestampGauge...),
		entryPointReqsCounter:          NewMultiCounterWithHeaders(entryPointReqsCounter...),
		entryPointReqsTLSCounter:       NewMultiCounterWithHeaders(entryPointReqsTLSCounter...),
		entryPointReqDurationHistogram: NewMultiScalableHistogramWithHeaders(entryPointReqDurationHistogram...),
		entryPointReqsBytesCounter:     NewMultiCounterWithHeaders(entryPointReqsBytesCounter...),
		entryPointRespsBytesCounter:    NewMultiCounterWithHeaders(entryPointRespsBytesCounter...),
		routerReqsCounter:              NewMultiCounterWithHeaders(routerReqsCounter...),
		routerReqsTLSCounter:           NewMultiCounterWithHeaders(routerReqsTLSCounter...),
		routerReqDurationHistogram:     NewMultiScalableHistogramWithHeaders(routerReqDurationHistogram...),
		routerReqsBytesCounter:         NewMultiCounterWithHeaders(routerReqsBytesCounter...),
		routerRespsBytesCounter:        NewMultiCounterWithHeaders(routerRespsBytesCounter...),
		serviceReqsCounter:             NewMultiCounterWithHeaders(serviceReqsCounter...),
		serviceReqsTLSCounter:          NewMultiCounterWithHeaders(serviceReqsTLSCounter...),
		serviceReqDurationHistogram:    NewMultiScalableHistogramWithHeaders(serviceReqDurationHistogram...),
		serviceRetriesCounter:          multi.NewCounter(serviceRetriesCounter...),
		serviceServerUpGauge:           multi.NewGauge(serviceServerUpGauge...),
		serviceReqsBytesCounter:        NewMultiCounterWithHeaders(serviceReqsBytesCounter...),
		serviceRespsBytesCounter:       NewMultiCounterWithHeaders(serviceRespsBytesCounter...),
	}
}

type standardRegistry struct {
	epEnabled                      bool
	routerEnabled                  bool
	svcEnabled                     bool
	configReloadsCounter           metrics.Counter
	lastConfigReloadSuccessGauge   metrics.Gauge
	openConnectionsGauge           metrics.Gauge
	tlsCertsNotAfterTimestampGauge metrics.Gauge
	entryPointReqsCounter          CounterWithHeaders
	entryPointReqsTLSCounter       CounterWithHeaders
	entryPointReqDurationHistogram ScalableHistogramWithHeaders
	entryPointReqsBytesCounter     CounterWithHeaders
	entryPointRespsBytesCounter    CounterWithHeaders
	routerReqsCounter              CounterWithHeaders
	routerReqsTLSCounter           CounterWithHeaders
	routerReqDurationHistogram     ScalableHistogramWithHeaders
	routerReqsBytesCounter         CounterWithHeaders
	routerRespsBytesCounter        CounterWithHeaders
	serviceReqsCounter             CounterWithHeaders
	serviceReqsTLSCounter          CounterWithHeaders
	serviceReqDurationHistogram    ScalableHistogramWithHeaders
	serviceRetriesCounter          metrics.Counter
	serviceServerUpGauge           metrics.Gauge
	serviceReqsBytesCounter        CounterWithHeaders
	serviceRespsBytesCounter       CounterWithHeaders
}

func (r *standardRegistry) IsEpEnabled() bool {
	return r.epEnabled
}

func (r *standardRegistry) IsRouterEnabled() bool {
	return r.routerEnabled
}

func (r *standardRegistry) IsSvcEnabled() bool {
	return r.svcEnabled
}

func (r *standardRegistry) ConfigReloadsCounter() metrics.Counter {
	return r.configReloadsCounter
}

func (r *standardRegistry) LastConfigReloadSuccessGauge() metrics.Gauge {
	return r.lastConfigReloadSuccessGauge
}

func (r *standardRegistry) OpenConnectionsGauge() metrics.Gauge {
	return r.openConnectionsGauge
}

func (r *standardRegistry) TLSCertsNotAfterTimestampGauge() metrics.Gauge {
	return r.tlsCertsNotAfterTimestampGauge
}

func (r *standardRegistry) EntryPointReqsCounter() CounterWithHeaders {
	return r.entryPointReqsCounter
}

func (r *standardRegistry) EntryPointReqsTLSCounter() CounterWithHeaders {
	return r.entryPointReqsTLSCounter
}

func (r *standardRegistry) EntryPointReqDurationHistogram() ScalableHistogramWithHeaders {
	return r.entryPointReqDurationHistogram
}

func (r *standardRegistry) EntryPointReqsBytesCounter() CounterWithHeaders {
	return r.entryPointReqsBytesCounter
}

func (r *standardRegistry) EntryPointRespsBytesCounter() CounterWithHeaders {
	return r.entryPointRespsBytesCounter
}

func (r *standardRegistry) RouterReqsCounter() CounterWithHeaders {
	return r.routerReqsCounter
}

func (r *standardRegistry) RouterReqsTLSCounter() CounterWithHeaders {
	return r.routerReqsTLSCounter
}

func (r *standardRegistry) RouterReqDurationHistogram() ScalableHistogramWithHeaders {
	return r.routerReqDurationHistogram
}

func (r *standardRegistry) RouterReqsBytesCounter() CounterWithHeaders {
	return r.routerReqsBytesCounter
}

func (r *standardRegistry) RouterRespsBytesCounter() CounterWithHeaders {
	return r.routerRespsBytesCounter
}

func (r *standardRegistry) ServiceReqsCounter() CounterWithHeaders {
	return r.serviceReqsCounter
}

func (r *standardRegistry) ServiceReqsTLSCounter() CounterWithHeaders {
	return r.serviceReqsTLSCounter
}

func (r *standardRegistry) ServiceReqDurationHistogram() ScalableHistogramWithHeaders {
	return r.serviceReqDurationHistogram
}

func (r *standardRegistry) ServiceRetriesCounter() metrics.Counter {
	return r.serviceRetriesCounter
}

func (r *standardRegistry) ServiceServerUpGauge() metrics.Gauge {
	return r.serviceServerUpGauge
}

func (r *standardRegistry) ServiceReqsBytesCounter() CounterWithHeaders {
	return r.serviceReqsBytesCounter
}

func (r *standardRegistry) ServiceRespsBytesCounter() CounterWithHeaders {
	return r.serviceRespsBytesCounter
}

// ScalableHistogram is a Histogram with a predefined time unit,
// used when producing observations without explicitly setting the observed value.
type ScalableHistogram interface {
	With(labelValues ...string) ScalableHistogram
	Observe(v float64)
	ObserveFromStart(start time.Time)
}

// HistogramWithScale is a histogram that will convert its observed value to the specified unit.
type HistogramWithScale struct {
	histogram metrics.Histogram
	unit      time.Duration
}

// With implements ScalableHistogram.
func (s *HistogramWithScale) With(labelValues ...string) ScalableHistogram {
	h, _ := NewHistogramWithScale(s.histogram.With(labelValues...), s.unit)
	return h
}

// ObserveFromStart implements ScalableHistogram.
func (s *HistogramWithScale) ObserveFromStart(start time.Time) {
	if s.unit <= 0 {
		return
	}

	d := float64(time.Since(start).Nanoseconds()) / float64(s.unit)
	if d < 0 {
		d = 0
	}
	s.histogram.Observe(d)
}

// Observe implements ScalableHistogram.
func (s *HistogramWithScale) Observe(v float64) {
	s.histogram.Observe(v)
}

// NewHistogramWithScale returns a ScalableHistogram. It returns an error if the given unit is <= 0.
func NewHistogramWithScale(histogram metrics.Histogram, unit time.Duration) (ScalableHistogram, error) {
	if unit <= 0 {
		return nil, errors.New("invalid time unit")
	}
	return &HistogramWithScale{
		histogram: histogram,
		unit:      unit,
	}, nil
}

// MultiHistogram collects multiple individual histograms and treats them as a unit.
type MultiHistogram []ScalableHistogram

// ObserveFromStart implements ScalableHistogram.
func (h MultiHistogram) ObserveFromStart(start time.Time) {
	for _, histogram := range h {
		histogram.ObserveFromStart(start)
	}
}

// Observe implements ScalableHistogram.
func (h MultiHistogram) Observe(v float64) {
	for _, histogram := range h {
		histogram.Observe(v)
	}
}

// With implements ScalableHistogram.
func (h MultiHistogram) With(labelValues ...string) ScalableHistogram {
	next := make(MultiHistogram, len(h))
	for i := range h {
		next[i] = h[i].With(labelValues...)
	}
	return next
}
