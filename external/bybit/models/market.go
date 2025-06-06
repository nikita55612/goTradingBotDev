package models

// CandleRawData представляет данные свечи за определенный период времени.
type CandleRawData struct {
	Category string      `json:"category"` // тип продукта (например, "inverse" - обратный контракт)
	Symbol   string      `json:"symbol"`   // название символа (например, "BTCUSD")
	List     [][7]string `json:"list"`     // массив данных свечей, отсортированный в обратном порядке по времени
}

// InstrumentsInfo представляет ответ API с информацией об инструментах
type InstrumentsInfo struct {
	Category string           `json:"category"` // Категория инструментов (spot, linear, inverse)
	List     []InstrumentInfo `json:"list"`     // Список инструментов
}

// InstrumentInfo представляет информацию о торговой паре (спот, фьючерсы, перпетуальные контракты)
type InstrumentInfo struct {
	Symbol             string `json:"symbol"`             // Название торговой пары (например BTCUSDT)
	ContractType       string `json:"contractType"`       // Тип контракта (LinearPerpetual, InversePerpetual и т.д.)
	Status             string `json:"status"`             // Статус инструмента (Trading, PreLaunch и т.д.)
	BaseCoin           string `json:"baseCoin"`           // Базовая монета (например BTC)
	QuoteCoin          string `json:"quoteCoin"`          // Котируемая монета (например USDT)
	LaunchTime         string `json:"launchTime"`         // Время запуска в timestamp (мс)
	DeliveryTime       string `json:"deliveryTime"`       // Время доставки для фьючерсов (0 для перпетуальных)
	DeliveryFeeRate    string `json:"deliveryFeeRate"`    // Ставка комиссии за доставку
	PriceScale         string `json:"priceScale"`         // Масштаб цены (количество знаков после запятой)
	Innovation         string `json:"innovation"`         // Является ли инструментом зоны инноваций (0: нет, 1: да)
	MarginTrading      string `json:"marginTrading"`      // Доступна ли маржинальная торговля
	StTag              string `json:"stTag"`              // Специальный тег (0: нет, 1: да)
	SettleCoin         string `json:"settleCoin"`         // Монета расчетов
	CopyTrading        string `json:"copyTrading"`        // Доступность копи-трейдинга (none, both и т.д.)
	UpperFundingRate   string `json:"upperFundingRate"`   // Верхний предел ставки фандинга
	LowerFundingRate   string `json:"lowerFundingRate"`   // Нижний предел ставки фандинга
	DisplayName        string `json:"displayName"`        // Отображаемое имя в UI
	FundingInterval    int    `json:"fundingInterval"`    // Интервал фандинга (в минутах)
	IsPreListing       bool   `json:"isPreListing"`       // Является ли премаркет-контрактом
	UnifiedMarginTrade bool   `json:"unifiedMarginTrade"` // Поддержка унифицированной маржи

	LeverageFilter struct {
		MinLeverage  string `json:"minLeverage"`  // Минимальное плечо
		MaxLeverage  string `json:"maxLeverage"`  // Максимальное плечо
		LeverageStep string `json:"leverageStep"` // Шаг изменения плеча
	} `json:"leverageFilter"`

	LotSizeFilter struct {
		BasePrecision       string `json:"basePrecision"`       // Точность базовой монеты
		QuotePrecision      string `json:"quotePrecision"`      // Точность котируемой монеты
		MinOrderQty         string `json:"minOrderQty"`         // Минимальное количество для ордера
		MaxOrderQty         string `json:"maxOrderQty"`         // Максимальное количество для Limit и PostOnly ордера
		MaxMktOrderQty      string `json:"maxMktOrderQty"`      // Максимальное количество для Market ордера
		MinOrderAmt         string `json:"minOrderAmt"`         // Минимальная сумма ордера
		MaxOrderAmt         string `json:"maxOrderAmt"`         // Максимальная сумма ордера
		QtyStep             string `json:"qtyStep"`             // Шаг изменения количества
		MinNotionalValue    string `json:"minNotionalValue"`    // Минимальная номинальная стоимость
		PostOnlyMaxOrderQty string `json:"postOnlyMaxOrderQty"` // Устарело, использовать maxOrderQty
	} `json:"lotSizeFilter"`

	PriceFilter struct {
		MinPrice string `json:"minPrice"` // Минимальная цена ордера
		MaxPrice string `json:"maxPrice"` // Максимальная цена ордера
		TickSize string `json:"tickSize"` // Шаг изменения цены (tick size)
	} `json:"priceFilter"`

	RiskParameters struct {
		PriceLimitRatioX string `json:"priceLimitRatioX"` // Коэффициент лимита цены X
		PriceLimitRatioY string `json:"priceLimitRatioY"` // Коэффициент лимита цены Y
	} `json:"riskParameters"`

	PreListingInfo *struct {
		CurAuctionPhase string `json:"curAuctionPhase"` // Текущая фаза аукциона
		Phases          []struct {
			Phase     string `json:"phase"`     // Фаза премаркет-трейдинга
			StartTime string `json:"startTime"` // Время начала фазы (timestamp мс)
			EndTime   string `json:"endTime"`   // Время окончания фазы (timestamp мс)
		} `json:"phases"` // Информация о фазах
		AuctionFeeInfo struct {
			AuctionFeeRate string `json:"auctionFeeRate"` // Ставка комиссии во время аукциона
			TakerFeeRate   string `json:"takerFeeRate"`   // Ставка тейкера в фазе непрерывного трейдинга
			MakerFeeRate   string `json:"makerFeeRate"`   // Ставка мейкера в фазе непрерывного трейдинга
		} `json:"auctionFeeInfo"` // Информация о комиссиях
	} `json:"preListingInfo"` // Информация о премаркете (если isPreListing=true)
}

// Tickers представляет информацию о текущих ценах и рыночных данных инструментов
type Tickers struct {
	Category string   `json:"category"` // Категория инструментов (spot, inverse, linear)
	List     []Ticker `json:"list"`     // Список тикеров
}

// Ticker содержит детальную информацию о рыночных данных инструмента
type Ticker struct {
	Symbol                 string `json:"symbol"`                 // Название торговой пары (например BTCUSD)
	LastPrice              string `json:"lastPrice"`              // Последняя цена сделки
	IndexPrice             string `json:"indexPrice"`             // Индексная цена
	MarkPrice              string `json:"markPrice"`              // Маркировочная цена
	PrevPrice24h           string `json:"prevPrice24h"`           // Цена 24 часа назад
	Price24hPcnt           string `json:"price24hPcnt"`           // Процентное изменение цены за 24 часа
	HighPrice24h           string `json:"highPrice24h"`           // Максимальная цена за 24 часа
	LowPrice24h            string `json:"lowPrice24h"`            // Минимальная цена за 24 часа
	PrevPrice1h            string `json:"prevPrice1h"`            // Цена 1 час назад
	OpenInterest           string `json:"openInterest"`           // Открытый интерес (количество открытых контрактов)
	OpenInterestValue      string `json:"openInterestValue"`      // Стоимость открытого интереса
	Turnover24h            string `json:"turnover24h"`            // Оборот за 24 часа
	Volume24h              string `json:"volume24h"`              // Объем торгов за 24 часа
	FundingRate            string `json:"fundingRate"`            // Текущая ставка финансирования
	NextFundingTime        string `json:"nextFundingTime"`        // Время следующего финансирования (мс)
	PredictedDeliveryPrice string `json:"predictedDeliveryPrice"` // Прогнозируемая цена доставки (за 30 мин до поставки)
	BasisRate              string `json:"basisRate"`              // Базисная ставка
	Basis                  string `json:"basis"`                  // Базис
	DeliveryFeeRate        string `json:"deliveryFeeRate"`        // Ставка комиссии за доставку
	DeliveryTime           string `json:"deliveryTime"`           // Время доставки (мс, для фьючерсов с истечением)
	Ask1Size               string `json:"ask1Size"`               // Объем лучшей ask-цены
	Bid1Price              string `json:"bid1Price"`              // Лучшая bid-цена
	Ask1Price              string `json:"ask1Price"`              // Лучшая ask-цена
	Bid1Size               string `json:"bid1Size"`               // Объем лучшей bid-цены
	PreOpenPrice           string `json:"preOpenPrice"`           // Предполагаемая цена открытия премаркета (теряет смысл после открытия рынка)
	PreQty                 string `json:"preQty"`                 // Предполагаемый объем открытия премаркета (теряет смысл после открытия рынка)
	CurPreListingPhase     string `json:"curPreListingPhase"`     // Текущая фаза премаркет-контракта
}

// CandleStreamRawData представляет потоковые данные свечи
type CandleStreamRawData struct {
	Ts    int64  `json:"ts"`    // Временная метка
	Type  string `json:"type"`  // Тип сообщения
	Topic string `json:"topic"` // Топик подписки

	Data []struct {
		Start     int64  `json:"start"`     // Начальное время
		End       int64  `json:"end"`       // Конечное время
		Timestamp int64  `json:"timestamp"` // Временная метка
		Interval  string `json:"interval"`  // Интервал
		Open      string `json:"open"`      // Цена открытия
		Close     string `json:"close"`     // Цена закрытия
		High      string `json:"high"`      // Максимальная цена
		Low       string `json:"low"`       // Минимальная цена
		Volume    string `json:"volume"`    // Объем
		Turnover  string `json:"turnover"`  // Оборот
		Confirm   bool   `json:"confirm"`   // Подтверждение
	} `json:"data"`
}
