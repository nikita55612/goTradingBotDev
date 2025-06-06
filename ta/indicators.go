package ta

import (
	"goTradingBot/cdl"
	"goTradingBot/utils/numeric"
	"math"
	"slices"
)

type AdxDi struct {
	ADX        []float64
	DiPlus     []float64
	DiMinus    []float64
	Len        int
	Period     int
	W          float64
	alpha      float64
	PrevCandle cdl.Candle
	dmPlus     float64
	dmMinus    float64
	atr        float64
}

func (a *AdxDi) Next(candle cdl.Candle) {
	tr := candle.Ratio(cdl.TrueRangeRatio, &a.PrevCandle)
	a.atr = tr*a.alpha + a.atr*(1-a.alpha)
	highDif := candle.H - a.PrevCandle.H
	lowDif := a.PrevCandle.L - candle.L
	if highDif >= lowDif {
		a.dmMinus = 0*a.alpha + a.dmMinus*(1-a.alpha)
		a.dmPlus = max(0, highDif)*a.alpha + a.dmPlus*(1-a.alpha)
	} else {
		a.dmPlus = 0*a.alpha + a.dmPlus*(1-a.alpha)
		a.dmMinus = max(0, lowDif)*a.alpha + a.dmMinus*(1-a.alpha)
	}
	if a.atr == 0 {
		a.DiPlus = append(a.DiPlus, 0)
		a.DiMinus = append(a.DiMinus, 0)
	} else {
		a.DiPlus = append(a.DiPlus, a.dmPlus/a.atr)
		a.DiMinus = append(a.DiMinus, a.dmMinus/a.atr)
	}
	dx := math.Abs(a.DiPlus[a.Len-1]-a.DiMinus[a.Len-1]) /
		math.Abs(a.DiPlus[a.Len-1]+a.DiMinus[a.Len-1])
	a.ADX = append(a.ADX, dx*a.alpha+a.ADX[a.Len-1]*(1-a.alpha))
	a.PrevCandle = candle
	a.Len++
}

func NewAdxDi(candles []cdl.Candle, period int, w float64) *AdxDi {
	var dmPlus float64
	var dmMinus float64

	diPlus := make([]float64, len(candles))
	diMinus := make([]float64, len(candles))
	diPlus[0], diMinus[0] = 0, 0

	adx := make([]float64, len(candles))
	adx[0] = 50

	atr := candles[0].Ratio(cdl.TrueRangeRatio, &candles[0])
	alpha := w / (float64(period) + w - 1)
	PrevCandle := candles[0]
	for i := 1; i < len(candles); i++ {
		candle := candles[i]
		tr := candle.Ratio(cdl.TrueRangeRatio, &PrevCandle)
		atr = tr*alpha + atr*(1-alpha)

		highDif := candle.H - PrevCandle.H
		lowDif := PrevCandle.L - candle.L

		if highDif >= lowDif {
			dmMinus = 0*alpha + dmMinus*(1-alpha)
			dmPlus = max(0, highDif)*alpha + dmPlus*(1-alpha)
		} else {
			dmPlus = 0*alpha + dmPlus*(1-alpha)
			dmMinus = max(0, lowDif)*alpha + dmMinus*(1-alpha)
		}
		if atr == 0 {
			diPlus[i], diMinus[i] = 0, 0
		} else {
			diPlus[i] = dmPlus / atr
			diMinus[i] = dmMinus / atr
		}
		dx := math.Abs(diPlus[i]-diMinus[i]) / math.Abs(diPlus[i]+diMinus[i])
		adx[i] = dx*alpha + adx[i-1]*(1-alpha)
		PrevCandle = candle
	}
	return &AdxDi{
		ADX:        adx,
		DiPlus:     diPlus,
		DiMinus:    diMinus,
		Len:        len(adx),
		Period:     period,
		W:          w,
		alpha:      alpha,
		PrevCandle: PrevCandle,
		dmPlus:     dmPlus,
		dmMinus:    dmMinus,
		atr:        atr,
	}
}

type SuperTrend struct {
	LongStop   []float64
	ShortStop  []float64
	Len        int
	Period     int
	CandleArg  cdl.CandleArg
	Factor     float64
	W          float64
	alpha      float64
	PrevCandle cdl.Candle
	atr        float64
}

func (st *SuperTrend) Next(candle cdl.Candle) {
	tr := candle.Ratio(cdl.TrueRangeRatio, &st.PrevCandle)
	st.atr = tr*st.alpha + st.atr*(1-st.alpha)
	src := candle.Arg(st.CandleArg)
	if st.PrevCandle.C > st.LongStop[st.Len-1] {
		st.LongStop = append(st.LongStop, max(src-st.atr*st.Factor, st.LongStop[st.Len-1]))
	} else {
		st.LongStop = append(st.LongStop, src-st.atr*st.Factor)
	}
	if st.PrevCandle.C < st.ShortStop[st.Len-1] {
		st.ShortStop = append(st.ShortStop, min(src+st.atr*st.Factor, st.ShortStop[st.Len-1]))
	} else {
		st.ShortStop = append(st.ShortStop, src+st.atr*st.Factor)
	}
	st.PrevCandle = candle
	st.Len++
}

func NewSuperTrend(candles []cdl.Candle, period int, arg cdl.CandleArg, factor float64, w float64) *SuperTrend {
	if len(candles) == 0 || period == 0 || w == 0 {
		return nil
	}
	longStop := make([]float64, len(candles))
	shortStop := make([]float64, len(candles))

	atr := candles[0].Ratio(cdl.TrueRangeRatio, &candles[0])
	longStop[0] = candles[0].C - atr*factor
	shortStop[0] = candles[0].C + atr*factor

	alpha := w / (float64(period) + w - 1)
	PrevCandle := candles[0]
	for i := 1; i < len(candles); i++ {
		candle := candles[i]
		tr := candle.Ratio(cdl.TrueRangeRatio, &PrevCandle)
		atr = tr*alpha + atr*(1-alpha)

		src := candle.Arg(arg)
		if PrevCandle.C > longStop[i-1] {
			longStop[i] = max(src-atr*factor, longStop[i-1])
		} else {
			longStop[i] = src - atr*factor
		}
		if PrevCandle.C < shortStop[i-1] {
			shortStop[i] = min(src+atr*factor, shortStop[i-1])
		} else {
			shortStop[i] = src + atr*factor
		}
		PrevCandle = candle
	}
	return &SuperTrend{
		LongStop:   longStop,
		ShortStop:  shortStop,
		Len:        len(longStop),
		Period:     period,
		CandleArg:  arg,
		Factor:     factor,
		W:          w,
		alpha:      alpha,
		PrevCandle: PrevCandle,
		atr:        atr,
	}
}

type ChandelierExit struct {
	LongStop   []float64
	ShortStop  []float64
	Len        int
	Period     int
	Factor     float64
	W          float64
	alpha      float64
	PrevCandle cdl.Candle
	highPer    []float64
	lowPer     []float64
	atr        float64
}

func (ce *ChandelierExit) Next(candle cdl.Candle) {
	tr := candle.Ratio(cdl.TrueRangeRatio, &ce.PrevCandle)
	ce.atr = tr*ce.alpha + ce.atr*(1-ce.alpha)
	ce.highPer = append(ce.highPer, candle.H)
	ce.lowPer = append(ce.lowPer, candle.L)
	if ce.Len >= ce.Period {
		ce.highPer = ce.highPer[1:]
		ce.lowPer = ce.lowPer[1:]
	}
	maxHigh := slices.Max(ce.highPer)
	minLow := slices.Min(ce.lowPer)
	if ce.PrevCandle.C > ce.LongStop[ce.Len-1] {
		ce.LongStop = append(ce.LongStop, max(maxHigh-ce.atr*ce.Factor, ce.LongStop[ce.Len-1]))
	} else {
		ce.LongStop = append(ce.LongStop, maxHigh-ce.atr*ce.Factor)
	}
	if ce.PrevCandle.C < ce.ShortStop[ce.Len-1] {
		ce.ShortStop = append(ce.ShortStop, min(minLow+ce.atr*ce.Factor, ce.ShortStop[ce.Len-1]))
	} else {
		ce.ShortStop = append(ce.ShortStop, minLow+ce.atr*ce.Factor)
	}
	ce.PrevCandle = candle
	ce.Len++
}

func NewChandelierExit(candles []cdl.Candle, period int, factor float64, w float64) *ChandelierExit {
	n := len(candles)
	if n == 0 || period <= 0 || w == 0 {
		return nil
	}
	longStop := make([]float64, n)
	shortStop := make([]float64, n)

	highPer := make([]float64, 0, period+1)
	lowPer := make([]float64, 0, period+1)

	highPer[0] = candles[0].H
	lowPer[0] = candles[0].L

	atr := candles[0].Ratio(cdl.TrueRangeRatio, &candles[0])
	longStop[0] = highPer[0] - atr*factor
	shortStop[0] = lowPer[0] + atr*factor

	alpha := w / (float64(period) + w - 1)
	PrevCandle := candles[0]
	for i := 1; i < n; i++ {
		candle := candles[i]
		tr := candle.Ratio(cdl.TrueRangeRatio, &PrevCandle)
		atr = tr*alpha + atr*(1-alpha)

		highPer = append(highPer, candle.H)
		lowPer = append(lowPer, candle.L)
		if i >= period {
			highPer = highPer[1:]
			lowPer = lowPer[1:]
		}
		maxHigh := slices.Max(highPer)
		minLow := slices.Min(lowPer)

		if PrevCandle.C > longStop[i-1] {
			longStop[i] = max(maxHigh-atr*factor, longStop[i-1])
		} else {
			longStop[i] = maxHigh - atr*factor
		}
		if PrevCandle.C < shortStop[i-1] {
			shortStop[i] = min(minLow+atr*factor, shortStop[i-1])
		} else {
			shortStop[i] = minLow + atr*factor
		}
		PrevCandle = candle
	}
	return &ChandelierExit{
		LongStop:   longStop,
		ShortStop:  shortStop,
		Len:        len(longStop),
		Period:     period,
		Factor:     factor,
		W:          w,
		alpha:      alpha,
		PrevCandle: PrevCandle,
		highPer:    highPer,
		lowPer:     lowPer,
		atr:        atr,
	}
}

type BollingerBands struct {
	Len        int
	UpperBand  []float64
	MiddleBand []float64
	LowerBand  []float64
	Period     int
	CandleArg  cdl.CandleArg
	StdDevMult float64
	sum        float64
	sumSq      float64
	middleMA   MovingAverage
	maT        MaType
}

func (bb *BollingerBands) Next(candles []cdl.Candle) {
	n := len(candles)
	if n < 2 || bb.Len == 0 {
		return
	}
	price := candles[n-1].Arg(bb.CandleArg)
	oldPrice := candles[n-min(n, bb.Period+1)].Arg(bb.CandleArg)

	bb.sum += price - oldPrice
	bb.sumSq += price*price - oldPrice*oldPrice
	mean := bb.sum / float64(bb.Period)
	variance := (bb.sumSq / float64(bb.Period)) - (mean * mean)
	multStdDev := math.Sqrt(variance) * bb.StdDevMult

	bb.middleMA.Next(candles)
	bb.MiddleBand = append(bb.MiddleBand, bb.middleMA.Last())

	bb.UpperBand = append(bb.UpperBand, bb.MiddleBand[bb.Len]+multStdDev)
	bb.LowerBand = append(bb.LowerBand, bb.MiddleBand[bb.Len]-multStdDev)
	bb.Len++
	if bb.Len%bb.Period == 0 {
		bb.middleMA.Crop()
	}
}

func NewBollingerBands(candles []cdl.Candle, arg cdl.CandleArg, maT MaType, period int, mult float64) *BollingerBands {
	n := len(candles)
	if n == 0 || period <= 0 {
		return nil
	}
	var middleMA MovingAverage
	var middleBand []float64
	switch maT {
	case S:
		middleMA = NewSMA(candles, arg, period)
		middleBand = middleMA.(*SMA[float64]).Res
	case E:
		middleMA = NewEMA(candles, arg, period, 2)
		middleBand = middleMA.(*EMA[float64]).Res
	case VW:
		middleMA = NewVWMA(candles, arg, period)
		middleBand = middleMA.(*VWMA).Res
	default:
		middleMA = NewVWMA(candles, arg, period)
		middleBand = middleMA.(*VWMA).Res
	}
	middleMA.Crop()
	upperBand := make([]float64, n)
	lowerBand := make([]float64, n)
	var sum, sumSq float64
	for i := 0; i < period && i < n; i++ {
		price := candles[i].Arg(arg)
		sum += price
		sumSq += price * price
		mean := sum / float64(i+1)
		variance := (sumSq / float64(i+1)) - (mean * mean)
		multStdDev := math.Sqrt(variance) * mult

		upperBand[i] = middleBand[i] + multStdDev
		lowerBand[i] = middleBand[i] - multStdDev
	}
	for i := period; i < n; i++ {
		price := candles[i].Arg(arg)
		oldPrice := candles[i-period].Arg(arg)
		sum += price - oldPrice
		sumSq += price*price - oldPrice*oldPrice
		mean := sum / float64(period)
		variance := (sumSq / float64(period)) - (mean * mean)
		multStdDev := math.Sqrt(variance) * mult

		upperBand[i] = middleBand[i] + multStdDev
		lowerBand[i] = middleBand[i] - multStdDev
	}
	return &BollingerBands{
		Len:        len(middleBand),
		UpperBand:  upperBand,
		MiddleBand: middleBand,
		LowerBand:  lowerBand,
		Period:     period,
		CandleArg:  arg,
		StdDevMult: mult,
		sum:        sum,
		sumSq:      sumSq,
		middleMA:   middleMA,
		maT:        maT,
	}
}

type RSI struct {
	Len       int
	Res       []float64
	Period    int
	CandleArg cdl.CandleArg
	avgGain   float64
	avgLoss   float64
}

func (r *RSI) Next(candles []cdl.Candle) {
	n := len(candles)
	if n < 2 || r.Len == 0 {
		return
	}
	priceDiff := candles[n-1].Arg(r.CandleArg) - candles[n-2].Arg(r.CandleArg)
	var gain, loss float64
	if priceDiff > 0 {
		gain = priceDiff
	} else {
		loss = math.Abs(priceDiff)
	}
	r.avgGain = (r.avgGain*(float64(r.Period)-1) + gain) / float64(r.Period)
	r.avgLoss = (r.avgLoss*(float64(r.Period)-1) + loss) / float64(r.Period)
	if r.avgLoss == 0 {
		r.Res = append(r.Res, 1)
	} else {
		rs := r.avgGain / r.avgLoss
		rsi := 1 - (1 / (1 + rs))
		r.Res = append(r.Res, rsi)
	}
	r.Len++
}

func NewRSI(candles []cdl.Candle, arg cdl.CandleArg, period int) *RSI {
	n := len(candles)
	if n < 2 || period <= 0 {
		return nil
	}
	res := make([]float64, n)
	var avgGain, avgLoss float64
	priceDiff := candles[1].Arg(arg) - candles[0].Arg(arg)
	if priceDiff > 0 {
		avgGain += priceDiff
	} else {
		avgLoss += math.Abs(priceDiff)
	}
	for i := 1; i < period && i < n; i++ {
		priceDiff := candles[i].Arg(arg) - candles[i-1].Arg(arg)
		if priceDiff > 0 {
			avgGain += (priceDiff - avgGain) / float64(i+1)
		} else {
			avgLoss += (math.Abs(priceDiff) - avgLoss) / float64(i+1)
		}
		if avgLoss == 0 {
			res[i] = 1
		} else {
			rs := avgGain / avgLoss
			rsi := 1 - (1 / (1 + rs))
			res[i] = rsi
		}
	}
	pv, pf := float64(period-1), float64(period)
	for i := period; i < n; i++ {
		priceDiff := candles[i].Arg(arg) - candles[i-1].Arg(arg)
		var gain, loss float64
		if priceDiff > 0 {
			gain = priceDiff
		} else {
			loss = math.Abs(priceDiff)
		}
		avgGain = (avgGain*pv + gain) / pf
		avgLoss = (avgLoss*pv + loss) / pf
		if avgLoss == 0 {
			res[i] = 1
		} else {
			rs := avgGain / avgLoss
			rsi := 1 - (1 / (1 + rs))
			res[i] = rsi
		}
	}
	return &RSI{
		Len:       len(res),
		Res:       res,
		Period:    period,
		CandleArg: arg,
		avgGain:   avgGain,
		avgLoss:   avgLoss,
	}
}

type TSI struct {
	Len       int
	Res       []float64
	Period    int
	CandleArg cdl.CandleArg
	sumX,
	sumXSqr,
	sumXY,
	sumY,
	sumYSqr float64
}

func (t *TSI) Next(candles []cdl.Candle) {
	n := len(candles)
	if n < 2 || t.Len == 0 {
		return
	}
	pf := float64(t.Period)
	price := candles[n-1].Arg(t.CandleArg)
	oldPrice := candles[n-min(n, t.Period+1)].Arg(t.CandleArg)
	t.sumX += price - oldPrice
	t.sumXSqr += price*price - oldPrice*oldPrice
	t.sumXY += price*float64(t.Period-1) - (t.sumX - price)
	meanX := t.sumX / pf
	meanY := t.sumY / pf
	varianceX := t.sumXSqr/pf - meanX*meanX
	varianceY := t.sumYSqr/pf - meanY*meanY
	if varianceX < 0 {
		varianceX = 0
	}
	if varianceY < 0 {
		varianceY = 0
	}
	stdDevX := math.Sqrt(varianceX)
	stdDevY := math.Sqrt(varianceY)
	if stdDevX <= 1e-10 || stdDevY <= 1e-10 {
		t.Res = append(t.Res, 0)
	} else {
		covariance := t.sumXY/pf - meanX*meanY
		correlation := covariance / (stdDevX * stdDevY)
		if correlation > 1 {
			correlation = 1
		} else if correlation < -1 {
			correlation = -1
		}
		t.Res = append(t.Res, correlation)
	}
	t.Len++
}

func NewTSI(candles []cdl.Candle, arg cdl.CandleArg, period int) *TSI {
	n := len(candles)
	if n < 2 || period <= 0 {
		return nil
	}
	res := make([]float64, n)
	var sumX, sumXSqr, sumXY float64
	sumY := float64(period * (period - 1) / 2)
	sumYSqr := float64((period - 1) * period * (2*period - 1) / 6)
	for i := 0; i < period && i < n; i++ {
		price := candles[i].Arg(arg)
		sumX += price
		sumXSqr += price * price
		sumXY += price * float64(i)
	}
	pf := float64(period)
	for i := period; i < n; i++ {
		price := candles[i].Arg(arg)
		oldPrice := candles[i-period].Arg(arg)

		sumX += price - oldPrice
		sumXSqr += price*price - oldPrice*oldPrice
		sumXY += price*float64(period-1) - (sumX - price)

		meanX := sumX / pf
		meanY := sumY / pf

		varianceX := sumXSqr/pf - meanX*meanX
		varianceY := sumYSqr/pf - meanY*meanY
		if varianceX < 0 {
			varianceX = 0
		}
		if varianceY < 0 {
			varianceY = 0
		}
		stdDevX := math.Sqrt(varianceX)
		stdDevY := math.Sqrt(varianceY)

		if stdDevX <= 1e-10 || stdDevY <= 1e-10 {
			res[i] = 0
		} else {
			covariance := sumXY/pf - meanX*meanY
			correlation := covariance / (stdDevX * stdDevY)
			if correlation > 1 {
				correlation = 1
			} else if correlation < -1 {
				correlation = -1
			}
			res[i] = correlation
		}
	}
	return &TSI{
		Len:       len(res),
		Res:       res,
		Period:    period,
		CandleArg: arg,
		sumX:      sumX,
		sumXSqr:   sumXSqr,
		sumXY:     sumXY,
		sumY:      sumY,
		sumYSqr:   sumYSqr,
	}
}

type MFI struct {
	Len    int
	Res    []float64
	Period int
	sumPosFlow,
	sumNegFlow float64
}

func (m *MFI) Next(candles []cdl.Candle) {
	n := len(candles)
	if n < m.Period+2 || m.Len == 0 {
		return
	}
	oldIndex := n - m.Period + 1
	price := candles[n].Arg(cdl.HLC)
	oldPrice := candles[oldIndex].Arg(cdl.HLC)
	prevPrice := candles[n-1].Arg(cdl.HLC)
	prevOldPrice := candles[oldIndex-1].Arg(cdl.HLC)
	oldPriceChange := oldPrice - prevOldPrice
	oldMoneyFlow := candles[oldIndex].Volume * oldPrice
	if oldPriceChange > 0 {
		m.sumPosFlow -= oldMoneyFlow
	} else {
		m.sumNegFlow -= oldMoneyFlow
	}
	newPriceChange := price - prevPrice
	newMoneyFlow := candles[n].Volume * price
	if newPriceChange > 0 {
		m.sumPosFlow += newMoneyFlow
	} else {
		m.sumNegFlow += newMoneyFlow
	}
	if m.sumNegFlow == 0 {
		m.Res = append(m.Res, 1)
	} else {
		ratio := m.sumPosFlow / m.sumNegFlow
		m.Res = append(m.Res, 1*ratio/(1+ratio))
	}
	m.Len++
}

func NewMFI(candles []cdl.Candle, period int) *MFI {
	n := len(candles)
	if n == 0 || period <= 0 || n < period+1 {
		return nil
	}
	res := make([]float64, n)
	hlcPrices := make([]float64, n)
	for i := 0; i < n; i++ {
		hlcPrices[i] = candles[i].Arg(cdl.HLC)
	}
	var sumPosFlow, sumNegFlow float64
	for i := 1; i <= period; i++ {
		priceChange := hlcPrices[i] - hlcPrices[i-1]
		moneyFlow := candles[i].Volume * hlcPrices[i]
		if priceChange > 0 {
			sumPosFlow += moneyFlow
		} else {
			sumNegFlow += moneyFlow
		}
	}
	if sumNegFlow == 0 {
		res[period] = 1
	} else {
		moneyFlowRatio := sumPosFlow / sumNegFlow
		res[period] = 1 * moneyFlowRatio / (1.0 + moneyFlowRatio)
	}
	for i := period + 1; i < n; i++ {
		oldIndex := i - period
		oldPriceChange := hlcPrices[oldIndex] - hlcPrices[oldIndex-1]
		oldMoneyFlow := candles[oldIndex].Volume * hlcPrices[oldIndex]
		if oldPriceChange > 0 {
			sumPosFlow -= oldMoneyFlow
		} else {
			sumNegFlow -= oldMoneyFlow
		}
		newPriceChange := hlcPrices[i] - hlcPrices[i-1]
		newMoneyFlow := candles[i].Volume * hlcPrices[i]
		if newPriceChange > 0 {
			sumPosFlow += newMoneyFlow
		} else {
			sumNegFlow += newMoneyFlow
		}
		if sumNegFlow == 0 {
			res[i] = 1
		} else {
			moneyFlowRatio := sumPosFlow / sumNegFlow
			res[i] = 1 * moneyFlowRatio / (1 + moneyFlowRatio)
		}
	}
	return &MFI{
		Len:        len(res),
		Res:        res,
		Period:     period,
		sumPosFlow: sumPosFlow,
		sumNegFlow: sumNegFlow,
	}
}

type MACD struct {
	Len       int
	Hist      []float64
	MACD      []float64
	Signal    []float64
	CandleArg cdl.CandleArg
	fAlpha,
	sAlpha,
	dAlpha,
	fastMa,
	slowMa float64
}

func (m *MACD) Next(candles []cdl.Candle) {
	n := len(candles)
	if n == 0 || m.Len == 0 {
		return
	}
	price := candles[n-1].Arg(m.CandleArg)
	m.fastMa = price*m.fAlpha + m.fastMa*(1-m.fAlpha)
	m.slowMa = price*m.sAlpha + m.slowMa*(1-m.sAlpha)
	var macd float64
	if price != 0 {
		macd = (m.fastMa - m.slowMa) / price
	} else {
		macd = 0
	}
	m.MACD = append(m.MACD, macd)
	signal := macd*m.dAlpha + m.Signal[m.Len-1]*(1-m.dAlpha)
	m.Signal = append(m.Signal, signal)
	m.Hist = append(m.Hist, macd-signal)
	m.Len++
}

func NewMACD(candles []cdl.Candle, arg cdl.CandleArg, fPeriod int, sPeriod int, dPeriod int) *MACD {
	n := len(candles)
	if n <= 1 || fPeriod <= 0 || sPeriod <= 0 || dPeriod <= 0 {
		return nil
	}
	hist := make([]float64, n)
	macd := make([]float64, n)
	signal := make([]float64, n)
	fAlpha := 2 / (float64(fPeriod) + 1)
	sAlpha := 2 / (float64(sPeriod) + 1)
	dAlpha := 2 / (float64(dPeriod) + 1)
	price := candles[0].Arg(arg)
	fastMa := (price + candles[0].O) / 2
	slowMa := (price + candles[0].O) / 2
	signal[0] = (price + candles[0].O) / 2
	for i := 1; i < n; i++ {
		price := candles[i].Arg(arg)
		fastMa = price*fAlpha + fastMa*(1-fAlpha)
		slowMa = price*sAlpha + slowMa*(1-sAlpha)
		if price != 0 {
			macd[i] = (fastMa - slowMa) / price
		} else {
			macd[i] = 0
		}
		signal[i] = macd[i]*dAlpha + signal[i-1]*(1-dAlpha)
		hist[i] = macd[i] - signal[i]
	}
	return &MACD{
		Len:       len(hist),
		Hist:      hist,
		MACD:      macd,
		Signal:    signal,
		CandleArg: arg,
		fAlpha:    fAlpha,
		sAlpha:    sAlpha,
		dAlpha:    dAlpha,
		fastMa:    fastMa,
		slowMa:    slowMa,
	}
}

type TsiForV[V numeric.Number] struct {
	Len    int
	Res    []float64
	Period int
	sumX,
	sumXSqr,
	sumXY,
	sumY,
	sumYSqr float64
}

func NewTsiForV[V numeric.Number](s []V, period int) *TsiForV[V] {
	n := len(s)
	if n < 2 || period <= 0 {
		return nil
	}
	res := make([]float64, n)
	var sumX, sumXSqr, sumXY float64
	sumY := float64(period * (period - 1) / 2)
	sumYSqr := float64((period - 1) * period * (2*period - 1) / 6)
	for i := 0; i < period && i < n; i++ {
		price := float64(s[i])
		sumX += price
		sumXSqr += price * price
		sumXY += price * float64(i)
	}
	pf := float64(period)
	for i := period; i < n; i++ {
		price := float64(s[i])
		oldPrice := float64(s[i-period])

		sumX += price - oldPrice
		sumXSqr += price*price - oldPrice*oldPrice
		sumXY += price*float64(period-1) - oldPrice*float64(0) - (sumX - price)

		meanX := sumX / pf
		meanY := sumY / pf
		varianceX := sumXSqr/pf - meanX*meanX
		varianceY := sumYSqr/pf - meanY*meanY

		if varianceX < 0 {
			varianceX = 0
		}
		if varianceY < 0 {
			varianceY = 0
		}
		stdDevX := math.Sqrt(varianceX)
		stdDevY := math.Sqrt(varianceY)
		covariance := sumXY/pf - meanX*meanY
		if stdDevX <= 1e-10 || stdDevY <= 1e-10 {
			res[i] = 0
		} else {
			correlation := covariance / (stdDevX * stdDevY)
			if correlation > 1 {
				correlation = 1
			} else if correlation < -1 {
				correlation = -1
			}
			res[i] = correlation
		}
	}
	return &TsiForV[V]{
		Len:     len(res),
		Res:     res,
		Period:  period,
		sumX:    sumX,
		sumXSqr: sumXSqr,
		sumXY:   sumXY,
		sumY:    sumY,
		sumYSqr: sumYSqr,
	}
}

func (t *TsiForV[V]) Next(s []V) {
	n := len(s)
	if n < 2 {
		return
	}
	pf := float64(t.Period)
	price := float64(s[n-1])
	oldPrice := float64(s[n-min(n, t.Period+1)])
	t.sumX += price - oldPrice
	t.sumXSqr += price*price - oldPrice*oldPrice
	t.sumXY += price*float64(t.Period-1) - oldPrice*float64(0) - (t.sumX - price)
	meanX := t.sumX / pf
	meanY := t.sumY / pf
	varianceX := t.sumXSqr/pf - meanX*meanX
	varianceY := t.sumYSqr/pf - meanY*meanY

	if varianceX < 0 {
		varianceX = 0
	}
	if varianceY < 0 {
		varianceY = 0
	}
	stdDevX := math.Sqrt(varianceX)
	stdDevY := math.Sqrt(varianceY)
	covariance := t.sumXY/pf - meanX*meanY
	if stdDevX <= 1e-10 || stdDevY <= 1e-10 {
		t.Res = append(t.Res, 0)
	} else {
		correlation := covariance / (stdDevX * stdDevY)
		if correlation > 1 {
			correlation = 1
		} else if correlation < -1 {
			correlation = -1
		}
		t.Res = append(t.Res, correlation)
	}
	t.Len++
}
