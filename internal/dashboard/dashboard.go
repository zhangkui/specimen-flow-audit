package dashboard

import (
	"github.com/zhangkui/specimen-flow-audit/internal/alert"
	"github.com/zhangkui/specimen-flow-audit/internal/ledger"
	"sort"
	"time"
)

type StationCard struct {
	Station     string    `json:"station"`
	Imported    int       `json:"imported"`
	Completed   int       `json:"completed"`
	Failed      int       `json:"failed"`
	OpenAlerts  int       `json:"open_alerts"`
	LastUpdated time.Time `json:"last_updated"`
}

func Build(jobs []ledger.Job, alerts []alert.Alert) []StationCard {
	cards := map[string]StationCard{}
	for _, job := range jobs {
		card := cards[job.Station]
		card.Station = job.Station
		card.Imported++
		if job.State == ledger.Completed {
			card.Completed++
		}
		if job.State == ledger.Failed {
			card.Failed++
		}
		if job.FinishedAt.After(card.LastUpdated) {
			card.LastUpdated = job.FinishedAt
		}
		cards[job.Station] = card
	}
	for _, item := range alerts {
		card := cards[item.Station]
		card.Station = item.Station
		card.OpenAlerts++
		if item.At.After(card.LastUpdated) {
			card.LastUpdated = item.At
		}
		cards[item.Station] = card
	}
	result := make([]StationCard, 0, len(cards))
	for _, card := range cards {
		result = append(result, card)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Station < result[j].Station })
	return result
}
