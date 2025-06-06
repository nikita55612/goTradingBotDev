package cdl

import (
	"goTradingBot/utils/numeric"
	"math"
)

func ListOfCandleArg(candles []Candle, arg CandleArg) []float64 {
	n := len(candles)
	list := make([]float64, n)
	for i := 0; i < n; i++ {
		list[i] = candles[i].Arg(arg)
	}
	return list
}

type CandleArg string

const (
	// Time - время открытия свечи
	Time CandleArg = "T"

	// Open - цена открытия свечи
	Open CandleArg = "O"

	// High - максимальная цена свечи
	High CandleArg = "H"

	// Low - минимальная цена свечи
	Low CandleArg = "L"

	// Close - цена закрытия свечи
	Close CandleArg = "C"

	// CloseLowAvg - среднее между ценой закрытия и минимумом (close + low)/2
	CL CandleArg = "CL"

	// CloseHighAvg - среднее между ценой закрытия и максимумом (close + high)/2
	CH CandleArg = "CH"

	// HighLowAvg - среднее между максимумом и минимумом (high + low)/2
	HL CandleArg = "HL"

	// HLC3 - типичная цена (high + low + close)/3
	HLC CandleArg = "HLC"

	// OHLC4 - среднее всех цен свечи (open + high + low + close)/4
	OHLC CandleArg = "OHLC"

	// HLCC4 - взвешенное среднее с удвоенным закрытием (high + low + close*2)/4
	HLCC CandleArg = "HLCC"

	// Volume - объем торгов
	Volume CandleArg = "V"

	// Turnover - оборот торгов
	Turnover CandleArg = "Turnover"

	// TrueRange - истинный диапазон (high - low)
	TrueRange CandleArg = "TR"

	// NormalizedRange - нормализованный диапазон (high - low)/open
	NormalizedRange CandleArg = "NR"

	// RateOfChange - процент изменения цены (close - open)/open
	RateOfChange CandleArg = "ROC"

	// Momentum - импульс (close - open)
	Momentum CandleArg = "M"

	// Acceleration - ускорение (close - open)/(high - low)
	Acceleration CandleArg = "Acc"

	// PriceVolume - произведение цены закрытия на объем (close * volume)
	PriceVolume CandleArg = "PV"

	// Body - абсолютное значение тела свечи (|close - open|)
	Body CandleArg = "AM"

	// UpperWick - длина верхней тени (high - max(open, close))
	UpperWick CandleArg = "UW"

	// LowerWick - длина нижней тени (min(open, close) - low)
	LowerWick CandleArg = "LW"

	// UpperWick / LowerWick
	WickRatio CandleArg = "WR"

	// BodyRangeRatio - отношение тела к общему диапазону (|close - open| / (high - low))
	BodyRangeRatio CandleArg = "AMRR"

	// Direction - направление свечи: 1 (бычья), -1 (медвежья), 0 (доджи)
	Direction CandleArg = "Dir"

	// WeightedClose - взвешенная цена закрытия (open + high + low + 2*close)/5
	WeightedClose CandleArg = "OHLCC"

	// VWAP - средневзвешенная цена по объёму (turnover / volume)
	VWAP CandleArg = "VWAP"

	// CloseLocationValue - положение закрытия относительно диапазона (close - low)/(high - low)
	CloseLocationValue CandleArg = "CLV"

	// ShadowRatio - отношение теней к телу (upperWick + lowerWick)/body
	ShadowRatio CandleArg = "SR"
)

func (c *Candle) Arg(a CandleArg) float64 {
	switch a {
	case Time:
		return float64(c.Time)
	case Open:
		return c.O
	case High:
		return c.H
	case Low:
		return c.L
	case Close:
		return c.C
	case CL:
		return numeric.SafeAvg2Val(c.C, c.L)
	case CH:
		return numeric.SafeAvg2Val(c.C, c.H)
	case HL:
		return numeric.SafeAvg2Val(c.H, c.L)
	case HLC:
		return (c.H + c.L + c.C) / 3
	case OHLC:
		return (c.O + c.H + c.L + c.C) / 4
	case HLCC:
		return (c.H + c.L + 2*c.C) / 4
	case TrueRange:
		return c.H - c.L
	case Momentum:
		return c.C - c.O
	case Acceleration:
		if c.H != c.L {
			return (c.C - c.O) / (c.H - c.L)
		}
	case NormalizedRange:
		if c.O != 0 {
			return (c.H - c.L) / c.O
		}
	case RateOfChange:
		if c.O != 0 {
			return (c.C - c.O) / c.O
		}
	case Volume:
		return c.Volume
	case PriceVolume:
		return c.Volume * c.C
	case Turnover:
		return c.Turnover
	case Body:
		return math.Abs(c.C - c.O)
	case UpperWick:
		return c.H - max(c.O, c.C)
	case LowerWick:
		return min(c.O, c.C) - c.L
	case BodyRangeRatio:
		if c.H != c.L {
			return math.Abs(c.C-c.O) / (c.H - c.L)
		}
	case Direction:
		if c.C > c.O {
			return 1
		} else if c.C < c.O {
			return -1
		}
		return 0
	case WeightedClose:
		return (c.O + c.H + c.L + 2*c.C) / 5
	case VWAP:
		if c.Volume != 0 {
			return c.Turnover / c.Volume
		}
	case CloseLocationValue:
		tr := (c.H - c.L)
		if tr != 0 {
			return (c.C - c.L) / tr
		}
		return 0
	case ShadowRatio:
		body := math.Abs(c.C - c.O)
		if body != 0 {
			return (c.H - c.L - body) / body
		}
		return 0
	}
	return 0
}
