package main

import (
	"context"
	"fmt"
	"goTradingBot/cdl"
	"goTradingBot/external/bybit"
	"goTradingBot/external/cryptos"
	"goTradingBot/external/cryptos/db"
	"goTradingBot/external/telebot"
	"goTradingBot/httpx"
	"goTradingBot/predict"
	"goTradingBot/predict/portal"
	"goTradingBot/utils/numeric"
	"goTradingBot/utils/saveform"
	"goTradingBot/web/app"
	"log/slog"
	"testing"
	"time"
)

func TestUpdCryptosDB(t *testing.T) {
	cryptosClient := cryptos.NewClient()

	cryptoList, err := cryptosClient.GetCryptoList(3)
	if err != nil {
		t.Fatal(err)
	}
	for _, crypto := range cryptoList {
		logoUrl := fmt.Sprintf("https://s2.coinmarketcap.com/static/img/coins/64x64/%d.png", crypto.ID)
		res := httpx.Get(logoUrl).Do()
		logo, err := res.ReadBody()
		if err != nil {
			continue
		}
		res.Close()
		dbCrypto := &db.Crypto{
			ID:     crypto.ID,
			Name:   crypto.Name,
			Symbol: crypto.Symbol,
			Logo:   logo,
		}
		db.InsertCrypto(dbCrypto)
	}
}

func TestCandles(t *testing.T) {
	client := bybit.NewClientFromEnv()
	symbol := "BTC"
	candles, _ := client.GetAllCandles(symbol+"USDT", cdl.H1)
	fmt.Println(len(candles))
	candles = candles[len(candles)-2000:]

	candlesMap := make(map[string][]float64)

	candlesMap["open"] = cdl.ListOfCandleArg(candles, cdl.Open)
	candlesMap["high"] = cdl.ListOfCandleArg(candles, cdl.High)
	candlesMap["low"] = cdl.ListOfCandleArg(candles, cdl.Low)
	candlesMap["close"] = cdl.ListOfCandleArg(candles, cdl.Close)

	saveform.ToCSV("candles.csv", candlesMap)
}

func TestSignals(t *testing.T) {
	portal.Start()

	client := bybit.NewClientFromEnv()
	symbol := "ETH"
	candles, _ := client.GetAllCandles(symbol+"USDT", cdl.M15)
	n := len(candles)

	features := predict.FeaturesGeneratorModel(predict.A6N21P9).
		GenTranspose(candles, n-2000, -1)

	res, err := portal.GetPrediction(features, "M15", "p4").Unwrap()
	if err != nil {
		return
	}
	signal, ok := res["xgb_TrendV1-M15_PerfectTrend-p4"]
	if !ok {
		return
	}
	data := cdl.CandlesAsMap(candles[n-2000:])
	data["signal"] = signal

	fmt.Println(len(data["close"]))
	fmt.Println(len(signal))

	portal.Stop()
	saveform.ToCSV("signals.csv", data)
}

func TestCandleSync(t *testing.T) {
	provider := bybit.NewClientFromEnv()

	ctx, cancel := context.WithCancel(context.Background())
	synchronizer := cdl.NewCandleSync(ctx, "BTCUSDT", cdl.M1, 10, provider)
	err := synchronizer.StartSync()
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan *cdl.CandleStreamData)
	go func() {
		for data := range ch {
			fmt.Printf("%+v\n", data)
		}
	}()
	done := synchronizer.Subscribe(ch)

	time.Sleep(time.Minute)

	close(done)
	cancel()
}

func TestClient(t *testing.T) {
	client := bybit.NewClientFromEnv(
		bybit.WithCategory("linear"),
		bybit.WithTimeout(3*time.Second),
	)
	price := 100.
	order, _ := client.PlaceOrder("HYPEUSDT", 0.18, &price)
	time.Sleep(time.Second)
	detail, _ := client.GetOrderHistoryDetail(order.OrderId)

	fmt.Printf("%+v\n", detail)
}

func TestTelebot(t *testing.T) {
	bot := telebot.NewBotFromEnv(telebot.WithWriteChatID("500295076"))
	logger := slog.New(slog.NewTextHandler(bot, nil))
	logger.Info("hello", "Info message", "...")
	logger.Error("hello", "Error message", "...")
	logger.Warn("hello", "Warn message", "...")
}

func TestPy(t *testing.T) {
	portal.SetAddr("localhost:8656")
	portal.Start()

	fg := predict.FeaturesGeneratorModel(predict.A6N21P9)
	client := bybit.NewClientFromEnv()
	symbol := "BTC"
	candles, _ := client.GetAllCandles(symbol+"USDT", cdl.H1)

	n := len(candles)
	features := fg.GenTranspose(candles, n-5, n)

	pred, err := portal.GetPrediction(features, "H1").Unwrap()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(pred)

	portal.Stop()
}

func TestVote(t *testing.T) {
	client := bybit.NewClientFromEnv()
	ctx, done := context.WithCancel(context.Background())
	candleSync := cdl.NewCandleSync(ctx, "BTCUSDT", cdl.M1, 12, client)
	candleSync.StartSync()
	candles := candleSync.GetCandles(3)
	cc, _ := client.GetCandles("BTCUSDT", cdl.M1, 3)
	fmt.Printf("cc %+v\n", cc)
	fmt.Printf("%+v\n\n", candles)
	time.Sleep(time.Second)
	candles = candleSync.GetCandles(3)
	cc, _ = client.GetCandles("BTCUSDT", cdl.M1, 3)
	fmt.Printf("cc %+v\n", cc)
	fmt.Printf("%+v\n\n", candles)
	time.Sleep(time.Second)
	candles = candleSync.GetCandles(3)
	cc, _ = client.GetCandles("BTCUSDT", cdl.M1, 3)
	fmt.Printf("cc %+v\n", cc)
	fmt.Printf("%+v\n\n", candles)
	time.Sleep(time.Second)
	candles = candleSync.GetCandles(3)
	cc, _ = client.GetCandles("BTCUSDT", cdl.M1, 3)
	fmt.Printf("cc %+v\n\n", cc)
	fmt.Printf("%+v\n", candles)
	done()
}

func TestPreds(t *testing.T) {
	portal.Start()
	client := bybit.NewClientFromEnv()
	candles, _ := client.GetCandles("ETHUSDT", cdl.M15, 7000)

	fg := predict.FeaturesGeneratorModel(predict.A6N21P9)
	features := fg.GenTranspose(candles, predict.FeatureOffset, -1)

	prediction, err := portal.GetPrediction(features, "M15").Unwrap()
	if err != nil {
		t.Fatal(err)
	}

	pt4 := make([]float64, len(candles))
	npt7 := make([]float64, len(candles))
	copy(pt4[predict.FeatureOffset:], prediction["xgb_TrendV1-M15_PerfectTrend-p4"])
	copy(npt7[predict.FeatureOffset:], prediction["xgb_TrendV1-M15_NextPerfectTrend-p7"])

	table := map[string][]float64{
		"price": cdl.ListOfCandleArg(candles, cdl.Close),
		"pt4":   pt4,
		"npt7":  npt7,
	}
	saveform.ToCSV("test.csv", table)
	portal.Stop()
}

func TestRunTerminal(t *testing.T) {
	portal.Start()
	defer portal.Stop()
	app.RunTerminal(":7788")
}

func TestRunOrderLog(t *testing.T) {
	app.RunOrderLog(":7789")
}

func TestTemp(t *testing.T) {
	portal.Start()
	defer portal.Stop()

	client := bybit.NewClientFromEnv(bybit.WithCategory("linear"))
	candles, _ := client.GetCandles("HYPEUSDT", cdl.M5, 1800)

	fg := predict.FeaturesGeneratorModel(predict.A6N21P9)
	features := fg.GenTranspose(candles, predict.FeatureOffset, -1)

	pred, err := portal.GetPrediction(
		features,
		"xgb_linear-M5_PerfectTrend-p4",
	).UnwrapSinglePredict()
	if err != nil {
		t.Fatal(err)
	}

	candles = candles[predict.FeatureOffset:]

	long := make([]float64, 0, len(pred)/4)
	short := make([]float64, 0, len(pred)/4)
	var entryPrice float64

	for i := 1; i < len(pred)-1; i++ {
		p := pred[i]
		pp := pred[i-1]
		if p > 0.5 && pp < 0.5 {
			if entryPrice > 0 {
				fmt.Println(entryPrice, candles[i+1].O, numeric.DiffPercent(entryPrice, candles[i+1].O))
				dp := numeric.DiffPercent(entryPrice, candles[i+1].O)
				if dp > 0 {
					short = append(short, 0)
				} else {
					short = append(short, 1)
				}
			}
			entryPrice = candles[i+1].O
		}
		if p < 0.5 && pp > 0.5 {
			if entryPrice > 0 {
				dp := numeric.DiffPercent(entryPrice, candles[i+1].O)
				if dp < 0 {
					long = append(long, 0)
				} else {
					long = append(long, 1)
				}
			}
			entryPrice = candles[i+1].O
		}
	}
	pos := 0
	for _, v := range short {
		if v == 1 {
			pos++
		}
	}
	fmt.Println(pos, len(short), numeric.DiffPercent(pos, len(short)))

	fmt.Println(pred)
	fmt.Println(len(pred))
	fmt.Println(len(candles[predict.FeatureOffset:]))
}
