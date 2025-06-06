package dataset

import (
	"goTradingBot/cdl"
	"goTradingBot/predict/signals"
	"goTradingBot/utils/numeric"
)

type Filter struct {
	signals []bool
	candles []cdl.Candle
	start   int
	end     int
}

func NewFilter(candles []cdl.Candle, start, end int) *Filter {
	signals := make([]bool, len(candles))
	for i := range signals {
		signals[i] = true
	}
	return &Filter{
		signals: signals,
		candles: candles,
		start:   start,
		end:     end,
	}
}

// FilterPerfectTrendFlat применяет фильтр консолидации к переданным данным
// candles - массив свечей для анализа
// sensitivity - чувствительность фильтра (0.0-1.0). Чем выше, тем больше значений будет отфильтровано
// data - указатели на массивы данных для фильтрации
func (f *Filter) AddPerfectTrendFlatFilter(factor float64) {
	newSignals := identifyPerfectTrendFlat(f.candles, factor)
	for i, s := range newSignals {
		if !s && f.signals[i] {
			f.signals[i] = false
		}
	}
}

func (f *Filter) Apply(matrices ...[][]float64) {
	signals := f.signals[f.start:f.end]
	isKeepAll := true
	for _, v := range signals {
		if !v {
			isKeepAll = false
			break
		}
	}
	if isKeepAll {
		return
	}
	for _, matrix := range matrices {
		for i, v := range matrix {
			if v == nil || len(v) != len(signals) {
				continue
			}
			filtered := make([]float64, 0, len(v))
			for j, keep := range signals {
				if keep {
					filtered = append(filtered, v[j])
				}
			}
			matrix[i] = filtered
		}
	}
}

// identifyConsolidationZones определяет зоны консолидации (бокового движения) на основе сигналов идеального тренда
// candles - массив свечей для анализа
// sensitivity - чувствительность (0.0-1.0) - чем выше, тем больше зон консолидации будет обнаружено
// Возвращает: Массив bool, где true означает трендовое движение, а false - консолидацию
func identifyPerfectTrendFlat(candles []cdl.Candle, factor float64) []bool {
	n := len(candles)
	if n == 0 {
		return nil
	}
	perfectTrend := signals.PerfectTrend(candles, 3)
	ptCandles := make([]cdl.Candle, 0, n/4)
	current := cdl.Candle{
		Time: candles[0].Time,
		O:    candles[0].O,
		H:    candles[0].H,
		L:    candles[0].L,
		C:    candles[0].C,
	}
	for i := 1; i < n; i++ {
		if perfectTrend[i] != perfectTrend[i-1] {
			ptCandles = append(ptCandles, current)
			current = cdl.Candle{
				Time: candles[i].Time,
				O:    candles[i].O,
				H:    candles[i].H,
				L:    candles[i].L,
				C:    candles[i].C,
			}
		} else {
			if candles[i].H > current.H {
				current.H = candles[i].H
			}
			if candles[i].L < current.L {
				current.L = candles[i].L
			}
			current.C = candles[i].C
		}
	}
	ptCandles = append(ptCandles, current)
	n = len(ptCandles)
	zeroSignalsTr := make([]float64, 0, n/4)
	ptSignals := make([]bool, n)
	for i := 1; i < n; i++ {
		curr := ptCandles[i]
		prev := ptCandles[i-1]
		if min(curr.O, curr.C) < prev.L || max(curr.O, curr.C) > prev.H {
			ptSignals[i] = true
			continue
		}
		if tr := ptCandles[i].Arg(cdl.TrueRange); tr > 0 {
			zeroSignalsTr = append(zeroSignalsTr, tr)
		}
	}
	threshold := numeric.Quantile(zeroSignalsTr, factor)
	for i := 1; i < n; i++ {
		if !ptSignals[i] {
			if tr := ptCandles[i].Arg(cdl.TrueRange); tr > threshold {
				ptSignals[i] = true
			}
		}
	}
	n = len(candles)
	ptSignalsIndex := 1
	signals := make([]bool, n)
	for i := 1; i < n-1; i++ {
		if ptSignals[ptSignalsIndex-1] {
			signals[i-1] = true
		}
		if ptCandles[ptSignalsIndex].Time == candles[i].Time {
			if ptSignalsIndex < len(ptSignals)-1 {
				ptSignalsIndex++
			}
		}
	}
	return signals
}
