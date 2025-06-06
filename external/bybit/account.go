package bybit

import (
	"fmt"
	"goTradingBot/external/bybit/models"
	"goTradingBot/httpx"
	"net/url"
	"strings"
)

// GetAccountInfo возвращает основную информацию об аккаунте
func (c *Client) GetAccountInfo() (*models.AccountInfo, *Error) {
	req := httpx.Get(c.baseURL + "/v5/account/info")
	var accountInfo models.AccountInfo
	if err := c.callAPI(req, "", &accountInfo); err != nil {
		return &accountInfo, err.SetEndpoint("GetAccountInfo")
	}
	return &accountInfo, nil
}

// GetWalletBalance возвращает баланс унифицированного кошелька
// coins - необязательный список монет для фильтрации (если не указаны - возвращаются все)
func (c *Client) GetWalletBalance(coins ...string) (*models.WalletAccountInfo, *Error) {
	params := map[string]any{
		"accountType": "UNIFIED",
	}
	if len(coins) > 0 {
		params["coin"] = strings.Join(coins, ",")
	}
	res, err := c.getWalletBalance(params)
	if err != nil {
		return nil, err
	}
	return extractWalletBalance(res), nil
}

// GetCoinBalance возвращает информацию о балансе конкретной монеты
// coin - символ монеты (например, "BTC")
func (c *Client) GetCoinBalance(coin string) (*models.CoinInfo, *Error) {
	unifiedWalletBalance, err := c.GetWalletBalance(coin)
	if err != nil {
		return nil, err
	}
	return extractCoinFromWallet(unifiedWalletBalance, coin), nil
}

// GetMultipleCoinsBalance возвращает баланс для нескольких монет
// coins - список символов монет (например, ["BTC", "USDT"])
func (c *Client) GetMultipleCoinsBalance(coins ...string) (map[string]*models.CoinInfo, *Error) {
	unifiedWalletBalance, err := c.GetWalletBalance(coins...)
	if err != nil {
		return nil, err
	}
	return extractCoinsFromWallet(unifiedWalletBalance), nil
}

// getWalletBalance внутренний метод для получения баланса кошелька
func (c *Client) getWalletBalance(params map[string]any) (*models.WalletBalance, *Error) {
	query := make(url.Values)
	for k, v := range params {
		query.Add(k, fmt.Sprintf("%v", v))
	}
	queryString := query.Encode()
	fullURL := fmt.Sprintf(
		"%s%s?%s",
		c.baseURL,
		"/v5/account/wallet-balance",
		queryString,
	)
	req := httpx.Get(fullURL)
	var walletBalance models.WalletBalance
	if err := c.callAPI(req, queryString, &walletBalance); err != nil {
		return &walletBalance, err.SetEndpoint("getWalletBalance")
	}
	return &walletBalance, nil
}
