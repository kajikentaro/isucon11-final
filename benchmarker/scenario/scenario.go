package scenario

import (
	"context"
	"net/url"
	"sync"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/pubsub"
	"github.com/isucon/isucon11-final/benchmarker/generate"
	"github.com/isucon/isucon11-final/benchmarker/model"
)

var (
	// Prepare, Load, Validationが返すエラー
	// Benchmarkが中断されたかどうか確認用
	Cancel failure.StringCode = "scenario-cancel"
)

type Scenario struct {
	BaseURL *url.URL
	UseTLS  bool
	NoLoad  bool

	sPubSub  *pubsub.PubSub
	cPubSub  *pubsub.PubSub
	courses  []*model.Course
	student  []*model.Student
	language string

	mu sync.Mutex
}

func NewScenario() (*Scenario, error) {
	initialStudents := generate.InitialStudents()
	return &Scenario{
		sPubSub: pubsub.NewPubSub(),
		cPubSub: pubsub.NewPubSub(),
		courses: []*model.Course{},
		student: initialStudents,
	}, nil
}

func (s *Scenario) Validation(context.Context, *isucandar.BenchmarkStep) error {
	if s.NoLoad {
		return nil
	}
	ContestantLogger.Printf("===> VALIDATION")

	return nil
}

func (s *Scenario) Language() string {
	return s.language
}
