package workflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/alert"
	"github.com/zhangkui/specimen-flow-audit/internal/audit"
	"github.com/zhangkui/specimen-flow-audit/internal/availability"
	"github.com/zhangkui/specimen-flow-audit/internal/baseline"
	"github.com/zhangkui/specimen-flow-audit/internal/compare"
	"github.com/zhangkui/specimen-flow-audit/internal/daily"
	"github.com/zhangkui/specimen-flow-audit/internal/dashboard"
	"github.com/zhangkui/specimen-flow-audit/internal/health"
	"github.com/zhangkui/specimen-flow-audit/internal/importer"
	"github.com/zhangkui/specimen-flow-audit/internal/incident"
	"github.com/zhangkui/specimen-flow-audit/internal/ledger"
	"github.com/zhangkui/specimen-flow-audit/internal/maintenance"
	"github.com/zhangkui/specimen-flow-audit/internal/notify"
	"github.com/zhangkui/specimen-flow-audit/internal/packageout"
	"github.com/zhangkui/specimen-flow-audit/internal/pipeline"
	"github.com/zhangkui/specimen-flow-audit/internal/remediation"
	"github.com/zhangkui/specimen-flow-audit/internal/report"
	"github.com/zhangkui/specimen-flow-audit/internal/review"
	"github.com/zhangkui/specimen-flow-audit/internal/ruleset"
	"github.com/zhangkui/specimen-flow-audit/internal/sla"
	"github.com/zhangkui/specimen-flow-audit/internal/spectrum"
	"github.com/zhangkui/specimen-flow-audit/internal/station"
	"github.com/zhangkui/specimen-flow-audit/internal/trend"
)

var ErrDuplicateJob = errors.New("analysis job already exists")

type SubmitRequest struct {
	JobID       string
	Format      importer.Format
	Station     string
	Start       time.Time
	SubmittedAt time.Time
	Body        io.Reader
}

type Completed struct {
	Job    ledger.Job
	Result pipeline.Result
	Alerts []alert.Alert
}

type Coordinator struct {
	stations  *station.Catalog
	rules     *ruleset.Registry
	ledger    *ledger.Ledger
	calendar  *maintenance.Calendar
	incidents *incident.Register
	reviews   *review.Queue
	notices   *notify.Outbox
	timeline  *audit.Timeline
	baseline  *baseline.Registry

	mu      sync.RWMutex
	results map[string]Completed
	alerts  []alert.Alert
}

func New(stations *station.Catalog, rules *ruleset.Registry, jobs *ledger.Ledger) *Coordinator {
	return &Coordinator{stations: stations, rules: rules, ledger: jobs, calendar: maintenance.New(), incidents: incident.New(), reviews: review.New(), notices: notify.New(), timeline: audit.New(), baseline: baseline.New(), results: make(map[string]Completed)}
}

func (c *Coordinator) Submit(ctx context.Context, request SubmitRequest) (Completed, error) {
	if request.JobID == "" || request.SubmittedAt.IsZero() {
		return Completed{}, errors.New("job id and submit time are required")
	}
	// Atomically reserve the job id in the ledger. This is the single
	// commit point: the first request to create the entry proceeds, and
	// any concurrent request for the same JobID observes an existing job
	// and returns a duplicate error without overwriting the winner's
	// station, rule version, or completion state.
	if !c.ledger.Submit(ledger.Job{ID: request.JobID, Station: request.Station, SubmittedAt: request.SubmittedAt}) {
		return Completed{}, ErrDuplicateJob
	}
	profile, err := c.stations.Get(request.Station)
	if err != nil {
		c.ledger.Finish(request.JobID, request.SubmittedAt, err)
		return Completed{}, fmt.Errorf("station profile: %w", err)
	}
	c.timeline.Add(audit.Entry{ID: request.JobID + ":submitted", Subject: request.JobID, Action: "submitted", Actor: "ingest", At: request.SubmittedAt, Detail: request.Station})
	version, err := c.rules.Active(request.Station, request.Start)
	if err != nil {
		c.ledger.Finish(request.JobID, request.SubmittedAt, err)
		return Completed{}, fmt.Errorf("active rule: %w", err)
	}
	if !c.ledger.Start(request.JobID, version.ID, request.SubmittedAt) {
		return Completed{}, ErrDuplicateJob
	}
	imported, err := importer.Read(ctx, importer.Request{ID: request.JobID, Format: request.Format, Station: request.Station, Start: request.Start, Body: request.Body, SubmittedAt: request.SubmittedAt})
	if err != nil {
		c.ledger.Finish(request.JobID, request.SubmittedAt, err)
		return Completed{}, err
	}
	curve := profile.Calibration
	curve.Station = request.Station
	result, err := pipeline.Run(pipeline.Request{Trace: imported.Trace, Calibration: curve, Policy: version.Policy, EventThreshold: profile.Trigger})
	if err != nil {
		c.ledger.Finish(request.JobID, request.SubmittedAt, err)
		return Completed{}, err
	}
	alerts := alert.Build(result.Trace, result.Findings, result.Events)
	if _, covered := c.calendar.At(request.Station, request.Start); covered {
		alerts = suppressQualityAlerts(alerts)
	}
	c.ledger.Finish(request.JobID, request.SubmittedAt, nil)
	job, _ := c.ledger.Get(request.JobID)
	completed := Completed{Job: job, Result: result, Alerts: alerts}
	c.mu.Lock()
	c.results[request.JobID] = completed
	c.alerts = append(c.alerts, alerts...)
	c.mu.Unlock()
	c.baseline.Add(result.Quality)
	for _, item := range alerts {
		if item.Level != alert.Critical {
			continue
		}
		id := fmt.Sprintf("%s:%s", item.Station, item.Code)
		_ = c.incidents.Open(incident.Incident{ID: id, Station: item.Station, Severity: item.Level, OpenedAt: item.At, Notes: []string{item.Code + ": " + item.Detail}})
		_ = c.reviews.Create(review.Task{ID: id, JobID: request.JobID, Station: item.Station, Alert: item, CreatedAt: request.SubmittedAt})
		c.notices.Queue(id, "seismic-ops@"+item.Station, item, request.SubmittedAt)
		c.timeline.Add(audit.Entry{ID: id + ":opened", Subject: request.JobID, Action: "critical-alert-opened", Actor: "quality", At: item.At, Detail: item.Code})
	}
	c.timeline.Add(audit.Entry{ID: request.JobID + ":completed", Subject: request.JobID, Action: "completed", Actor: "quality", At: request.SubmittedAt, Detail: version.ID})
	return completed, nil
}

func suppressQualityAlerts(alerts []alert.Alert) []alert.Alert {
	result := make([]alert.Alert, 0, len(alerts))
	for _, item := range alerts {
		if item.Code == "noise-rms" || item.Code == "missing-data" || item.Code == "amplitude" {
			continue
		}
		result = append(result, item)
	}
	return result
}

func (c *Coordinator) ScheduleMaintenance(window maintenance.Window) error {
	return c.calendar.Schedule(window)
}
func (c *Coordinator) OpenIncidents(station string) []incident.Incident {
	return c.incidents.OpenForStation(station)
}
func (c *Coordinator) OpenReviews(station string) []review.Task { return c.reviews.Open(station) }
func (c *Coordinator) PendingNotifications() []notify.Message   { return c.notices.Pending() }
func (c *Coordinator) AuditTrail(jobID string) []audit.Entry    { return c.timeline.ForSubject(jobID) }
func (c *Coordinator) AcknowledgeIncident(id, assignee, note string, at time.Time) (incident.Incident, error) {
	return c.incidents.Acknowledge(id, assignee, note, at)
}
func (c *Coordinator) ResolveIncident(id, note string, at time.Time) (incident.Incident, error) {
	return c.incidents.Resolve(id, note, at)
}
func (c *Coordinator) IncidentSLA(id string, target sla.Target, now time.Time) (sla.Result, error) {
	item, err := c.incidents.Get(id)
	if err != nil {
		return sla.Result{}, err
	}
	return sla.Evaluate(item, target, now), nil
}

func (c *Coordinator) Result(jobID string) (Completed, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result, ok := c.results[jobID]
	return result, ok
}

func (c *Coordinator) Dashboard() []dashboard.StationCard {
	c.mu.RLock()
	alerts := append([]alert.Alert(nil), c.alerts...)
	c.mu.RUnlock()
	jobs := []ledger.Job{}
	for _, card := range c.stations.List() {
		jobs = append(jobs, c.ledger.ByStation(card.Code)...)
	}
	return dashboard.Build(jobs, alerts)
}

func (c *Coordinator) Daily(location *time.Location) []daily.Summary {
	c.mu.RLock()
	reports := make([]pipeline.Result, 0, len(c.results))
	for _, result := range c.results {
		reports = append(reports, result.Result)
	}
	c.mu.RUnlock()
	qualities := make([]report.Quality, 0, len(reports))
	for _, result := range reports {
		qualities = append(qualities, result.Quality)
	}
	return daily.Summarize(qualities, location)
}

// Baseline builds a station quality baseline from its most recent completed analyses.
func (c *Coordinator) Baseline(stationCode string, window int) (baseline.Model, error) {
	return c.baseline.Build(stationCode, window)
}

// AnomalyScore compares a completed job's noise level against its station baseline.
func (c *Coordinator) AnomalyScore(jobID string, window int) (float64, error) {
	completed, ok := c.Result(jobID)
	if !ok {
		return 0, fmt.Errorf("analysis result %q not found", jobID)
	}
	model, err := c.Baseline(completed.Result.Quality.Station, window)
	if err != nil {
		return 0, err
	}
	return baseline.Score(model, completed.Result.Quality), nil
}

// Compare calculates the best bounded-lag correlation between two completed traces.
func (c *Coordinator) Compare(jobLeft, jobRight string, maxLag int) (compare.Result, error) {
	left, ok := c.Result(jobLeft)
	if !ok {
		return compare.Result{}, fmt.Errorf("analysis result %q not found", jobLeft)
	}
	right, ok := c.Result(jobRight)
	if !ok {
		return compare.Result{}, fmt.Errorf("analysis result %q not found", jobRight)
	}
	return compare.Correlate(left.Result.Trace, right.Result.Trace, maxLag)
}

// Package creates a portable archive containing all quality reports and alerts for a station.
func (c *Coordinator) Package(stationCode string) ([]byte, packageout.Manifest, error) {
	c.mu.RLock()
	reports := make([]report.Quality, 0)
	alerts := make([]alert.Alert, 0)
	for _, completed := range c.results {
		if completed.Result.Quality.Station == stationCode {
			reports = append(reports, completed.Result.Quality)
		}
	}
	for _, item := range c.alerts {
		if item.Station == stationCode {
			alerts = append(alerts, item)
		}
	}
	c.mu.RUnlock()
	sort.Slice(reports, func(i, j int) bool { return reports[i].ObservedAt.Before(reports[j].ObservedAt) })
	sort.Slice(alerts, func(i, j int) bool { return alerts[i].At.Before(alerts[j].At) })
	return packageout.Build(stationCode, reports, alerts)
}

// Health returns a station-level operational assessment for a requested observation window.
func (c *Coordinator) Health(stationCode string, start, end time.Time, baselineWindow int) (health.Snapshot, error) {
	if _, err := c.stations.Get(stationCode); err != nil {
		return health.Snapshot{}, err
	}
	c.mu.RLock()
	var latest pipeline.Result
	hasLatest := false
	received := 0
	for _, completed := range c.results {
		trace := completed.Result.Trace
		if trace.Station != stationCode {
			continue
		}
		if !hasLatest || trace.Start.After(latest.Trace.Start) {
			latest = completed.Result
			hasLatest = true
		}
		if !trace.Start.Before(start) && trace.Start.Before(end) {
			received += len(trace.Samples)
		}
	}
	c.mu.RUnlock()
	if !hasLatest {
		return health.Assess(health.Input{Station: stationCode}), nil
	}
	availabilitySummary := availability.Measure(stationCode, start, end, latest.Trace.SampleRate, received, c.calendar)
	model, baselineErr := c.Baseline(stationCode, baselineWindow)
	input := health.Input{Station: stationCode, Availability: availabilitySummary, Latest: latest.Quality, HasLatest: true}
	if baselineErr == nil {
		input.Baseline = model
		input.HasBaseline = true
		input.AnomalyScore = baseline.Score(model, latest.Quality)
	}
	open := c.incidents.OpenForStation(stationCode)
	input.OpenIncidents = len(open)
	for _, item := range open {
		if item.Severity == alert.Critical {
			input.CriticalIncidents++
		}
	}
	return health.Assess(input), nil
}

func (c *Coordinator) Spectrum(jobID string, config spectrum.Config) (spectrum.Report, error) {
	completed, ok := c.Result(jobID)
	if !ok {
		return spectrum.Report{}, fmt.Errorf("analysis result %q not found", jobID)
	}
	return spectrum.Analyze(completed.Result.Trace, config)
}

func (c *Coordinator) SpectralFindings(jobID string, policy spectrum.Policy) ([]spectrum.Finding, error) {
	report, err := c.Spectrum(jobID, spectrum.Config{})
	if err != nil {
		return nil, err
	}
	return spectrum.Evaluate(report, policy)
}

func (c *Coordinator) CompareSpectrum(jobLeft, jobRight string, config spectrum.Config) (spectrum.Comparison, error) {
	left, err := c.Spectrum(jobLeft, config)
	if err != nil {
		return spectrum.Comparison{}, err
	}
	right, err := c.Spectrum(jobRight, config)
	if err != nil {
		return spectrum.Comparison{}, err
	}
	return spectrum.Compare(left, right), nil
}

func (c *Coordinator) Trend(stationCode string, since time.Time, config trend.Config) (trend.Snapshot, error) {
	if _, err := c.stations.Get(stationCode); err != nil {
		return trend.Snapshot{}, err
	}
	c.mu.RLock()
	qualities := make([]report.Quality, 0)
	for _, completed := range c.results {
		quality := completed.Result.Quality
		if quality.Station == stationCode && (since.IsZero() || !quality.ObservedAt.Before(since)) {
			qualities = append(qualities, quality)
		}
	}
	c.mu.RUnlock()
	return trend.Analyze(stationCode, qualities, config)
}

func (c *Coordinator) Remediation(jobID string, start, end time.Time, baselineWindow int) (remediation.Plan, error) {
	completed, ok := c.Result(jobID)
	if !ok {
		return remediation.Plan{}, fmt.Errorf("analysis result %q not found", jobID)
	}
	stationCode := completed.Result.Quality.Station
	healthSnapshot, err := c.Health(stationCode, start, end, baselineWindow)
	if err != nil {
		return remediation.Plan{}, err
	}
	trendSnapshot, err := c.Trend(stationCode, time.Time{}, trend.Config{})
	if err != nil {
		return remediation.Plan{}, err
	}
	spectral, err := c.Spectrum(jobID, spectrum.Config{})
	if err != nil {
		return remediation.Plan{}, err
	}
	return remediation.Build(remediation.Input{Health: healthSnapshot, Trend: trendSnapshot, Spectrum: spectral, HasSpectrum: true}), nil
}
