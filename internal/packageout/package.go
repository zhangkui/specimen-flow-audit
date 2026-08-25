package packageout

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/alert"
	"github.com/zhangkui/specimen-flow-audit/internal/report"
)

type Manifest struct {
	Station   string    `json:"station"`
	CreatedAt time.Time `json:"created_at"`
	Reports   int       `json:"reports"`
	Alerts    int       `json:"alerts"`
	Files     []string  `json:"files"`
}

func Build(station string, reports []report.Quality, alerts []alert.Alert) ([]byte, Manifest, error) {
	buffer := bytes.NewBuffer(nil)
	writer := zip.NewWriter(buffer)
	manifest := Manifest{Station: station, CreatedAt: time.Now().UTC(), Reports: len(reports), Alerts: len(alerts), Files: []string{"reports.json", "alerts.json", "manifest.json"}}
	for _, entry := range []struct {
		Name  string
		Value any
	}{{"reports.json", reports}, {"alerts.json", alerts}, {"manifest.json", manifest}} {
		file, err := writer.Create(entry.Name)
		if err != nil {
			return nil, Manifest{}, err
		}
		raw, err := json.MarshalIndent(entry.Value, "", "  ")
		if err != nil {
			return nil, Manifest{}, err
		}
		if _, err := file.Write(raw); err != nil {
			return nil, Manifest{}, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, Manifest{}, err
	}
	return buffer.Bytes(), manifest, nil
}
func Write(writer io.Writer, content []byte) error {
	count, err := writer.Write(content)
	if err != nil {
		return err
	}
	if count != len(content) {
		return fmt.Errorf("short package write: %d of %d", count, len(content))
	}
	return nil
}
