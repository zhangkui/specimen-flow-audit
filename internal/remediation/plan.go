package remediation

import (
	"sort"

	"github.com/zhangkui/specimen-flow-audit/internal/health"
	"github.com/zhangkui/specimen-flow-audit/internal/spectrum"
	"github.com/zhangkui/specimen-flow-audit/internal/trend"
)

type Priority string

const (
	Urgent Priority = "urgent"
	High   Priority = "high"
	Normal Priority = "normal"
)

type Action struct {
	Code     string   `json:"code"`
	Priority Priority `json:"priority"`
	Summary  string   `json:"summary"`
	Evidence string   `json:"evidence"`
}

type Input struct {
	Health      health.Snapshot
	Trend       trend.Snapshot
	Spectrum    spectrum.Report
	HasSpectrum bool
}

type Plan struct {
	Station string       `json:"station"`
	State   health.State `json:"state"`
	Actions []Action     `json:"actions"`
}

type rule func(Input) *Action

func Build(input Input) Plan {
	plan := Plan{Station: input.Health.Station, State: input.Health.State}
	seen := map[string]struct{}{}
	for _, evaluate := range rules() {
		action := evaluate(input)
		if action == nil {
			continue
		}
		if _, exists := seen[action.Code]; exists {
			continue
		}
		seen[action.Code] = struct{}{}
		plan.Actions = append(plan.Actions, *action)
	}
	if len(plan.Actions) == 0 {
		plan.Actions = []Action{{Code: "continue-observation", Priority: Normal, Summary: "保持常规监测", Evidence: "当前质量、趋势和频谱均未触发处置阈值"}}
	}
	sort.Slice(plan.Actions, func(i, j int) bool {
		left, right := priorityRank(plan.Actions[i].Priority), priorityRank(plan.Actions[j].Priority)
		if left == right {
			return plan.Actions[i].Code < plan.Actions[j].Code
		}
		return left < right
	})
	return plan
}

func rules() []rule {
	return []rule{
		missingTraceRule,
		criticalStateRule,
		availabilityRule,
		completenessRule,
		trendRule,
		narrowBandRule,
		broadBandRule,
	}
}

func missingTraceRule(input Input) *Action {
	if input.Health.State != health.Unknown {
		return nil
	}
	return &Action{Code: "verify-telemetry", Priority: Urgent, Summary: "核查台站遥测链路", Evidence: "评估窗口内没有已完成波形"}
}

func criticalStateRule(input Input) *Action {
	if input.Health.State != health.Critical {
		return nil
	}
	return &Action{Code: "dispatch-on-call", Priority: Urgent, Summary: "通知值班人员处置严重台站状态", Evidence: "台站健康状态为 critical"}
}

func availabilityRule(input Input) *Action {
	if !input.Health.HasLatest {
		return nil
	}
	availability := input.Health.Availability.Availability
	if availability >= 0.95 {
		return nil
	}
	priority := Normal
	if availability < 0.80 {
		priority = High
	}
	return &Action{Code: "inspect-uplink", Priority: priority, Summary: "检查数据链路、缓存队列和传输丢包", Evidence: "有效采集率低于 95%"}
}

func completenessRule(input Input) *Action {
	if !input.Health.HasLatest || input.Health.Latest.Completeness >= 0.95 {
		return nil
	}
	priority := Normal
	if input.Health.Latest.Completeness < 0.80 {
		priority = High
	}
	return &Action{Code: "backfill-waveform", Priority: priority, Summary: "从上游数据源回补缺失波形", Evidence: "最新波形完整率低于 95%"}
}

func trendRule(input Input) *Action {
	if input.Trend.State != trend.Degrading {
		return nil
	}
	return &Action{Code: "schedule-calibration", Priority: High, Summary: "安排传感器校准和现场环境检查", Evidence: "长期质量趋势为 degrading"}
}

func narrowBandRule(input Input) *Action {
	if !input.HasSpectrum || input.Spectrum.SpectralEntropy >= 0.35 || input.Spectrum.TotalPower == 0 {
		return nil
	}
	return &Action{Code: "inspect-narrowband-noise", Priority: Normal, Summary: "排查工频、机械共振或窄带电磁干扰", Evidence: "谱熵偏低，能量集中在少数频率"}
}

func broadBandRule(input Input) *Action {
	if !input.HasSpectrum || input.Spectrum.SpectralEntropy <= 0.85 || input.Health.State == health.Healthy {
		return nil
	}
	return &Action{Code: "inspect-broadband-noise", Priority: Normal, Summary: "排查宽带环境噪声与传感器安装状态", Evidence: "谱熵偏高且台站未处于健康状态"}
}

func priorityRank(priority Priority) int {
	switch priority {
	case Urgent:
		return 0
	case High:
		return 1
	default:
		return 2
	}
}
