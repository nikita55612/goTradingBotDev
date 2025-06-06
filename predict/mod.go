package predict

import (
	"goTradingBot/cdl"
	"goTradingBot/predict/features"
	"sync"
)

type Model string

const (
	A6N21P9       Model = "A6N21P9"
	FeatureOffset int   = 9
)

var (
	Models      = [1]Model{A6N21P9}
	fgenerators = make(map[Model]features.GeneratorBuilder)
	mu          sync.Mutex
)

func GetModelOffset(model Model) int {
	switch model {
	case A6N21P9:
		return 9
	default:
		return 0
	}
}

func GetModelWinSize(model Model) int {
	switch model {
	case A6N21P9:
		return 21
	default:
		return 0
	}
}

func FeaturesGeneratorModel(model Model) *features.Generator {
	mu.Lock()
	defer mu.Unlock()
	if fgb, ok := fgenerators[model]; ok {
		return fgb.Build()
	}
	fgb := features.NewGeneratorBuilder()
	if model == A6N21P9 {
		args := []cdl.CandleArg{
			cdl.Open,
			cdl.High,
			cdl.Low,
			cdl.Close,
			cdl.Volume,
			cdl.Turnover,
		}
		fgb = fgb.AddCandleArgs(args, 21, 9)
		fgenerators[model] = fgb
	}
	return fgb.Build()
}

func FeaturesGeneratorModels() map[Model]*features.Generator {
	fModels := make(map[Model]*features.Generator)
	for _, model := range &Models {
		fModels[model] = FeaturesGeneratorModel(model)
	}
	return fModels
}
