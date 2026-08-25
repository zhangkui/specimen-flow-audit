package channel

import (
	"errors"
	"sort"
	"strings"
	"sync"
)

var (
	ErrInvalid = errors.New("invalid channel profile")
	ErrUnknown = errors.New("unknown channel")
)

type Profile struct {
	Station    string  `json:"station"`
	Code       string  `json:"code"`
	SampleRate float64 `json:"sample_rate"`
	Unit       string  `json:"unit"`
	Required   bool    `json:"required"`
	Enabled    bool    `json:"enabled"`
}

func (p Profile) Validate() error {
	if strings.TrimSpace(p.Station) == "" || strings.TrimSpace(p.Code) == "" || p.SampleRate <= 0 || strings.TrimSpace(p.Unit) == "" {
		return ErrInvalid
	}
	return nil
}

type Catalog struct {
	mu       sync.RWMutex
	profiles map[string]Profile
}

func New() *Catalog                   { return &Catalog{profiles: make(map[string]Profile)} }
func key(station, code string) string { return station + "/" + code }
func (c *Catalog) Upsert(profile Profile) error {
	if err := profile.Validate(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.profiles[key(profile.Station, profile.Code)] = profile
	return nil
}
func (c *Catalog) Get(station, code string) (Profile, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	profile, ok := c.profiles[key(station, code)]
	if !ok {
		return Profile{}, ErrUnknown
	}
	return profile, nil
}
func (c *Catalog) ForStation(station string) []Profile {
	c.mu.RLock()
	defer c.mu.RUnlock()
	items := []Profile{}
	for _, profile := range c.profiles {
		if profile.Station == station {
			items = append(items, profile)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Code < items[j].Code })
	return items
}
