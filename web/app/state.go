package app

import (
	"goTradingBot/cdl"
	"goTradingBot/external/bybit"
	"goTradingBot/external/cryptos"
	"goTradingBot/predict"
	"goTradingBot/predict/features"
	"sync"
)

var (
	state *appState
	once  sync.Once
)

func initAppState() {
	once.Do(func() {
		state = &appState{
			cryptos:     cryptos.NewClient(),
			cdlProvider: bybit.NewClientFromEnv(bybit.WithCategory("linear")),
			fgModels:    predict.FeaturesGeneratorModels(),
		}
	})
}

type appState struct {
	cryptos     *cryptos.Client
	cdlProvider cdl.CandleProvider
	fgModels    map[predict.Model]*features.Generator
	// mu          sync.Mutex
}
