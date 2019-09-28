package scenario

import (
	"context"
	"math/rand"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/xrandom"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"github.com/chibiegg/isucon9-final/bench/payment"
)

// NormalScenario は基本的な予約フローのシナリオです
func NormalScenario(ctx context.Context) error {
	client, err := isutrain.NewClient()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	paymentClient, err := payment.NewClient()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if config.Debug {
		client.ReplaceMockTransport()
	}

	user, err := xrandom.GetRandomUser()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.Signup(ctx, user.Email, user.Password, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.Login(ctx, user.Email, user.Password, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListStations(ctx, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	useAt := xrandom.GetRandomUseAt()
	departure, arrival := xrandom.GetRandomSection()
	trains, err := client.SearchTrains(ctx, useAt, departure, arrival, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if len(trains) == 0 {
		return nil
	}

	trainIdx := rand.Intn(len(trains))
	train := trains[trainIdx]
	var (
		carNumMin = 1
		carNumMax = 16
		carNum    = rand.Intn(carNumMax-carNumMin) + carNumMin
	)
	seatResp, err := client.ListTrainSeats(ctx,
		useAt,
		train.Class, train.Name, carNum, departure, arrival, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	validSeats, err := assertListTrainSeats(seatResp)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	reservation, err := client.Reserve(ctx,
		train.Class, train.Name,
		xrandom.GetSeatClass(train.Class, carNum), validSeats,
		departure, arrival, useAt,
		carNum, 1, 1, "isle", nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	cardToken, err := paymentClient.RegistCard(ctx, "11111111", "222", "10/50")
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.CommitReservation(ctx, reservation.ReservationID, cardToken, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListReservations(ctx, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	reservation2, err := client.ShowReservation(ctx, reservation.ReservationID, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if reservation.ReservationID != reservation2.ReservationID {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if err := client.Logout(ctx, nil); err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	return nil
}

// 予約キャンセル含む(Commit後にキャンセル)
func NormalCancelScenario(ctx context.Context) error {
	client, err := isutrain.NewClient()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	paymentClient, err := payment.NewClient()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if config.Debug {
		client.ReplaceMockTransport()
	}

	user, err := xrandom.GetRandomUser()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.Signup(ctx, user.Email, user.Password, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.Login(ctx, user.Email, user.Password, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListStations(ctx, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	var (
		useAt     = xrandom.GetRandomUseAt()
		departure = xrandom.GetRandomStations()
		arrival   = xrandom.GetRandomStations()
	)
	trains, err := client.SearchTrains(ctx, useAt, departure, arrival, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	trainIdx := rand.Intn(len(trains))
	train := trains[trainIdx]
	var (
		carNumMin = 1
		carNumMax = 16
		carNum    = rand.Intn(carNumMax-carNumMin) + carNumMin
	)
	seatResp, err := client.ListTrainSeats(ctx,
		useAt,
		train.Class, train.Name, carNum, departure, arrival, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	validSeats, err := assertListTrainSeats(seatResp)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	reservation, err := client.Reserve(ctx,
		train.Class, train.Name,
		xrandom.GetSeatClass(train.Class, carNum),
		validSeats, departure, arrival, useAt,
		carNum, 1, 1, "isle", nil,
	)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	cardToken, err := paymentClient.RegistCard(ctx, "11111111", "222", "10/50")
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	err = client.CommitReservation(ctx, reservation.ReservationID, cardToken, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	_, err = client.ListReservations(ctx, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	reservation2, err := client.ShowReservation(ctx, reservation.ReservationID, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if reservation.ReservationID != reservation2.ReservationID {
		return bencherror.BenchmarkErrs.AddError(bencherror.NewSimpleCriticalError("予約確認で得られる予約IDが一致していません: got=%d, want=%d", reservation2.ReservationID, reservation.ReservationID))
	}

	err = client.CancelReservation(ctx, reservation.ReservationID, nil)
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if err := client.Logout(ctx, nil); err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	return nil
}

// 曖昧検索シナリオ
func NormalAmbigiousSearchScenario(ctx context.Context) error {
	client, err := isutrain.NewClient()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if config.Debug {
		client.ReplaceMockTransport()
	}

	user, err := xrandom.GetRandomUser()
	if err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if err = client.Signup(ctx, user.Email, user.Password, nil); err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	if err := client.Login(ctx, user.Email, user.Password, nil); err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	reserveResp, err := client.Reserve(ctx,
		"最速", "1", "premium", isutrain.TrainSeats{},
		"東京", "大阪", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		1, 1, 1, "isle", nil)
	if err != nil {
		bencherror.BenchmarkErrs.AddError(err)
	}

	if err := assertReserve(reserveResp); err != nil {
		return bencherror.BenchmarkErrs.AddError(err)
	}

	return nil
}

// オリンピック開催期間に負荷をあげるシナリオ
// FIXME: initializeで指定された日数に応じ、負荷レベルを変化させる
func NormalOlympicParticipantsScenario(ctx context.Context) error {

	// NOTE: webappから見て、明らかに負荷が上がったと感じるレベルに持ってく必要がある
	// NOTE: 指定できる最大の日数で負荷をかける際、飽和しないようにする
	// NOTE:

	return nil
}