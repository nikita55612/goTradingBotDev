package config

type TradingBotConfig struct {
	ChannelBufferSize  int `json:"channelBufferSize"`  // размер буфера канала для приёма ордеров
	SubDataBufferSize  int `json:"subDataBufferSize"`  // размер буфера исторических данных
	PlaceOrderInterval int `json:"placeOrderInterval"` // интервал между попытками размещения (мс)
	PlaceOrderTimeout  int `json:"placeOrderTimeout"`  // таймаут размещения ордера (мс)
	CheckOrderInterval int `json:"checkOrderInterval"` // интервал проверки статуса (мс)
	LongCheckInterval  int `json:"longCheckInterval"`  // увеличенный интервал проверки (мс)
	OrderStatusTimeout int `json:"orderStatusTimeout"` // таймаут ожидания закрытия (мс)
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultTradingBotConfig() *TradingBotConfig {
	return &TradingBotConfig{
		ChannelBufferSize:  64,
		SubDataBufferSize:  2000,
		PlaceOrderInterval: 200,
		PlaceOrderTimeout:  2000,
		CheckOrderInterval: 500,
		LongCheckInterval:  5000,
		OrderStatusTimeout: 3600000,
	}
}

type Strategy struct {
	Tag string
}
