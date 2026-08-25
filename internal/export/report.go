package export

import (
	"encoding/csv"
	"github.com/zhangkui/specimen-flow-audit/internal/report"
	"io"
	"strconv"
)

func WriteCSV(writer io.Writer, reports []report.Quality) error {
	output := csv.NewWriter(writer)
	if err := output.Write([]string{"station", "duration_seconds", "gap_count", "noise_rms", "completeness"}); err != nil {
		return err
	}
	for _, item := range reports {
		if err := output.Write([]string{item.Station, strconv.FormatFloat(item.DurationSeconds, 'f', 6, 64), strconv.Itoa(item.GapCount), strconv.FormatFloat(item.NoiseRMS, 'f', 6, 64), strconv.FormatFloat(item.Completeness, 'f', 6, 64)}); err != nil {
			return err
		}
	}
	output.Flush()
	return output.Error()
}
