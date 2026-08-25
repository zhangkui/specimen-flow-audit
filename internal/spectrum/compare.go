package spectrum

import (
	"math"
	"sort"
)

type Comparison struct {
	LeftStation                 string  `json:"left_station"`
	RightStation                string  `json:"right_station"`
	DominantFrequencyDifference float64 `json:"dominant_frequency_difference"`
	EntropyDifference           float64 `json:"entropy_difference"`
	BandShareDistance           float64 `json:"band_share_distance"`
	Similarity                  float64 `json:"similarity"`
}

func Compare(left, right Report) Comparison {
	comparison := Comparison{
		LeftStation:                 left.Station,
		RightStation:                right.Station,
		DominantFrequencyDifference: math.Abs(left.DominantFrequency - right.DominantFrequency),
		EntropyDifference:           math.Abs(left.SpectralEntropy - right.SpectralEntropy),
	}
	comparison.BandShareDistance = bandShareDistance(left.Bands, right.Bands)
	resolution := math.Max(left.FrequencyResolution, right.FrequencyResolution)
	if resolution <= 0 {
		resolution = 1
	}
	normalizedFrequency := comparison.DominantFrequencyDifference / resolution
	comparison.Similarity = 1 / (1 + normalizedFrequency + comparison.EntropyDifference + comparison.BandShareDistance)
	return comparison
}

func bandShareDistance(left, right []BandPower) float64 {
	shares := make(map[string][2]float64, len(left)+len(right))
	for _, band := range left {
		pair := shares[band.Name]
		pair[0] = band.Share
		shares[band.Name] = pair
	}
	for _, band := range right {
		pair := shares[band.Name]
		pair[1] = band.Share
		shares[band.Name] = pair
	}
	if len(shares) == 0 {
		return 0
	}
	names := make([]string, 0, len(shares))
	for name := range shares {
		names = append(names, name)
	}
	sort.Strings(names)
	sumSquares := 0.0
	for _, name := range names {
		pair := shares[name]
		delta := pair[0] - pair[1]
		sumSquares += delta * delta
	}
	return math.Sqrt(sumSquares / float64(len(names)))
}
