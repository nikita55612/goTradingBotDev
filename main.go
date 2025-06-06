package main

import (
	"context"
	"goTradingBot/cdl"
	"goTradingBot/external/bybit"
	"goTradingBot/external/telebot"
	"goTradingBot/predict"
	"goTradingBot/predict/dataset"
	"goTradingBot/predict/portal"
	"goTradingBot/predict/signals"
	"goTradingBot/trading"
	"goTradingBot/trading/strategies"
	"goTradingBot/utils/slogx"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewContext() (context.Context, context.CancelFunc, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx, stop := signal.NotifyContext(ctx,
		os.Interrupt,
		syscall.SIGTERM,
	)
	return ctx, cancel, stop
}

func GenDataset() {
	params := dataset.DatasetParams{
		Name:                     "linear-trend-H1",
		RootDir:                  "D:/datasets",
		Interval:                 cdl.H1,
		LimitOfInstruments:       290,
		MinInstrumentSecDuration: 15000000, // 163468800
		PercInitialMargin:        0.3,
		IndentationFromEnd:       100,
		FilterPerfectTrendFlat:   true,
	}
	fg := predict.FeaturesGeneratorModel(predict.A6N21P9)
	client := bybit.NewClientFromEnv(bybit.WithCategory("linear"), bybit.WithTimeout(30*time.Second))
	sgb := signals.NewGeneratorBuilder()
	sgb = sgb.AddPerfectTrend(4)
	sgb = sgb.AddNextPerfectTrend(4)
	sgb = sgb.AddPerfectTrend(9)
	sgb = sgb.AddNextPerfectTrend(9)
	dataset.CreateDataset(client, params, fg, sgb.Build())
}

func Run(ctx context.Context) {
	if err := portal.StartWithContext(ctx); err != nil {
		log.Fatal(err)
	}

	cli := bybit.NewClientFromEnv(
		// bybit.WithContext(ctx),
		bybit.WithCategory("linear"),
		bybit.WithTimeout(3*time.Second),
	)
	logger := slog.New(slogx.Fanout(
		slog.NewJSONHandler(os.Stdout, nil),
		telebot.NewBotSlogHandlerFromEnv("500295076", nil),
	))
	bot := trading.NewTradingBot(
		ctx,
		cli.TradingClientImpl(),
		cli.DataProviderImpl(),
		logger,
		nil,
	)

	bot.AddStrategys(
		strategies.NewStrategy(
			"HYPEUSDT", cdl.M5,
			"xgb_linear-M5_PerfectTrend-p4",
			15, 0.6, 0.02,
		),
	)
}

func main() {
	ctx, cancel, stop := NewContext()
	defer func() {
		cancel()
		stop()
	}()
	go Run(ctx)
	<-ctx.Done()
	time.Sleep(5 * time.Second)
}
