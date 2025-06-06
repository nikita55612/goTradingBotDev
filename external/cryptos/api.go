package cryptos

import (
	"fmt"
	"goTradingBot/external/cryptos/models"
	"goTradingBot/httpx"
	"math"
	"net/url"
	"strconv"
	"time"
)

func (c *Client) GetCryptoList(limit int) ([]models.CryptoInfo, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("CoinMarketCapAPI: GetCryptoList limit <= 0")
	}
	u, _ := url.Parse("https://api.coinmarketcap.com/data-api/v3/cryptocurrency/listing")
	cryptosList := make([]models.CryptoInfo, 0, limit)
	query := u.Query()
	query.Set("convert", "USD")
	query.Set("aux", "date_added,high24h,low24h")
	start := 1
	for i := 1; i <= int(math.Ceil(float64(limit)/100.0)); i++ {
		req := httpx.Get(u.String()).
			SetQueryParam("start", strconv.Itoa(start)).
			SetQueryParam("limit", strconv.Itoa(min(100, limit-100*(i-1))))
		var cryptoCurrencyListing models.CryptoCurrencyListing
		if err := c.callAPI(req, &cryptoCurrencyListing); err != nil {
			return cryptosList, fmt.Errorf("GetCryptoList: %w", err)
		}
		extractCryptos := cryptoCurrencyListing.ExtractCryptos()
		if len(extractCryptos) == 0 {
			break
		}
		cryptosList = append(cryptosList, extractCryptos...)
		start += 100
	}
	return cryptosList, nil
}

func (c *Client) GetFearAndGreedChart(limit int) (*models.FearAndGreedChart, error) {
	now := int(time.Now().UTC().Unix())
	period := (time.Hour*24)*time.Duration(limit) + time.Hour*23
	req := httpx.Get("https://api.coinmarketcap.com/data-api/v3/fear-greed/chart").
		SetQueryParam("start", strconv.Itoa(now-int(period.Seconds()))).
		SetQueryParam("end", strconv.Itoa(now))
	var fearAndGreedChart models.FearAndGreedChart
	if err := c.callAPI(req, &fearAndGreedChart); err != nil {
		return nil, fmt.Errorf("GetFearAndGreedMetrics: %w", err)
	}
	return &fearAndGreedChart, nil
}

func (c *Client) GetFearAndGreedMetrics() (*models.FearAndGreedMetrics, error) {
	fearAndGreedChart, err := c.GetFearAndGreedChart(1)
	if err != nil {
		return nil, err
	}
	return fearAndGreedChart.ExtractMetrics(), nil
}

func (c *Client) getCMC100Chart(rng string) (*models.CMC100Chart, error) {
	req := httpx.Get("https://api.coinmarketcap.com/data-api/v3/top100/historical/chart").
		SetQueryParam("range", rng)
	var cmc100Chart models.CMC100Chart
	if err := c.callAPI(req, &cmc100Chart); err != nil {
		return nil, fmt.Errorf("GetCMC100Chart24H: %w", err)
	}
	return &cmc100Chart, nil
}

func (c *Client) GetCMC100Chart24H() (*models.CMC100Chart, error) {
	return c.getCMC100Chart("24h")
}

func (c *Client) GetCMC100Chart7D() (*models.CMC100Chart, error) {
	return c.getCMC100Chart("7d")
}

func (c *Client) GetCryptoLiteDetail(cryptoId int) (*models.CryptoLiteDetail, error) {
	req := httpx.Get("https://api.coinmarketcap.com/data-api/v3/cryptocurrency/detail/lite?id=1").
		SetQueryParam("id", strconv.Itoa(cryptoId))
	var cryptoLiteDetail models.CryptoLiteDetail
	if err := c.callAPI(req, &cryptoLiteDetail); err != nil {
		return nil, fmt.Errorf("GetCryptoLiteDetail: %w", err)
	}
	return &cryptoLiteDetail, nil
}

func (c *Client) GetCryptoFullDetail(cryptoId int) (*models.CryptoFullDetail, error) {
	req := httpx.Get("https://api.coinmarketcap.com/data-api/v3/cryptocurrency/detail").
		SetQueryParam("id", strconv.Itoa(cryptoId))
	var cryptoFullDetail models.CryptoFullDetail
	if err := c.callAPI(req, &cryptoFullDetail); err != nil {
		return nil, fmt.Errorf("GetCryptoLiteDetail: %w", err)
	}
	return &cryptoFullDetail, nil
}
