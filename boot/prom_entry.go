// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkgin

import (
	"context"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-prom"
	"strings"
)

var (
	// Why 1608? It is the year of first telescope was invented
	defaultPort = uint64(1608)
	defaultPath = "/metrics"
)

// Boot config which is for prom entry.
//
// 1: Path: PromEntry path, /metrics is default value.
// 2: Enabled: Enable prom entry.
// 3: Pusher.Enabled: Enable pushgateway pusher.
// 4: Pusher.IntervalMS: Interval of pushing metrics to remote pushgateway in milliseconds.
// 5: Pusher.JobName: Job name would be attached as label while pushing to remote pushgateway.
// 6: Pusher.RemoteAddress: Pushgateway address, could be form of http://x.x.x.x or x.x.x.x
// 7: Pusher.BasicAuth: Basic auth used to interact with remote pushgateway.
// 8: Pusher.Cert.Ref: Reference of rkentry.CertEntry.
// 9: Cert.Ref: Reference of rkentry.CertEntry.
type BootConfigProm struct {
	Path    string `yaml:"path"`
	Enabled bool   `yaml:"enabled"`
	Pusher  struct {
		Enabled       bool   `yaml:"enabled"`
		IntervalMS    int64  `yaml:"intervalMS"`
		JobName       string `yaml:"jobName"`
		RemoteAddress string `yaml:"remoteAddress"`
		BasicAuth     string `yaml:"basicAuth"`
		Cert          struct {
			Ref string `yaml:"ref"`
		} `yaml:"cert"`
	} `yaml:"pusher"`
	Cert struct {
		Ref string `yaml:"ref"`
	} `yaml:"cert"`
}

// Prometheus entry which implements rkentry.Entry.
//
// 1: Pusher            Periodic pushGateway pusher
// 2: ZapLoggerEntry    rkentry.ZapLoggerEntry
// 3: EventLoggerEntry  rkentry.EventLoggerEntry
// 4: Port              Exposed port by prom entry
// 5: Path              Exposed path by prom entry
// 6: Registry          Prometheus registry
// 7: Registerer        Prometheus registerer
// 8: Gatherer          Prometheus gatherer
// 9: CertStore         rkentry.CertStore
type PromEntry struct {
	Pusher           *rkprom.PushGatewayPusher
	entryName        string
	entryType        string
	ZapLoggerEntry   *rkentry.ZapLoggerEntry
	EventLoggerEntry *rkentry.EventLoggerEntry
	Port             uint64
	Path             string
	Registry         *prometheus.Registry
	Registerer       prometheus.Registerer
	Gatherer         prometheus.Gatherer
	CertStore        *rkentry.CertStore
}

// Prom entry option used while initializing prom entry via code
type PromEntryOption func(*PromEntry)

// Port of prom entry
func WithPortProm(port uint64) PromEntryOption {
	return func(entry *PromEntry) {
		entry.Port = port
	}
}

// Path of prom entry
func WithPathProm(path string) PromEntryOption {
	return func(entry *PromEntry) {
		entry.Path = path
	}
}

// rkentry.ZapLoggerEntry of prom entry
func WithZapLoggerEntryProm(zapLoggerEntry *rkentry.ZapLoggerEntry) PromEntryOption {
	return func(entry *PromEntry) {
		entry.ZapLoggerEntry = zapLoggerEntry
	}
}

// rkentry.EventLoggerEntry of prom entry
func WithEventLoggerEntryProm(eventLoggerEntry *rkentry.EventLoggerEntry) PromEntryOption {
	return func(entry *PromEntry) {
		entry.EventLoggerEntry = eventLoggerEntry
	}
}

// PushGateway of prom entry
func WithPusherProm(pusher *rkprom.PushGatewayPusher) PromEntryOption {
	return func(entry *PromEntry) {
		entry.Pusher = pusher
	}
}

// PushGateway of prom entry
func WithCertStoreProm(store *rkentry.CertStore) PromEntryOption {
	return func(entry *PromEntry) {
		entry.CertStore = store
	}
}

// Provide a new prometheus registry
func WithPromRegistryProm(registry *prometheus.Registry) PromEntryOption {
	return func(entry *PromEntry) {
		if registry != nil {
			entry.Registry = registry
		}
	}
}

// Create a prom entry with options and add prom entry to rk_ctx.GlobalAppCtx
func NewPromEntry(opts ...PromEntryOption) *PromEntry {
	entry := &PromEntry{
		Port:             defaultPort,
		Path:             defaultPath,
		EventLoggerEntry: rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
		ZapLoggerEntry:   rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
		entryName:        "gin-prom",
		entryType:        "gin-prom",
		Registerer:       prometheus.DefaultRegisterer,
		Gatherer:         prometheus.DefaultGatherer,
	}

	for i := range opts {
		opts[i](entry)
	}

	// Trim space by default
	entry.Path = strings.TrimSpace(entry.Path)

	if len(entry.Path) < 1 {
		// Invalid path, use default one
		entry.Path = defaultPath
	}

	if !strings.HasPrefix(entry.Path, "/") {
		entry.Path = "/" + entry.Path
	}

	if entry.ZapLoggerEntry == nil {
		entry.ZapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
	}

	if entry.EventLoggerEntry == nil {
		entry.EventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
	}

	if entry.Registry != nil {
		entry.Registerer = entry.Registry
		entry.Gatherer = entry.Registry
	}

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

// Start prometheus client
func (entry *PromEntry) Bootstrap(context.Context) {
	// start pusher
	if entry.Pusher != nil {
		entry.Pusher.Start()
	}
}

// Shutdown prometheus client
func (entry *PromEntry) Interrupt(context.Context) {
	if entry.Pusher != nil {
		entry.Pusher.Stop()
	}
}

// Return name of prom entry
func (entry *PromEntry) GetName() string {
	return entry.entryName
}

// Return type of prom entry
func (entry *PromEntry) GetType() string {
	return entry.entryType
}

// Stringfy prom entry
func (entry *PromEntry) String() string {
	m := map[string]interface{}{
		"entry_name": entry.entryName,
		"entry_type": entry.entryType,
		"prom_path":  entry.Path,
		"prom_port":  entry.Port,
	}

	if entry.Pusher != nil {
		m["pusher_remote_addr"] = entry.Pusher.RemoteAddress
		m["pusher_interval_ms"] = entry.Pusher.IntervalMS
		m["pusher_job_name"] = entry.Pusher.JobName
	}

	bytes, _ := json.Marshal(m)

	return string(bytes)
}

// Register collectors in default registry
func (entry *PromEntry) RegisterCollectors(collectors ...prometheus.Collector) error {
	var err error
	for i := range collectors {
		if innerErr := entry.Registerer.Register(collectors[i]); innerErr != nil {
			err = innerErr
		}
	}

	return err
}
