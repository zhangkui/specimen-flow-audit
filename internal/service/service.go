package service

import (
	"github.com/zhangkui/specimen-flow-audit/internal/pipeline"
	"github.com/zhangkui/specimen-flow-audit/internal/report"
	"github.com/zhangkui/specimen-flow-audit/internal/store"
	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
)

type Service struct{ store *store.Store }

func New() *Service { return &Service{store: store.New()} }
func (s *Service) Analyze(trace waveform.Trace) (report.Quality, error) {
	if err := waveform.Validate(trace); err != nil {
		return report.Quality{}, err
	}
	s.store.Save(trace)
	return report.Build(trace), nil
}
func (s *Service) Latest(station string) (report.Quality, error) {
	trace, err := s.store.Load(station)
	if err != nil {
		return report.Quality{}, err
	}
	return report.Build(trace), nil
}

func (s *Service) Run(request pipeline.Request) (pipeline.Result, error) {
	result, err := pipeline.Run(request)
	if err != nil {
		return pipeline.Result{}, err
	}
	s.store.Save(result.Trace)
	return result, nil
}
