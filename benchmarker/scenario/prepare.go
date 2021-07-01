package scenario

import (
	"context"
	"fmt"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucon11-final/benchmarker/api"
	"github.com/isucon/isucon11-final/benchmarker/fails"
	"github.com/isucon/isucon11-final/benchmarker/model"
)

func (s *Scenario) Prepare(ctx context.Context, step *isucandar.BenchmarkStep) error {
	ContestantLogger.Printf("===> PREPARE")

	if err := s.prepareCheck(ctx, step); err != nil {
		return Cancel
	}

	step.Result().Score.Reset()

	if s.NoLoad {
		return nil
	}

	f := model.StaticFacultyData
	_, err := InitializeAction(ctx, f.Agent)
	for err != nil {
		step.AddError(err) // for InitializeAction
		return Cancel
	}

	return nil
}

// prepareFailCheck は全Phaseの簡易チェックを行う
func (s *Scenario) prepareCheck(ctx context.Context, step *isucandar.BenchmarkStep) error {
	sd := model.StaticStudentsData[0]
	// StaticStudentsDataのAgentを利用しないように新たにAgentを載せたStudentを作成する
	initializeStudent := model.NewStudent(sd.Name, sd.Number, sd.RawPassword)
	initializeStudent.Agent.BaseURL = s.BaseURL
	initializeStudent.Agent.Name = "student_user_agent"

	f := model.StaticFacultyData
	initializeFaculty := model.NewFaculty(f.Name, f.Number, f.RawPassword)
	initializeFaculty.Agent.BaseURL = s.BaseURL
	initializeFaculty.Agent.Name = "faculty_user_agent"

	errors := step.Result().Errors
	hasErrors := func() bool {
		errors.Wait()

		return len(errors.All()) > 0
	}

	lang, err := InitializeAction(ctx, initializeFaculty.Agent)
	for err != nil {
		step.AddError(err) // for InitializeAction
		return Cancel
	}
	s.language = lang

	errs := LoginAction(ctx, initializeFaculty.Agent, initializeFaculty.UserData)
	if len(errs) > 0 {
		for _, err := range errs {
			step.AddError(err) // for LoginAction
		}
		return Cancel
	}

	// 履修登録期間
	if err := api.ChangePhaseToRegister(ctx, initializeFaculty.Agent); err != nil {
		return Cancel
	}
	// AccessCheckは並列でリクエストを行うのでerrorはstep.Results().Errorsを確認
	s.prepareAccessCheckInRegister(ctx, initializeStudent, initializeFaculty, step)
	if hasErrors() {
		return Cancel
	}
	if err := s.prepareFastCheckInRegister(ctx, initializeStudent, step); err != nil {
		return Cancel
	}

	// 講義期間
	if err := api.ChangePhaseToClasses(ctx, initializeFaculty.Agent); err != nil {
		return Cancel
	}
	s.prepareAccessCheckInClass(ctx, initializeStudent, initializeFaculty, step)
	if hasErrors() {
		return Cancel
	}
	if err := s.prepareFastCheckInClass(ctx, initializeStudent, initializeFaculty, step); err != nil {
		return Cancel
	}

	// 成績開示期間
	if err := api.ChangePhaseToResult(ctx, initializeFaculty.Agent); err != nil {
		return Cancel
	}

	s.prepareAccessCheckInResult(ctx, initializeStudent, initializeFaculty, step)
	if hasErrors() {
		return Cancel
	}

	if err := s.prepareFastCheckInResult(ctx, initializeStudent, initializeFaculty, step); err != nil {
		return Cancel
	}

	return nil
}

// TODO: 以下のTODOをaction.goあたりにまとめる
func (s *Scenario) prepareAccessCheckInRegister(ctx context.Context, student *model.Student, faculty *model.Faculty, step *isucandar.BenchmarkStep) {
	// 履修登録期間でのアクセス制御チェック
	// TODO: goroutineで各エンドポイントへアクセス確認.
	return
}
func (s *Scenario) prepareAccessCheckInClass(ctx context.Context, student *model.Student, faculty *model.Faculty, step *isucandar.BenchmarkStep) {
	// 講義期間でのアクセス制御チェック
	// TODO: goroutineで各エンドポイントへアクセス確認.
	return
}
func (s *Scenario) prepareAccessCheckInResult(ctx context.Context, student *model.Student, faculty *model.Faculty, step *isucandar.BenchmarkStep) {
	// 履修登録期間でのアクセス制御チェック
	// TODO: goroutineで各エンドポイントへアクセス確認.
	return
}

func (s *Scenario) prepareFastCheckInRegister(ctx context.Context, student *model.Student, step *isucandar.BenchmarkStep) error {
	// 履修登録期間での動作確認
	student.Agent.ClearCookie()

	if errs := LoginAction(ctx, student.Agent, student.UserData); len(errs) > 0 {
		for _, err := range errs {
			step.AddError(err)
		}
		err := failure.NewError(fails.ErrCritical, fmt.Errorf("初期走行のログイン処理が失敗しました"))
		return err
	}

	if errs := AccessMyPageAction(ctx, student.Agent); len(errs) > 0 {
		err := failure.NewError(fails.ErrCritical, fmt.Errorf("初期走行でマイページの描画に失敗しました"))
		step.AddError(err)
		return err
	}

	wantRegCourses := []*model.Course{model.StaticCoursesData[0]}

	// 希望のコースを仮登録
	if errs := AccessRegPageAction(ctx, student.Agent); len(errs) > 0 {
		err := failure.NewError(fails.ErrCritical, fmt.Errorf("初期走行で履修登録ページの描画に失敗しました"))
		step.AddError(err)
		return err
	}
	var semiRegCourses []*model.Course
	for _, c := range wantRegCourses {
		errs := SearchCoursesAction(ctx, student.Agent, c)
		if len(errs) > 0 {
			for _, err := range errs {
				step.AddError(err)
			}
		} else {
			semiRegCourses = append(semiRegCourses, c)
		}
	}
	if len(semiRegCourses) == 0 {
		err := failure.NewError(fails.ErrCritical, fmt.Errorf("初期走行で講義検索が一度も成功しませんでした"))
		step.AddError(err)
		return err
	}

	// 仮登録した講義を登録
	if err := RegisterCoursesAction(ctx, student, semiRegCourses); err != nil {
		step.AddError(err)
		return err
	}
	registered, err := FetchRegisteredCoursesAction(ctx, student)
	if err != nil {
		step.AddError(err)
		return err
	}
	if len(registered) == 0 {
		err := failure.NewError(fails.ErrCritical, fmt.Errorf("初期走行で講義が１つも登録できていませんでした"))
		step.AddError(err)
		return err
	}

	expected := student.Courses()
	if !equalCourses(expected, registered) {
		err := failure.NewError(fails.ErrCritical, fmt.Errorf("登録成功した講義と登録されている講義が一致しません"))
		step.AddError(err)
		return err
	}

	return nil
}
func (s *Scenario) prepareFastCheckInClass(ctx context.Context, student *model.Student, faculty *model.Faculty, step *isucandar.BenchmarkStep) error {
	// 講義期間での動作確認
	// TODO: Facultyによる資料追加
	// TODO: Facultyによるお知らせ追加
	// TODO: Studentによるお知らせ確認
	// TODO: Studentによる資料DL
	// TODO: Facultyによる出席コード追加
	// TODO: Studentによる出席コード入力
	// TODO: Facultyによる課題追加
	// TODO: Studentによる課題提出
	return nil
}
func (s *Scenario) prepareFastCheckInResult(ctx context.Context, student *model.Student, faculty *model.Faculty, step *isucandar.BenchmarkStep) error {
	// 成績開示期間での動作確認
	// TODO: Facultyによる講義Aの課題確認

	// 教師による成績登録
	faculty.Agent.ClearCookie()

	if errs := LoginAction(ctx, faculty.Agent, faculty.UserData); len(errs) > 0 {
		for _, err := range errs {
			step.AddError(err)
		}
		err := failure.NewError(fails.ErrCritical, fmt.Errorf("初期走行のadminログイン処理が失敗しました"))
		return err
	}

	courses := student.Courses()
	hasErr := false
	for _, c := range courses {
		if err := RegisterGradeAction(ctx, faculty, student, c); err != nil {
			step.AddError(err)
			hasErr = true
		}
	}

	if hasErr {
		err := failure.NewError(fails.ErrCritical, fmt.Errorf("初期走行で成績登録が失敗しました"))
		step.AddError(err)
		return err
	}

	// 生徒による成績確認
	student.Agent.ClearCookie()
	if errs := LoginAction(ctx, student.Agent, student.UserData); len(errs) > 0 {
		for _, err := range errs {
			step.AddError(err)
		}
		err := failure.NewError(fails.ErrCritical, fmt.Errorf("初期走行のログイン処理が失敗しました"))
		return err
	}

	if errs := AccessMyPageAction(ctx, student.Agent); len(errs) > 0 {
		err := failure.NewError(fails.ErrCritical, fmt.Errorf("初期走行でマイページの描画に失敗しました"))
		step.AddError(err)
		return err
	}

	grades, err := FetchGradesAction(ctx, student)
	if err != nil {
		step.AddError(err)
		return err
	}

	expectedGrade := student.FirseSemesterGrade()
	if len(grades) != len(expectedGrade) {
		err = failure.NewError(fails.ErrCritical, fmt.Errorf("初期走行で取得できた成績の数が期待したものと違いました"))
		step.AddError(err)
		return err
	}

	for id, grade := range grades {
		v, ok := expectedGrade[id]
		if !ok || grade != v {
			err = failure.NewError(fails.ErrCritical, fmt.Errorf("初期走行で成績が一致しませんでした"))
			step.AddError(err)
			return err
		}
	}

	return nil
}