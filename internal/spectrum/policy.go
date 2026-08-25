package spectrum

import (
	"errors"
	"sort"
)

var ErrInvalidPolicy = errors.New("invalid spectral quality policy")

type Severity string

const (
	Advisory Severity = "advisory"
	Warning  Severity = "warning"
	Failure  Severity = "failure"
)

type Policy struct {
	MinEntropy       float64            `json:"min_entropy"`
	MaxDominantShare float64            `json:"max_dominant_share"`
	MaxBandShare     map[string]float64 `json:"max_band_share"`
}

type Finding struct {
	Code     string   `json:"code"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Value    float64  `json:"value"`
	Limit    float64  `json:"limit"`
}

func DefaultPolicy() Policy {
	return Policy{
		MinEntropy:       0.30,
		MaxDominantShare: 0.70,
		MaxBandShare: map[string]float64{
			"high-frequency": 0.55,
		},
	}
}

func Evaluate(report Report, policy Policy) ([]Finding, error) {
	policy = normalizePolicy(policy)
	if err := validatePolicy(policy); err != nil {
		return nil, err
	}
	findings := make([]Finding, 0)
	if report.TotalPower == 0 {
		return []Finding{{Code: "silent-trace", Severity: Advisory, Message: "波形没有可分析的交流能量", Value: 0}}, nil
	}
	dominantShare := report.DominantPower / report.TotalPower
	if report.SpectralEntropy < policy.MinEntropy {
		findings = append(findings, Finding{Code: "low-spectral-entropy", Severity: Warning, Message: "频谱能量过度集中，可能存在窄带干扰", Value: report.SpectralEntropy, Limit: policy.MinEntropy})
	}
	if dominantShare > policy.MaxDominantShare {
		findings = append(findings, Finding{Code: "dominant-tone", Severity: Failure, Message: "单一频率能量占比超过允许值", Value: dominantShare, Limit: policy.MaxDominantShare})
	}
	for _, band := range report.Bands {
		limit, monitored := policy.MaxBandShare[band.Name]
		if monitored && band.Share > limit {
			findings = append(findings, Finding{Code: "band-power-" + band.Name, Severity: Warning, Message: "监测频段能量占比超过允许值", Value: band.Share, Limit: limit})
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Severity == findings[j].Severity {
			return findings[i].Code < findings[j].Code
		}
		return severityRank(findings[i].Severity) < severityRank(findings[j].Severity)
	})
	return findings, nil
}

func normalizePolicy(policy Policy) Policy {
	defaults := DefaultPolicy()
	if policy.MinEntropy <= 0 {
		policy.MinEntropy = defaults.MinEntropy
	}
	if policy.MaxDominantShare <= 0 {
		policy.MaxDominantShare = defaults.MaxDominantShare
	}
	if policy.MaxBandShare == nil {
		policy.MaxBandShare = defaults.MaxBandShare
	}
	return policy
}

func validatePolicy(policy Policy) error {
	if policy.MinEntropy <= 0 || policy.MinEntropy > 1 || policy.MaxDominantShare <= 0 || policy.MaxDominantShare > 1 {
		return ErrInvalidPolicy
	}
	for _, limit := range policy.MaxBandShare {
		if limit <= 0 || limit > 1 {
			return ErrInvalidPolicy
		}
	}
	return nil
}

func severityRank(severity Severity) int {
	switch severity {
	case Failure:
		return 0
	case Warning:
		return 1
	default:
		return 2
	}
}
