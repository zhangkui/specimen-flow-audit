package health

import (
	"github.com/zhangkui/specimen-flow-audit/internal/availability"
	"github.com/zhangkui/specimen-flow-audit/internal/baseline"
	"github.com/zhangkui/specimen-flow-audit/internal/report"
)

type State string

const (
	Unknown  State = "unknown"
	Healthy  State = "healthy"
	Degraded State = "degraded"
	Critical State = "critical"
)

type Input struct {
	Station           string
	Availability      availability.Summary
	Latest            report.Quality
	HasLatest         bool
	Baseline          baseline.Model
	HasBaseline       bool
	AnomalyScore      float64
	OpenIncidents     int
	CriticalIncidents int
}

type Snapshot struct {
	Station           string               `json:"station"`
	State             State                `json:"state"`
	Availability      availability.Summary `json:"availability"`
	Latest            report.Quality       `json:"latest"`
	HasLatest         bool                 `json:"has_latest"`
	Baseline          baseline.Model       `json:"baseline"`
	HasBaseline       bool                 `json:"has_baseline"`
	AnomalyScore      float64              `json:"anomaly_score"`
	OpenIncidents     int                  `json:"open_incidents"`
	CriticalIncidents int                  `json:"critical_incidents"`
	Reasons           []string             `json:"reasons"`
}

func Assess(input Input) Snapshot {
	snapshot := Snapshot{
		Station:           input.Station,
		State:             Healthy,
		Availability:      input.Availability,
		Latest:            input.Latest,
		HasLatest:         input.HasLatest,
		Baseline:          input.Baseline,
		HasBaseline:       input.HasBaseline,
		AnomalyScore:      input.AnomalyScore,
		OpenIncidents:     input.OpenIncidents,
		CriticalIncidents: input.CriticalIncidents,
	}
	if !input.HasLatest {
		snapshot.State = Unknown
		snapshot.Reasons = []string{"没有可用于评估的已完成波形"}
		return snapshot
	}

	if input.CriticalIncidents > 0 {
		return critical(snapshot, "存在未关闭的严重事件")
	}
	if input.Availability.Availability < 0.80 {
		return critical(snapshot, "有效采集率低于 80%")
	}
	if input.Latest.Completeness < 0.80 {
		return critical(snapshot, "最新波形完整率低于 80%")
	}
	if input.HasBaseline && input.Baseline.Points >= 2 && input.AnomalyScore >= 5 {
		return critical(snapshot, "噪声偏离历史基线超过 5 个标准差")
	}

	if input.OpenIncidents > 0 {
		snapshot.Reasons = append(snapshot.Reasons, "存在待处置事件")
	}
	if input.Availability.Availability < 0.95 {
		snapshot.Reasons = append(snapshot.Reasons, "有效采集率低于 95%")
	}
	if input.Latest.Completeness < 0.95 {
		snapshot.Reasons = append(snapshot.Reasons, "最新波形完整率低于 95%")
	}
	if input.HasBaseline && input.Baseline.Points >= 2 && input.AnomalyScore >= 3 {
		snapshot.Reasons = append(snapshot.Reasons, "噪声偏离历史基线超过 3 个标准差")
	}
	if len(snapshot.Reasons) > 0 {
		snapshot.State = Degraded
	}
	return snapshot
}

func critical(snapshot Snapshot, reason string) Snapshot {
	snapshot.State = Critical
	snapshot.Reasons = []string{reason}
	return snapshot
}
