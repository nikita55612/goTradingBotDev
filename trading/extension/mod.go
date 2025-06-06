package extension

// type DatasetParams struct {
// 	Name                     string       `json:"name"`
// 	RootDir                  string       `json:"rootDir"`
// 	Interval                 cdl.Interval `json:"interval"`
// 	LimitOfInstruments       int          `json:"limitOfInstruments"`
// 	MinInstrumentSecDuration int          `json:"minInstrumentSecDuration"` // 63072
// 	PercInitialMargin        float64      `json:"percInitialMargin"`
// 	IndentationFromEnd       int          `json:"indentationFromEnd"`
// 	FilterPerfectTrendFlat   bool         `json:"filterPerfectTrendFlat"`
// }

// type SampleInfo struct {
// 	Index  int    `json:"index"`
// 	Symbol string `json:"symbol"`
// 	Client string `json:"client"`
// 	XShape [2]int `json:"xShape"`
// 	YShape [2]int `json:"yShape"`
// 	XPath  string `json:"xPath"`
// 	YPath  string `json:"yPath"`
// }

// type DatasetInfo struct {
// 	Params        DatasetParams `json:"params"`
// 	TotalFeatures int           `json:"totalFeatures"`
// 	TotalSignals  int           `json:"totalSignals"`
// 	Signals       []string      `json:"signals"`
// 	Features      []string      `json:"features"`
// 	TotalRows     int           `json:"TotalRows"`
// 	TotalSamples  int           `json:"totalSamples"`
// 	Samples       []SampleInfo  `json:"samples"`
// }

// // CandleProvider определяет интерфейс для работы с поставщиком свечных данных
// type CandleProvider interface {
// 	GetAllCandles(symbol string, interval cdl.Interval) ([]cdl.Candle, error)
// }

// func GenTrainDataset(cp CandleProvider, params DatasetParams) {
// 	cryptosClient := cryptos.NewClient()
// 	cryptoList, err := cryptosClient.GetCryptoList(params.LimitOfInstruments)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	var wg sync.WaitGroup
// 	var mu sync.Mutex

// 	var globalNormAvgRange float64

// 	workers := 4
// 	for j, crypto := range cryptoList {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()

// 			symbol := crypto.Symbol + "USDT"
// 			candles, _ := cp.GetAllCandles(symbol, params.Interval)
// 			if len(candles) == 0 {
// 				return
// 			}

// 			n := len(candles)
// 		}()
// 		wg.Done()
// 	}
// }
