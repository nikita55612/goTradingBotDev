package strategies

import (
	"goTradingBot/cdl"
	"goTradingBot/predict"
	"goTradingBot/predict/portal"
	"goTradingBot/trading/types"
	"goTradingBot/utils/numeric"
	"goTradingBot/utils/seqs"
	"math"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Strategy struct {
	StrategyABC
	symbol            string
	interval          cdl.Interval
	model             string
	balance           float64
	longRatio         float64
	orderLog          *seqs.OrderedMap[string, *types.Order]
	qtyPrecision      int
	minOrderAmt       float64
	tickSize          float64
	tickSizePrecision int
	closeOrderTimeout time.Duration
	confirmCandleChan chan *cdl.CandleStreamData
	lastPriceChan     chan float64
	lastPrice         atomic.Pointer[float64]
	limitOrderOffset  float64
	limitCeilPrice    atomic.Pointer[float64]
	limitFloorPrice   atomic.Pointer[float64]
}

func NewStrategy(
	symbol string,
	interval cdl.Interval,
	model string,
	balance float64,
	longRatio float64,
	limitOrderOffset float64,
) *Strategy {
	closeOrderTimeout := time.Duration(interval.AsSeconds())*time.Second - 5

	return &Strategy{
		symbol:            symbol,
		interval:          interval,
		model:             model,
		balance:           balance,
		longRatio:         longRatio,
		orderLog:          seqs.NewOrderedMap[string, *types.Order](32),
		closeOrderTimeout: closeOrderTimeout,
		confirmCandleChan: make(chan *cdl.CandleStreamData),
		lastPriceChan:     make(chan float64, 8),
		limitOrderOffset:  limitOrderOffset,
	}
}

func (s *Strategy) Go() error {
	info, err := s.subData.GetInstrumentInfo(s.symbol)
	if err != nil {
		return err
	}

	s.qtyPrecision = info.QtyPrecision
	s.minOrderAmt = info.MinOrderAmt
	s.tickSize = info.TickSize
	s.tickSizePrecision = numeric.DecimalPlaces(s.tickSize)

	go s.background()
	go s.confirmCandleHandler()
	go s.observeCandleStreamData()
	go func() {
		<-s.ctx.Done()
		close(s.confirmCandleChan)
		close(s.lastPriceChan)
		s.close()
	}()

	return nil
}

func (s *Strategy) close() {
	qty := -s.qtyPosition()
	order := types.NewOrder(s.symbol, qty, nil)
	linkId := uuid.NewString()
	s.orderRequest <- &types.OrderRequest{
		LinkId:       linkId,
		Tag:          "test",
		Order:        order,
		CloseTimeout: s.closeOrderTimeout,
		Reply:        nil,
	}
}

func (s *Strategy) background() {
	ticker := time.NewTicker(8 * time.Second)
	defer ticker.Stop()

	lastPrice := <-s.lastPriceChan
	s.lastPrice.Store(&lastPrice)
	s.limitCeilPrice.Store(&lastPrice)
	s.limitCeilPrice.Store(&lastPrice)
	for lastPrice := range s.lastPriceChan {
		s.lastPrice.Store(&lastPrice)
		select {
		case <-ticker.C:
			limitCeilPrice := numeric.TruncateFloat(
				lastPrice*(1+s.limitOrderOffset), s.tickSizePrecision,
			)
			s.limitCeilPrice.Store(&limitCeilPrice)
			limitFloorPrice := numeric.TruncateFloat(
				lastPrice*(1-s.limitOrderOffset), s.tickSizePrecision,
			)
			s.limitFloorPrice.Store(&limitFloorPrice)
		default:
		}
	}
}

func (s *Strategy) observeCandleStreamData() {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			ch := make(chan *cdl.CandleStreamData)
			done, err := s.subData.SubscribeChan(s.symbol, s.interval, ch)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}

			for data := range ch {
				s.lastPriceChan <- data.Candle.C
				if data.Confirm {
					s.confirmCandleChan <- data
				}
			}
			close(done)
		}
	}
}

func (s *Strategy) qtyPosition() float64 {
	var qtyPosition float64
	s.orderLog.Range(func(_ string, o *types.Order) bool {
		o.Lock()
		defer o.Unlock()

		if o.ID != "" {
			qtyPosition += o.ExecQty
		}
		return true
	})
	return qtyPosition
}

func (s *Strategy) confirmCandleHandler() {
	for data := range s.confirmCandleChan {
		signal, _ := s.getSignal(data)
		if signal == types.Hold {
			continue
		}

		calcQty := func() float64 {
			qty := s.balance / *s.lastPrice.Load()
			if signal == types.Buy {
				return qty * s.longRatio
			} else if signal == types.Sell {
				return -qty * (1 - s.longRatio)
			}
			return 0
		}

		var qty float64
		if s.orderLog.Len() == 0 {
			qty = calcQty()
		} else {
			qty = -s.qtyPosition() + calcQty()
		}

		qty = numeric.RoundFloat(qty, s.qtyPrecision)
		if math.Abs(qty**s.lastPrice.Load()) < s.minOrderAmt {
			continue
		}

		var price float64
		if qty > 0 {
			price = *s.limitCeilPrice.Load()
		} else {
			price = *s.limitFloorPrice.Load()
		}
		order := types.NewOrder(s.symbol, qty, &price)
		linkId := uuid.NewString()
		s.orderLog.Set(linkId, order)
		select {
		case <-s.ctx.Done():
		case s.orderRequest <- &types.OrderRequest{
			LinkId:       linkId,
			Tag:          "test",
			Order:        order,
			CloseTimeout: s.closeOrderTimeout,
			Reply:        nil,
		}:
		}
	}
}

func (s *Strategy) getSignal(data *cdl.CandleStreamData) (types.Signal, error) {
	limit := predict.GetModelWinSize(predict.A6N21P9) + predict.FeatureOffset
	candles, err := s.subData.GetCandles(s.symbol, data.Interval, limit)
	if err != nil {
		return types.Hold, err
	}

	intervalDuration := time.Duration(data.Interval.AsSeconds())
	candlesSince := time.Since(time.UnixMilli(candles[len(candles)-1].Time))
	if candlesSince > intervalDuration {
		candles = append(candles, data.Candle)
	}

	features := predict.FeaturesGeneratorModel(predict.A6N21P9).
		GenTranspose(candles, predict.FeatureOffset, -1)
	features = features[len(features)-2:]

	prediction, err := portal.GetPrediction(
		features,
		s.model,
	).UnwrapSinglePredict()

	if err != nil {
		return types.Hold, err
	}

	n := len(prediction)
	if prediction[n-1] > 0.5 && prediction[n-2] < 0.5 {
		return types.Buy, nil
	}
	if prediction[n-1] < 0.5 && prediction[n-2] > 0.5 {
		return types.Sell, nil
	}

	return types.Hold, nil
}
