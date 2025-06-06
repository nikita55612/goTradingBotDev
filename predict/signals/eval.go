package signals

import "goTradingBot/cdl"

// PerfectTrend определяет "идеальные" точки входа в long/short на основе фракталов.
// Функция ищет последовательные максимумы (для short) и минимумы (для long),
// отфильтровывая менее значимые экстремумы.
// Возвращает массив, где:
// 0 - сигнал к long (после бычьего фрактала)
// 1 - сигнал к short (после медвежьего фрактала)
// period - количество свечей по обе стороны, которые должны быть ниже/выше фрактала
func PerfectTrend(candles []cdl.Candle, period int) []float64 {
	n := len(candles)
	if n == 0 || period < 2 || n <= 2*period {
		return nil
	}
	upFractals := make([]bool, n)
	downFractals := make([]bool, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	for i := 0; i < n; i++ {
		highs[i] = candles[i].Arg(cdl.CH)
		lows[i] = candles[i].Arg(cdl.CL)
	}
	var lastFr, lastFrIndex int
	var lastV float64
	for i := period; i < n-period; i++ {
		highV := highs[i]
		lowV := lows[i]
		isUpFractal := true
		for j := 1; j <= period && isUpFractal; j++ {
			if highs[i-j] >= highV || highs[i+j] >= highV {
				isUpFractal = false
			}
		}
		if isUpFractal {
			if lastFr == 1 {
				if highV >= lastV {
					upFractals[lastFrIndex] = false
				} else {
					isUpFractal = false
				}
			}
			if isUpFractal {
				upFractals[i] = true
				lastFr = 1
				lastV = highV
				lastFrIndex = i
			}
		}
		isDownFractal := true
		for j := 1; j <= period && isDownFractal; j++ {
			if lows[i-j] <= lowV || lows[i+j] <= lowV {
				isDownFractal = false
			}
		}
		if isDownFractal {
			if lastFr == -1 {
				if lowV <= lastV {
					downFractals[lastFrIndex] = false
				} else {
					isDownFractal = false
				}
			}
			if isDownFractal {
				downFractals[i] = true
				lastFr = -1
				lastV = lowV
				lastFrIndex = i
			}
		}
	}
	highs, lows = nil, nil
	firstSignal := 0
	for i := 0; firstSignal == 0 && i < n; i++ {
		if upFractals[i] {
			firstSignal++
		}
		if downFractals[i] {
			firstSignal--
		}
	}
	lastIsUp := false
	if firstSignal == 1 {
		lastIsUp = true
	}
	signals := make([]float64, n)
	for i := 1; i < n; i++ {
		if upFractals[i-1] {
			lastIsUp = true
		}
		if downFractals[i-1] {
			lastIsUp = false
		}
		if lastIsUp {
			signals[i] = 0
			continue
		}
		signals[i] = 1
	}
	return signals
}

// NextDirSignal возвращает сигнал (1 или 0) в зависимости от направления следующей свечи:
// 1 - если следующая свеча бычья (закрытие > открытия), 0 - если медвежья или нейтральная.
func NextDirSignal(candles []cdl.Candle) []float64 {
	n := len(candles)
	if n == 0 {
		return nil
	}
	signals := make([]float64, n)
	for i := 0; i < n-1; i++ {
		dir := candles[i+1].Arg(cdl.Direction)
		if dir > 0 {
			signals[i] = 1
			continue
		}
		signals[i] = 0
	}
	return signals
}

// NextBodyWiderSignal возвращает 1, если тело следующей свечи шире, чем текущий TrueRange (волатильность растёт).
// 0 - если тело уже или равно текущему диапазону.
func NextBodyWiderSignal(candles []cdl.Candle) []float64 {
	n := len(candles)
	if n == 0 {
		return nil
	}
	signals := make([]float64, n)
	for i := 0; i < n-1; i++ {
		tr := candles[i].Arg(cdl.TrueRange)
		nextBody := candles[i+1].Arg(cdl.Body)
		if nextBody > tr {
			signals[i] = 1
			continue
		}
		signals[i] = 0
	}
	return signals
}

// NextCandleOutsideSignal возвращает 1, если следующая свеча выходит за границы текущей (внешний бар).
// 0 - если тело полностью находится внутри.
func NextCandleOutsideSignal(candles []cdl.Candle) []float64 {
	n := len(candles)
	if n == 0 {
		return nil
	}
	signals := make([]float64, n)
	for i := 0; i < n-1; i++ {
		nextCandle := candles[i+1]
		nextMin := min(nextCandle.O, nextCandle.C)
		nextMax := max(nextCandle.O, nextCandle.C)
		if nextMin < candles[i].L || nextMax > candles[i].H {
			signals[i] = 1
			continue
		}
		signals[i] = 0
	}
	return signals
}
