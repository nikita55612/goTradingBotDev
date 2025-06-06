package cdl

import "math"

// CandleRatio представляет тип соотношения между свечами
type CandleRatio string

const (
	BodyStrengthRatio  CandleRatio = "AMR" // Отношение размеров тел свечей (|Close-Open|)
	LowerWickRatio     CandleRatio = "LWR" // Отношение нижних теней текущей/предыдущей свечи
	UpperWickRatio     CandleRatio = "UWR" // Отношение верхних теней текущей/предыдущей свечи
	ClosePositionRatio CandleRatio = "CPR" // Положение закрытия текущей свечи относительно диапазона предыдущей
	MomentumRatio      CandleRatio = "MR"  // Отношение моментумов (Close-Open) текущей/предыдущей свечи
	BreakoutPower      CandleRatio = "BP"  // Сила пробоя относительно диапазона предыдущей свечи
	VolumeRatio        CandleRatio = "VR"  // Отношение объемов текущей/предыдущей свечи
	TrueRangeRatio     CandleRatio = "TRR" // Истинный диапазон (макс из High-Low, |High-PrevClose|, |Low-PrevClose|)
)

// ListOfCandleRatio вычисляет соотношения для списка свечей с заданным сдвигом
func ListOfCandleRatio(candles []Candle, r CandleRatio, shift int) []float64 {
	if shift == 0 {
		panic("ListOfCandleRatio: shift не может быть равен 0")
	}
	n := len(candles)
	ratios := make([]float64, n)
	for i := shift; i < n; i++ {
		ratios[i] = candles[i].Ratio(r, &candles[i-shift])
	}
	return ratios
}

// Ratio вычисляет соотношение между двумя свечами
func (c *Candle) Ratio(r CandleRatio, pc *Candle) float64 {
	if pc == nil {
		return 0
	}
	switch r {
	case BodyStrengthRatio:
		currentBody := c.Arg(Body)
		prevBody := pc.Arg(Body)
		if prevBody == 0 {
			return 0
		}
		return currentBody / prevBody
	case LowerWickRatio:
		currentLW := c.Arg(LowerWick)
		prevLW := pc.Arg(LowerWick)
		if prevLW == 0 {
			return 0
		}
		return currentLW / prevLW
	case UpperWickRatio:
		currentUW := c.Arg(UpperWick)
		prevUW := pc.Arg(UpperWick)
		if prevUW == 0 {
			return 0
		}
		return currentUW / prevUW
	case ClosePositionRatio:
		prevRange := pc.Arg(TrueRange)
		if prevRange == 0 {
			return 0
		}
		return (c.C - pc.L) / prevRange
	case MomentumRatio:
		prevMoment := pc.Arg(Momentum)
		if prevMoment == 0 {
			return 0
		}
		return c.Arg(Momentum) / prevMoment
	case BreakoutPower:
		prevTr := pc.Arg(TrueRange)
		if prevTr == 0 {
			return 0
		}
		if c.C > pc.H {
			return (c.C - pc.H) / prevTr
		}
		if c.C <= pc.L {
			return (pc.L - c.C) / prevTr
		}
	case VolumeRatio:
		if pc.Volume == 0 {
			return 1
		}
		return c.Volume / pc.Volume
	case TrueRangeRatio:
		return max(
			c.Arg(TrueRange),
			math.Abs(c.H-pc.C),
			math.Abs(c.L-pc.C),
		)
	}
	return 0
}
