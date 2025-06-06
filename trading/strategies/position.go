package strategies

import (
	"goTradingBot/trading/types"
	"goTradingBot/utils/seqs"
	"sync"
)

type Position struct {
	OrderLog    *seqs.OrderedMap[string, *types.Order]
	avalBalance float64
	balance     float64
	mu          sync.Mutex
}

func (p *Position) XXX(order *types.Order) {
	p.mu.Lock()
	defer p.mu.Unlock()

}
