package retention

import (
	"sort"
	"time"
)

type Record struct {
	ID          string
	Station     string
	CollectedAt time.Time
	SizeBytes   int64
}
type Plan struct {
	Delete         []Record
	Keep           []Record
	ReclaimedBytes int64
}

func Build(records []Record, cutoff time.Time) Plan {
	plan := Plan{}
	for _, record := range records {
		if record.CollectedAt.Before(cutoff) {
			plan.Delete = append(plan.Delete, record)
			plan.ReclaimedBytes += record.SizeBytes
		} else {
			plan.Keep = append(plan.Keep, record)
		}
	}
	sort.Slice(plan.Delete, func(i, j int) bool { return plan.Delete[i].CollectedAt.Before(plan.Delete[j].CollectedAt) })
	return plan
}
