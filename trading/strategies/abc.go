package strategies

import (
	"context"
	"goTradingBot/trading/types"
)

type StrategyABC struct {
	ctx          context.Context
	orderRequest chan<- *types.OrderRequest
	subData      *types.SubData
}

func (s *StrategyABC) Init(ctx context.Context, subData *types.SubData, req chan<- *types.OrderRequest) {
	s.ctx = ctx
	s.subData = subData
	s.orderRequest = req
}
