package models

type Status struct {
	Timestamp    string `json:"timestamp"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Elapsed      string `json:"elapsed"`
	CreditCount  int    `json:"credit_count"`
}

type CryptoQuote struct {
	Price                 float64 `json:"price"`
	Volume24h             float64 `json:"volume24h"`
	MarketCap             float64 `json:"marketCap"`
	PercentChange1h       float64 `json:"percentChange1h"`
	PercentChange24h      float64 `json:"percentChange24h"`
	PercentChange7d       float64 `json:"percentChange7d"`
	PercentChange30d      float64 `json:"percentChange30d"`
	PercentChange60d      float64 `json:"percentChange60d"`
	PercentChange90d      float64 `json:"percentChange90d"`
	FullyDilutedMarketCap float64 `json:"fullyDilluttedMarketCap"`
	Name                  string  `json:"name"`
	LastUpdated           string  `json:"lastUpdated"`
}

type CryptoCurrency struct {
	ID          int           `json:"id"`
	IsActive    int           `json:"isActive"`
	High24h     float64       `json:"high24h"`
	Low24h      float64       `json:"low24h"`
	Name        string        `json:"name"`
	Symbol      string        `json:"symbol"`
	Slug        string        `json:"slug"`
	LastUpdated string        `json:"lastUpdated"`
	DateAdded   string        `json:"dateAdded"`
	Quotes      []CryptoQuote `json:"quotes"`
}

type CryptoCurrencyListing struct {
	CryptoCurrencyList []CryptoCurrency `json:"cryptoCurrencyList"`
	TotalCount         string           `json:"totalCount"`
}

type CryptoInfo struct {
	Price                 float64 `json:"price"`
	Volume24h             float64 `json:"volume24h"`
	MarketCap             float64 `json:"marketCap"`
	High24h               float64 `json:"high24h"`
	Low24h                float64 `json:"low24h"`
	PercentChange1h       float64 `json:"percentChange1h"`
	PercentChange24h      float64 `json:"percentChange24h"`
	PercentChange7d       float64 `json:"percentChange7d"`
	PercentChange30d      float64 `json:"percentChange30d"`
	PercentChange60d      float64 `json:"percentChange60d"`
	PercentChange90d      float64 `json:"percentChange90d"`
	FullyDilutedMarketCap float64 `json:"fullyDilutedMarketCap"`
	ID                    int     `json:"id"`
	IsActive              int     `json:"isActive"`
	Name                  string  `json:"name"`
	Symbol                string  `json:"symbol"`
	Slug                  string  `json:"slug"`
	LastUpdated           string  `json:"lastUpdated"`
	DateAdded             string  `json:"dateAdded"`
}

func (c *CryptoCurrencyListing) ExtractCryptos() []CryptoInfo {
	n := len(c.CryptoCurrencyList)
	cryptos := make([]CryptoInfo, 0, n)
	for _, crypto := range c.CryptoCurrencyList {
		if len(crypto.Quotes) == 0 {
			continue
		}
		quote := crypto.Quotes[0]
		cryptos = append(cryptos, CryptoInfo{
			ID:                    crypto.ID,
			Name:                  crypto.Name,
			Symbol:                crypto.Symbol,
			Slug:                  crypto.Slug,
			High24h:               crypto.High24h,
			Low24h:                crypto.Low24h,
			IsActive:              crypto.IsActive,
			LastUpdated:           crypto.LastUpdated,
			DateAdded:             crypto.DateAdded,
			Price:                 quote.Price,
			Volume24h:             quote.Volume24h,
			MarketCap:             quote.MarketCap,
			PercentChange1h:       quote.PercentChange1h,
			PercentChange24h:      quote.PercentChange24h,
			PercentChange7d:       quote.PercentChange7d,
			PercentChange30d:      quote.PercentChange30d,
			PercentChange60d:      quote.PercentChange60d,
			PercentChange90d:      quote.PercentChange90d,
			FullyDilutedMarketCap: quote.FullyDilutedMarketCap,
		})
	}
	return cryptos
}

type CMC100Chart struct {
	Values []CMC100Value `json:"values"`
}

type CMC100Value struct {
	Value     float64 `json:"value"`
	Timestamp string  `json:"timestamp"`
}

type VoteResult struct {
	Bullish int    `json:"bullish"`
	Bearish int    `json:"bearish"`
	Total   int    `json:"total"`
	Votable bool   `json:"votable"`
	MyVote  string `json:"myVote"`
}

type CryptoStatistics struct {
	Price                    float64 `json:"price"`
	PriceChangePercentage24h float64 `json:"priceChangePercentage24h"`
	MarketCap                float64 `json:"marketCap"`
	CirculatingSupply        float64 `json:"circulatingSupply"`
	TotalSupply              float64 `json:"totalSupply"`
	MaxSupply                float64 `json:"maxSupply"`
	Rank                     int     `json:"rank"`
}

type CryptoLiteDetail struct {
	ID         int              `json:"id"`
	Volume     float64          `json:"volume"`
	Statistics CryptoStatistics `json:"statistics"`
	Name       string           `json:"name"`
	Symbol     string           `json:"symbol"`
	Slug       string           `json:"slug"`
	Status     string           `json:"status"`
	WatchCount string           `json:"watchCount"`
}

type CryptoFullDetail struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Slug     string `json:"slug"`
	Category string `json:"category"`
	// Description      string  `json:"description"`
	DateAdded        string  `json:"dateAdded"`
	ActualTimeStart  string  `json:"actualTimeStart"`
	Status           string  `json:"status"`
	SubStatus        string  `json:"subStatus"`
	Notice           string  `json:"notice"`
	AlertType        int     `json:"alertType"`
	AlertLink        string  `json:"alertLink"`
	LatestUpdateTime string  `json:"latestUpdateTime"`
	WatchCount       string  `json:"watchCount"`
	WatchListRanking int     `json:"watchListRanking"`
	DateLaunched     string  `json:"dateLaunched"`
	LatestAdded      bool    `json:"latestAdded"`
	LaunchPrice      float64 `json:"launchPrice"`
	// CryptoTags                   []CryptoTag               `json:"tags"`
	Urls                      CryptoUrls           `json:"urls"`
	Volume                    float64              `json:"volume"`
	VolumeChangePercentage24h float64              `json:"volumeChangePercentage24h"`
	CexVolume                 float64              `json:"cexVolume"`
	DexVolume                 float64              `json:"dexVolume"`
	Statistics                CryptoFullStatistics `json:"statistics"`
	// Quotes                    []any                `json:"quotes"`
	// RelatedCoins                 []CryptoRelatedCoin       `json:"relatedCoins"`
	// Wallets    []CryptoWallet `json:"wallets"`
	IsAudited bool `json:"isAudited"`
	// AuditInfos []any `json:"auditInfos"`
	// Holders                      CryptoHolders             `json:"holders"`
	DisplayTV           int    `json:"displayTV"`
	IsInfiniteMaxSupply int    `json:"isInfiniteMaxSupply"`
	TVCoinSymbol        string `json:"tvCoinSymbol"`
	// SupportWalletInfos           []CryptoSupportWalletInfo `json:"supportWalletInfos"`
	CdpTotalHolder               string         `json:"cdpTotalHolder"`
	HolderHistoricalFlag         bool           `json:"holderHistoricalFlag"`
	HolderListFlag               bool           `json:"holderListFlag"`
	HoldersFlag                  bool           `json:"holdersFlag"`
	RatingsFlag                  bool           `json:"ratingsFlag"`
	AnalysisFlag                 bool           `json:"analysisFlag"`
	SocialsFlag                  bool           `json:"socialsFlag"`
	CirculatingSupplyBlockerFlag bool           `json:"circulatingSupplyBlockerFlag"`
	CryptoRating                 []CryptoRating `json:"cryptoRating"`
	Analysis                     CryptoAnalysis `json:"analysis"`
	// CoinBitesVideo               CoinBitesVideo            `json:"coinBitesVideo"`
	HasExtraInfoFlag bool `json:"hasExtraInfoFlag"`
	// EarnList                     []CryptoEarn              `json:"earnList"`
	Upcoming               CryptoUpcoming      `json:"upcoming"`
	AnnotationFlag         bool                `json:"annotationFlag"`
	SimilarCoins           []CryptoSimilarCoin `json:"similarCoins"`
	HasPcsFlag             bool                `json:"hasPcsFlag"`
	ProfileCompletionScore CryptoProfileScore  `json:"profileCompletionScore"`
}

// type CryptoTag struct {
// 	Slug     string `json:"slug"`
// 	Name     string `json:"name"`
// 	Category string `json:"category"`
// 	Status   int    `json:"status"`
// 	Priority int    `json:"priority"`
// }

type CryptoUrls struct {
	Website      []string `json:"website"`
	TechnicalDoc []string `json:"technical_doc"`
	Explorer     []string `json:"explorer"`
	SourceCode   []string `json:"source_code"`
	MessageBoard []string `json:"message_board"`
	Chat         []string `json:"chat"`
	Announcement []string `json:"announcement"`
	Reddit       []string `json:"reddit"`
	Facebook     []string `json:"facebook"`
	Twitter      []string `json:"twitter"`
}

type CryptoFullStatistics struct {
	Price                                    float64 `json:"price"`
	PriceChangePercentage1h                  float64 `json:"priceChangePercentage1h"`
	PriceChangePercentage24h                 float64 `json:"priceChangePercentage24h"`
	PriceChangePercentage7d                  float64 `json:"priceChangePercentage7d"`
	PriceChangePercentage30d                 float64 `json:"priceChangePercentage30d"`
	PriceChangePercentage60d                 float64 `json:"priceChangePercentage60d"`
	PriceChangePercentage90d                 float64 `json:"priceChangePercentage90d"`
	PriceChangePercentage1y                  float64 `json:"priceChangePercentage1y"`
	PriceChangePercentageAll                 float64 `json:"priceChangePercentageAll"`
	MarketCap                                float64 `json:"marketCap"`
	MarketCapChangePercentage24h             float64 `json:"marketCapChangePercentage24h"`
	FullyDilutedMarketCap                    float64 `json:"fullyDilutedMarketCap"`
	FullyDilutedMarketCapChangePercentage24h float64 `json:"fullyDilutedMarketCapChangePercentage24h"`
	CirculatingSupply                        float64 `json:"circulatingSupply"`
	TotalSupply                              float64 `json:"totalSupply"`
	MaxSupply                                float64 `json:"maxSupply"`
	MarketCapDominance                       float64 `json:"marketCapDominance"`
	ROI                                      float64 `json:"roi"`
	Low24h                                   float64 `json:"low24h"`
	High24h                                  float64 `json:"high24h"`
	Low7d                                    float64 `json:"low7d"`
	High7d                                   float64 `json:"high7d"`
	Low30d                                   float64 `json:"low30d"`
	High30d                                  float64 `json:"high30d"`
	Low90d                                   float64 `json:"low90d"`
	High90d                                  float64 `json:"high90d"`
	Low52w                                   float64 `json:"low52w"`
	High52w                                  float64 `json:"high52w"`
	LowAllTime                               float64 `json:"lowAllTime"`
	HighAllTime                              float64 `json:"highAllTime"`
	LowAllTimeChangePercentage               float64 `json:"lowAllTimeChangePercentage"`
	HighAllTimeChangePercentage              float64 `json:"highAllTimeChangePercentage"`
	LowYesterday                             float64 `json:"lowYesterday"`
	HighYesterday                            float64 `json:"highYesterday"`
	OpenYesterday                            float64 `json:"openYesterday"`
	CloseYesterday                           float64 `json:"closeYesterday"`
	PriceChangePercentageYesterday           float64 `json:"priceChangePercentageYesterday"`
	VolumeYesterday                          float64 `json:"volumeYesterday"`
	Turnover                                 float64 `json:"turnover"`
	YtdPriceChangePercentage                 float64 `json:"ytdPriceChangePercentage"`
	Rank                                     int     `json:"rank"`
	VolumeRank                               int     `json:"volumeRank"`
	VolumeMcRank                             int     `json:"volumeMcRank"`
	McTotalNum                               int     `json:"mcTotalNum"`
	VolumeTotalNum                           int     `json:"volumeTotalNum"`
	VolumeMcTotalNum                         int     `json:"volumeMcTotalNum"`
	LowAllTimeTimestamp                      string  `json:"lowAllTimeTimestamp"`
	HighAllTimeTimestamp                     string  `json:"highAllTimeTimestamp"`
}

// type CryptoRelatedCoin struct {
// 	ID                       int     `json:"id"`
// 	Name                     string  `json:"name"`
// 	Slug                     string  `json:"slug"`
// 	Price                    float64 `json:"price"`
// 	PriceChangePercentage24h float64 `json:"priceChangePercentage24h"`
// 	PriceChangePercentage7d  float64 `json:"priceChangePercentage7d"`
// }

// type CryptoWallet struct {
// 	ID            int     `json:"id"`
// 	Name          string  `json:"name"`
// 	Tier          int     `json:"tier"`
// 	URL           string  `json:"url"`
// 	Chains        string  `json:"chains"`
// 	Types         string  `json:"types"`
// 	Introduction  string  `json:"introduction"`
// 	Star          float64 `json:"star"`
// 	Security      int     `json:"security"`
// 	EasyToUse     int     `json:"easyToUse"`
// 	Decentration  bool    `json:"decentration"`
// 	FocusNumber   int     `json:"focusNumber"`
// 	Rank          int     `json:"rank"`
// 	Logo          string  `json:"logo"`
// 	MultipleChain bool    `json:"multipleChain"`
// }

// type CryptoSupportWalletInfo struct {
// 	ID            int    `json:"id"`
// 	Name          string `json:"name"`
// 	URL           string `json:"url"`
// 	Chains        string `json:"chains"`
// 	Decentration  bool   `json:"decentration"`
// 	Logo          string `json:"logo"`
// 	MultipleChain bool   `json:"multipleChain"`
// }

// type CryptoHolders struct {
// 	HolderCount           int      `json:"holderCount"`
// 	DailyActive           int      `json:"dailyActive"`
// 	HolderList            []Holder `json:"holderList"`
// 	TopTenHolderRatio     float64  `json:"topTenHolderRatio"`
// 	TopTwentyHolderRatio  float64  `json:"topTwentyHolderRatio"`
// 	TopFiftyHolderRatio   float64  `json:"topFiftyHolderRatio"`
// 	TopHundredHolderRatio float64  `json:"topHundredHolderRatio"`
// }

// type Holder struct {
// 	Address string  `json:"address"`
// 	Balance float64 `json:"balance"`
// 	Share   float64 `json:"share"`
// }

type CryptoRating struct {
	Score      float64 `json:"score"`
	Rating     float64 `json:"rating"`
	Type       string  `json:"type"`
	UpdateTime string  `json:"updateTime"`
	Link       string  `json:"link"`
}

type CryptoAnalysis struct {
	HoldingWhalesPercent  float64       `json:"holdingWhalesPercent"`
	HoldingAddressesCount int           `json:"holdingAddressesCount"`
	AddressByTimeHeld     AddressByTime `json:"addressByTimeHeld"`
}

type AddressByTime struct {
	HoldersPercent  float64 `json:"holdersPercent"`
	CruisersPercent float64 `json:"cruisersPercent"`
	TradersPercent  float64 `json:"tradersPercent"`
}

// type CoinBitesVideo struct {
// 	ID           string `json:"id"`
// 	Category     string `json:"category"`
// 	VideoURL     string `json:"videoUrl"`
// 	Title        string `json:"title"`
// 	Description  string `json:"description"`
// 	PreviewImage string `json:"previewImage"`
// }

// type CryptoEarn struct {
// 	ID        string       `json:"id"`
// 	Rank      int          `json:"rank"`
// 	Provider  EarnProvider `json:"provider"`
// 	APR       []float64    `json:"apr"`
// 	NetAPY    []float64    `json:"netApy"`
// 	Fee       []float64    `json:"fee"`
// 	TypeName  string       `json:"typeName"`
// 	Type      string       `json:"type"`
// 	SubType   string       `json:"subType"`
// 	Ecosystem string       `json:"ecosystem"`
// }

type EarnProvider struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CryptoUpcoming struct {
	Status bool `json:"status"`
}

type CryptoSimilarCoin struct {
	ID     int    `json:"id"`
	Symbol string `json:"symbol"`
}

type CryptoProfileScore struct {
	Point               float64 `json:"point"`
	Percentage          float64 `json:"percentage"`
	PcsResultUpdateTime int64   `json:"pcsResultUpdateTime"`
	HiddenPcs           bool    `json:"hiddenPcs"`
}
