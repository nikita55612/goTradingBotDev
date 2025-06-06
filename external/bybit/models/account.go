package models

// AccountInfo содержит информацию об аккаунте из API Bybit
type AccountInfo struct {
	UnifiedMarginStatus int    `json:"unifiedMarginStatus"` // Статус унифицированной маржи (0: обычный аккаунт, 1: унифицированная маржа)
	TimeWindow          int    `json:"timeWindow"`          // Временное окно (устарело)
	SmpGroup            int    `json:"smpGroup"`            // SMP группа (устарело)
	IsMasterTrader      bool   `json:"isMasterTrader"`      // Является ли аккаунт мастер-аккаунтом
	MarginMode          string `json:"marginMode"`          // Режим маржи (ISOLATED_MARGIN, REGULAR_MARGIN и т.д.)
	SpotHedgingStatus   string `json:"spotHedgingStatus"`   // Статус хеджирования спот-позиций
	UpdatedTime         string `json:"updatedTime"`         // Время последнего обновления
	DcpStatus           string `json:"dcpStatus"`           // Статус DCP (устарело)
}

// WalletAccountInfo содержит информацию о кошельке
type WalletAccountInfo struct {
	AccountType            string     `json:"accountType"`            // Тип аккаунта (UNIFIED, CONTRACT, SPOT)
	AccountLTV             string     `json:"accountLTV"`             // Устаревшее поле (всегда 0)
	AccountIMRate          string     `json:"accountIMRate"`          // Коэффициент начальной маржи
	AccountMMRate          string     `json:"accountMMRate"`          // Коэффициент поддерживающей маржи
	TotalEquity            string     `json:"totalEquity"`            // Общий капитал
	TotalWalletBalance     string     `json:"totalWalletBalance"`     // Общий баланс кошелька
	TotalMarginBalance     string     `json:"totalMarginBalance"`     // Общий маржинальный баланс
	TotalAvailableBalance  string     `json:"totalAvailableBalance"`  // Доступный баланс
	TotalPerpUPL           string     `json:"totalPerpUPL"`           // Нереализованный P&L
	TotalInitialMargin     string     `json:"totalInitialMargin"`     // Общая начальная маржа
	TotalMaintenanceMargin string     `json:"totalMaintenanceMargin"` // Общая поддерживающая маржа
	Coins                  []CoinInfo `json:"coin"`                   // Информация по монетам
}

// CoinInfo содержит информацию о конкретной монете
type CoinInfo struct {
	Coin                string `json:"coin"`                // Название монеты
	Equity              string `json:"equity"`              // Капитал монеты
	UsdValue            string `json:"usdValue"`            // Стоимость в USD
	WalletBalance       string `json:"walletBalance"`       // Баланс кошелька
	Free                string `json:"free"`                // Доступный баланс
	Locked              string `json:"locked"`              // Заблокированный баланс
	SpotHedgingQty      string `json:"spotHedgingQty"`      // Количество для хеджирования
	BorrowAmount        string `json:"borrowAmount"`        // Сумма займа
	AvailableToWithdraw string `json:"availableToWithdraw"` // Доступно для вывода
	AccruedInterest     string `json:"accruedInterest"`     // Начисленные проценты
	TotalOrderIM        string `json:"totalOrderIM"`        // Занятая маржа для ордеров
	TotalPositionIM     string `json:"totalPositionIM"`     // Начальная маржа позиций
	TotalPositionMM     string `json:"totalPositionMM"`     // Поддерживающая маржа позиций
	UnrealisedPnl       string `json:"unrealisedPnl"`       // Нереализованный P&L
	CumRealisedPnl      string `json:"cumRealisedPnl"`      // Накопленный реализованный P&L
	Bonus               string `json:"bonus"`               // Бонусы
	AvailableToBorrow   string `json:"availableToBorrow"`   // Доступно для займа
	MarginCollateral    bool   `json:"marginCollateral"`    // Использование как залога (платформа)
	CollateralSwitch    bool   `json:"collateralSwitch"`    // Использование как залога (пользователь)
}

// WalletBalance представляет ответ API с информацией о балансе кошелька
type WalletBalance struct {
	List []WalletAccountInfo `json:"list"`
}
