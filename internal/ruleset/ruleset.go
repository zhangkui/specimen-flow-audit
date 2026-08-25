package ruleset

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/qc"
)

var (
	ErrInvalidVersion  = errors.New("invalid ruleset version")
	ErrNoActiveVersion = errors.New("no active ruleset version")
)

type Version struct {
	ID          string    `json:"id"`
	Station     string    `json:"station"`
	EffectiveAt time.Time `json:"effective_at"`
	Policy      qc.Policy `json:"policy"`
	Author      string    `json:"author"`
	Note        string    `json:"note"`
}

func (v Version) Validate() error {
	if v.ID == "" || v.Station == "" || v.EffectiveAt.IsZero() || v.Author == "" || v.Policy.MaxRMS <= 0 || v.Policy.MaxGapFraction < 0 || v.Policy.MaxGapFraction > 1 || v.Policy.MaxAbsolute <= 0 {
		return ErrInvalidVersion
	}
	return nil
}

type Registry struct {
	mu       sync.RWMutex
	versions map[string][]Version
}

func New() *Registry { return &Registry{versions: make(map[string][]Version)} }
func (r *Registry) Publish(version Version) error {
	if err := version.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	list := append(r.versions[version.Station], version)
	sort.Slice(list, func(i, j int) bool { return list[i].EffectiveAt.Before(list[j].EffectiveAt) })
	r.versions[version.Station] = list
	return nil
}
func (r *Registry) Active(station string, at time.Time) (Version, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := r.versions[station]
	for index := len(list) - 1; index >= 0; index-- {
		if !list[index].EffectiveAt.After(at) {
			return list[index], nil
		}
	}
	return Version{}, ErrNoActiveVersion
}
func (r *Registry) History(station string) []Version {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]Version(nil), r.versions[station]...)
}
