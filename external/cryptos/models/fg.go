package models

import "goTradingBot/utils/numeric"

type FearAndGreedChart struct {
	DataList []FearAndGreedDataItem `json:"dataList"`
	// DialConfig       []fearAndGreedDialConfigItemObj `json:"dialConfig"`
	HistoricalValues FearAndGreedHistoricalValues `json:"historicalValues"`
}

type FearAndGreedDataItem struct {
	Score     int    `json:"score"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
	// BTCPrice  string `json:"btcPrice"`
	// BTCVolume string `json:"btcVolume"`
}

type FearAndGreedHistoricalValues struct {
	Now        FearAndGreedDataItem `json:"now"`
	Yesterday  FearAndGreedDataItem `json:"yesterday"`
	LastWeek   FearAndGreedDataItem `json:"lastWeek"`
	LastMonth  FearAndGreedDataItem `json:"lastMonth"`
	YearlyHigh FearAndGreedDataItem `json:"yearlyHigh"`
	YearlyLow  FearAndGreedDataItem `json:"yearlyLow"`
}

func (fgc *FearAndGreedChart) ExtractMetrics() *FearAndGreedMetrics {
	percentChange24h := numeric.DiffPercent(
		fgc.HistoricalValues.Yesterday.Score,
		fgc.HistoricalValues.Now.Score,
	)
	return &FearAndGreedMetrics{
		PercentChange24h: percentChange24h,
		NowScore:         fgc.HistoricalValues.Now.Score,
		YesterdayScore:   fgc.HistoricalValues.Yesterday.Score,
		LastWeekScore:    fgc.HistoricalValues.LastWeek.Score,
		LastMonthScore:   fgc.HistoricalValues.LastMonth.Score,
		YearlyHigh:       fgc.HistoricalValues.YearlyHigh.Score,
		YearlyLow:        fgc.HistoricalValues.YearlyLow.Score,
		Name:             fgc.HistoricalValues.Now.Name,
	}
}

type FearAndGreedMetrics struct {
	PercentChange24h float64 `json:"percentChange24h"`
	NowScore         int     `json:"nowScore"`
	YesterdayScore   int     `json:"yesterdayScore"`
	LastWeekScore    int     `json:"lastWeekScore"`
	LastMonthScore   int     `json:"lastMonthScore"`
	YearlyHigh       int     `json:"yearlyHigh"`
	YearlyLow        int     `json:"yearlyLow"`
	Name             string  `json:"name"`
}
