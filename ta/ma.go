package ta

import (
	"goTradingBot/cdl"
	"goTradingBot/utils/numeric"
)

type MaType string

const (
	S  MaType = "SMA"
	E  MaType = "EMA"
	VW MaType = "VWMA"
)

type MovingAverage interface {
	Next(candles []cdl.Candle)
	Last() float64
	Crop()
	MaRes() []float64
}

func NewMovingAverage(maT MaType, candles []cdl.Candle, arg cdl.CandleArg, period int) MovingAverage {
	switch maT {
	case S:
		return NewSMA(candles, arg, period)
	case E:
		return NewEMA(candles, arg, period, 2)
	case VW:
		return NewVWMA(candles, arg, period)
	default:
		return NewVWMA(candles, arg, period)
	}
}

func CMA[V numeric.Number](s []V) []float64 {
	n := len(s)
	if n == 0 {
		return []float64{}
	}
	res := make([]float64, n)
	for i := 0; i < n; i++ {
		res[i] = numeric.Avg(s[:i+1])
	}
	return res
}

type SMA[V numeric.Number] struct {
	Res       []float64
	Len       int
	Period    int
	sum       float64
	CandleArg cdl.CandleArg
}

func (s *SMA[V]) MaRes() []float64 {
	return s.Res
}

func (s *SMA[V]) Crop() {
	if len(s.Res) == 0 {
		return
	}
	s.Res = s.Res[max(0, s.Len-s.Period):]
	newRes := make([]float64, len(s.Res))
	copy(newRes, s.Res)
	s.Res = newRes
	s.Len = len(s.Res)
}

func (s *SMA[V]) Last() float64 {
	return s.Res[s.Len-1]
}

func (a *SMA[V]) NextForV(s []V) {
	n := len(s)
	a.sum += float64(s[n-1]) - float64(s[n-min(n, a.Period+1)])
	a.Res = append(a.Res, a.sum/float64(a.Period))
	a.Len++
}

func (a *SMA[V]) Next(candles []cdl.Candle) {
	n := len(candles)
	a.sum = a.sum - candles[n-min(n, a.Period+1)].Arg(a.CandleArg) +
		candles[n-1].Arg(a.CandleArg)
	a.Res = append(a.Res, a.sum/float64(a.Period))
	a.Len++
}

func NewSmaForV[V numeric.Number](s []V, period int) *SMA[V] {
	n := len(s)
	if n == 0 || period <= 0 {
		return nil
	}
	res := make([]float64, n)
	var sum float64
	for i := 0; i < period && i < n; i++ {
		sum += float64(s[i])
		res[i] = sum / float64(i+1)
	}
	for i := period; i < n; i++ {
		sum = sum - float64(s[i-period]) + float64(s[i])
		res[i] = sum / float64(period)
	}
	return &SMA[V]{
		Res:    res,
		Len:    len(res),
		Period: period,
		sum:    sum,
	}
}

func NewSMA[V float64](candles []cdl.Candle, arg cdl.CandleArg, period int) *SMA[V] {
	n := len(candles)
	if n == 0 || period <= 0 {
		return nil
	}
	res := make([]float64, n)
	var sum float64
	for i := 0; i < period && i < n; i++ {
		sum += candles[i].Arg(arg)
		res[i] = sum / float64(i+1)
	}
	for i := period; i < n; i++ {
		sum += candles[i].Arg(arg) - candles[i-period].Arg(arg)
		res[i] = sum / float64(period)
	}
	return &SMA[V]{
		Res:       res,
		Len:       len(res),
		Period:    period,
		sum:       sum,
		CandleArg: arg,
	}
}

type EMA[V numeric.Number] struct {
	Res       []float64
	Len       int
	Period    int
	W         float64
	alpha     float64
	CandleArg cdl.CandleArg
}

func (s *EMA[V]) MaRes() []float64 {
	return s.Res
}

func (s *EMA[V]) Last() float64 {
	return s.Res[s.Len-1]
}

func (s *EMA[V]) Crop() {
	if len(s.Res) == 0 {
		return
	}
	s.Res = s.Res[max(0, s.Len-s.Period):]
	newRes := make([]float64, len(s.Res))
	copy(newRes, s.Res)
	s.Res = newRes
	s.Len = len(s.Res)
}

func (e *EMA[V]) NextForV(v V) {
	e.Res = append(e.Res, float64(v)*e.alpha+e.Res[e.Len-1]*(1-e.alpha))
	e.Len++
}

func (e *EMA[V]) Next(candles []cdl.Candle) {
	n := len(candles)
	e.Res = append(e.Res, candles[n-1].Arg(e.CandleArg)*e.alpha+e.Res[e.Len-1]*(1-e.alpha))
	e.Len++
}

func NewEmaForV[V numeric.Number](s []V, period int, w float64) *EMA[V] {
	n := len(s)
	if n == 0 || period <= 0 {
		return nil
	}
	res := make([]float64, n)
	res[0] = float64(s[0])
	alpha := w / (float64(period) + w - 1)
	for i := 1; i < n; i++ {
		res[i] = float64(s[i])*alpha + res[i-1]*(1-alpha)
	}
	return &EMA[V]{
		Res:    res,
		Len:    len(res),
		Period: period,
		W:      w,
		alpha:  alpha,
	}
}

func NewEMA[V float64](candles []cdl.Candle, arg cdl.CandleArg, period int, w float64) *EMA[V] {
	n := len(candles)
	if n == 0 || period <= 0 {
		return nil
	}
	res := make([]float64, n)
	res[0] = candles[0].Arg(arg)
	alpha := w / (float64(period) + w - 1)
	for i := 1; i < n; i++ {
		res[i] = candles[i].Arg(arg)*alpha + res[i-1]*(1-alpha)
	}
	return &EMA[V]{
		Res:       res,
		Len:       len(res),
		Period:    period,
		W:         w,
		alpha:     alpha,
		CandleArg: arg,
	}
}

type VWMA struct {
	Res         []float64
	Len         int
	Period      int
	CandleArg   cdl.CandleArg
	sumPriceVol float64
	sumVolume   float64
}

func (s *VWMA) MaRes() []float64 {
	return s.Res
}

func (s *VWMA) Last() float64 {
	return s.Res[s.Len-1]
}

func (s *VWMA) Crop() {
	if len(s.Res) == 0 {
		return
	}
	s.Res = s.Res[max(0, s.Len-s.Period):]
	newRes := make([]float64, len(s.Res))
	copy(newRes, s.Res)
	s.Res = newRes
	s.Len = len(s.Res)
}

func (v *VWMA) Next(candles []cdl.Candle) {
	n := len(candles)
	price := candles[n-1].Arg(v.CandleArg)
	volume := candles[n-1].Volume
	oldPrice := candles[n-min(n, v.Period+1)].Arg(v.CandleArg)
	oldVolume := candles[n-min(n, v.Period+1)].Volume

	v.sumPriceVol += (price * volume) - (oldPrice * oldVolume)
	v.sumVolume += volume - oldVolume

	v.Res = append(v.Res, v.sumPriceVol/v.sumVolume)
	v.Len++
}

func NewVWMA(candles []cdl.Candle, arg cdl.CandleArg, period int) *VWMA {
	n := len(candles)
	if n == 0 || period <= 0 {
		return nil
	}
	res := make([]float64, n)
	var sumPriceVol float64
	var sumVolume float64
	for i := 0; i < period && i < n; i++ {
		price := candles[i].Arg(arg)
		volume := candles[i].Volume
		sumPriceVol += price * volume
		sumVolume += volume
		res[i] = sumPriceVol / sumVolume
	}
	for i := period; i < n; i++ {
		price := candles[i].Arg(arg)
		volume := candles[i].Volume
		oldPrice := candles[i-period].Arg(arg)
		oldVolume := candles[i-period].Volume

		sumPriceVol += (price * volume) - (oldPrice * oldVolume)
		sumVolume += volume - oldVolume

		res[i] = sumPriceVol / sumVolume
	}
	return &VWMA{
		Res:         res,
		Len:         len(res),
		Period:      period,
		CandleArg:   arg,
		sumPriceVol: sumPriceVol,
		sumVolume:   sumVolume,
	}
}
