package dataset

import (
	"encoding/json"
	"fmt"
	"goTradingBot/cdl"
	"goTradingBot/external/cryptos"
	"goTradingBot/predict/features"
	"goTradingBot/predict/signals"
	"goTradingBot/utils"
	"goTradingBot/utils/numeric"
	"goTradingBot/utils/saveform"
	"log"
	"log/slog"
	"os"
	"path"
	"sync"
)

type DatasetParams struct {
	Name                     string       `json:"name"`
	RootDir                  string       `json:"rootDir"`
	Interval                 cdl.Interval `json:"interval"`
	LimitOfInstruments       int          `json:"limitOfInstruments"`
	MinInstrumentSecDuration int          `json:"minInstrumentSecDuration"` // 63072
	PercInitialMargin        float64      `json:"percInitialMargin"`
	IndentationFromEnd       int          `json:"indentationFromEnd"`
	FilterPerfectTrendFlat   bool         `json:"filterPerfectTrendFlat"`
}

type SampleInfo struct {
	Index  int    `json:"index"`
	Symbol string `json:"symbol"`
	Client string `json:"client"`
	XShape [2]int `json:"xShape"`
	YShape [2]int `json:"yShape"`
	XPath  string `json:"xPath"`
	YPath  string `json:"yPath"`
}

type DatasetInfo struct {
	Params        DatasetParams `json:"params"`
	TotalFeatures int           `json:"totalFeatures"`
	TotalSignals  int           `json:"totalSignals"`
	Signals       []string      `json:"signals"`
	Features      []string      `json:"features"`
	TotalRows     int           `json:"TotalRows"`
	TotalSamples  int           `json:"totalSamples"`
	Samples       []SampleInfo  `json:"samples"`
}

// CandleProvider определяет интерфейс для работы с поставщиком свечных данных
type CandleProvider interface {
	GetAllCandles(symbol string, interval cdl.Interval) ([]cdl.Candle, error)
}

func CreateDataset(cp CandleProvider, params DatasetParams, fg *features.Generator, sg *signals.Generator) {
	datasetPath := path.Join(params.RootDir, params.Name)
	if utils.PathExists(datasetPath) {
		log.Fatalf("Dataset %s already exists...", datasetPath)
	}
	if err := os.MkdirAll(datasetPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	datasetSamples := path.Join(datasetPath, "samples")
	if err := os.MkdirAll(datasetSamples, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	datasetInfo := new(DatasetInfo)
	datasetInfo.Params = params
	datasetInfo.Features = fg.Labels()
	datasetInfo.Signals = sg.Labels()
	datasetInfo.TotalFeatures = len(datasetInfo.Features)
	datasetInfo.TotalSignals = len(datasetInfo.Signals)

	cryptosClient := cryptos.NewClient()
	cryptoList, err := cryptosClient.GetCryptoList(params.LimitOfInstruments)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	var globalNormAvgRange float64

	workers := 4
	for j, crypto := range cryptoList {
		wg.Add(1)
		go func() {
			defer wg.Done()

			symbol := crypto.Symbol + "USDT"
			candles, _ := cp.GetAllCandles(symbol, params.Interval)
			if len(candles) == 0 {
				return
			}

			n := len(candles)

			start := int(float64(n) * params.PercInitialMargin)
			end := n - params.IndentationFromEnd

			// Отсев ----------------------------

			timeDiff := (candles[n-1].Time - candles[0].Time) / 1000
			if timeDiff < int64(params.MinInstrumentSecDuration) {
				slog.Info("time filter", "symbol", symbol, "timeDiff", timeDiff)
				return
			}

			trList := cdl.ListOfCandleArg(candles[start:end], cdl.NormalizedRange)
			normAvgRange := numeric.Avg(trList)
			mu.Lock()
			if j == 0 {
				globalNormAvgRange = normAvgRange
			}
			if normAvgRange < globalNormAvgRange/2 {
				mu.Unlock()
				slog.Info("avgRange filter", "symbol", symbol)
				return
			}
			globalNormAvgRange = (globalNormAvgRange + normAvgRange) / 2
			perfectTrendFlatFilterFactor := 0.5
			if normAvgRange < globalNormAvgRange*0.8 {
				perfectTrendFlatFilterFactor = 0.7
			}
			mu.Unlock()

			// --------------------------------------

			features := fg.Gen(candles, start, end)
			signals := sg.Gen(candles, start, end)

			filter := NewFilter(candles, start, end)
			if params.FilterPerfectTrendFlat {
				filter.AddPerfectTrendFlatFilter(perfectTrendFlatFilterFactor)
			}
			filter.Apply(features, signals)

			itemPath := path.Join(datasetSamples, fmt.Sprintf("%d-%s-bybit", j+1, crypto.Symbol))
			if err := os.MkdirAll(itemPath, os.ModePerm); err != nil {
				return
			}

			XPath := path.Join(itemPath, "X.csv")
			yPath := path.Join(itemPath, "y.csv")

			sampleInfo := SampleInfo{
				Index:  j + 1,
				Symbol: crypto.Symbol,
				Client: "bybit",
				XShape: [2]int{len(features[0]), len(features)},
				YShape: [2]int{len(signals[0]), len(signals)},
				XPath:  XPath,
				YPath:  yPath,
			}

			mu.Lock()
			datasetInfo.Samples = append(datasetInfo.Samples, sampleInfo)
			datasetInfo.TotalRows += sampleInfo.XShape[0]
			mu.Unlock()

			sampleData, _ := json.MarshalIndent(&sampleInfo, "", "    ")
			fmt.Println(string(sampleData))
			fmt.Printf("Saving %d %s...\n", j+1, crypto.Symbol)

			saveform.ColumnsToCSV(XPath, features, nil)
			saveform.ColumnsToCSV(yPath, signals, nil)
		}()
		if j == 0 {
			wg.Wait()
		} else if (j+1)%workers == 0 {
			wg.Wait()
		}
	}
	wg.Wait()
	datasetInfo.TotalSamples = len(datasetInfo.Samples)
	datasetInfoPath := path.Join(datasetPath, "metadata.json")
	saveform.ToJSON(datasetInfoPath, datasetInfo)
}
