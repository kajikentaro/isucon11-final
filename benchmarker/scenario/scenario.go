package scenario

import (
	"context"
	"math/rand"
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

const (
	InitialStudentsCount = 50
	RegisterCourseLimit  = 20
	SearchCourseLimit    = 5
	InitialCourseCount   = 20
	CourseProcessLimit   = 5
)

type Scenario struct {
	BaseURL *url.URL
	UseTLS  bool
	NoLoad  bool

	sPubSub             *pubsub.PubSub
	cPubSub             *pubsub.PubSub
	courses             []*model.Course
	faculties           []*model.Faculty
	studentPool         *userPool
	activeStudent       []*model.Student
	activeStudentCount  int // FIXME Debug
	finishedCourseCount int // FIXME Debug
	language            string

	mu sync.RWMutex
}

func NewScenario() (*Scenario, error) {
	studentsData, err := generate.LoadStudentsData()
	if err != nil {
		return nil, err
	}
	facultiesData, err := generate.LoadFacultiesData()
	if err != nil {
		return nil, err
	}
	faculties := make([]*model.Faculty, len(facultiesData))
	for i, f := range facultiesData {
		faculties[i] = model.NewFaculty(f)
	}

	return &Scenario{
		sPubSub:       pubsub.NewPubSub(),
		cPubSub:       pubsub.NewPubSub(),
		courses:       []*model.Course{}, // 全コース
		faculties:     faculties,
		studentPool:   NewUserPool(studentsData),
		activeStudent: make([]*model.Student, 0, InitialStudentsCount),
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

func (s *Scenario) AddActiveStudent(student *model.Student) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.activeStudent = append(s.activeStudent, student)
}
func (s *Scenario) ActiveStudentCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.activeStudentCount
}

func (s *Scenario) CourseCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.courses)
}
func (s *Scenario) FinishedCourseCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.finishedCourseCount
}

func (s *Scenario) AddCourse(course *model.Course) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.courses = append(s.courses, course)
}

func (s *Scenario) GetRandomFaculty() *model.Faculty {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.faculties[rand.Intn(len(s.faculties))]
}

func (s *Scenario) SetFacultiesURL(url *url.URL) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, v := range s.faculties {
		v.Agent.BaseURL = url
	}
}
