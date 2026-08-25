package station

import (
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/zhangkui/specimen-flow-audit/internal/calibration"
	"github.com/zhangkui/specimen-flow-audit/internal/qc"
)

var (
	ErrInvalidProfile = errors.New("invalid station profile")
	ErrUnknownStation = errors.New("unknown station")
)

type Profile struct {
	Code        string            `json:"code"`
	Network     string            `json:"network"`
	Channels    []string          `json:"channels"`
	Calibration calibration.Curve `json:"calibration"`
	Policy      qc.Policy         `json:"policy"`
	Trigger     float64           `json:"trigger"`
}

func (p Profile) Validate() error {
	if strings.TrimSpace(p.Code) == "" || len(p.Channels) == 0 || p.Trigger <= 0 || p.Policy.MaxRMS <= 0 || p.Policy.MaxGapFraction < 0 || p.Policy.MaxGapFraction > 1 || p.Policy.MaxAbsolute <= 0 {
		return ErrInvalidProfile
	}
	if p.Calibration.Station != "" && p.Calibration.Station != p.Code {
		return ErrInvalidProfile
	}
	return nil
}

type Catalog struct {
	mu       sync.RWMutex
	profiles map[string]Profile
}

func New() *Catalog { return &Catalog{profiles: make(map[string]Profile)} }
func (c *Catalog) Upsert(profile Profile) error {
	if err := profile.Validate(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	profile.Channels = append([]string(nil), profile.Channels...)
	c.profiles[profile.Code] = profile
	return nil
}
func (c *Catalog) Get(code string) (Profile, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	profile, ok := c.profiles[code]
	if !ok {
		return Profile{}, ErrUnknownStation
	}
	profile.Channels = append([]string(nil), profile.Channels...)
	return profile, nil
}
func (c *Catalog) List() []Profile {
	c.mu.RLock()
	defer c.mu.RUnlock()
	items := make([]Profile, 0, len(c.profiles))
	for _, profile := range c.profiles {
		profile.Channels = append([]string(nil), profile.Channels...)
		items = append(items, profile)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Code < items[j].Code })
	return items
}
