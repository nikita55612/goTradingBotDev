package app

import (
	"encoding/json"
	"fmt"
	"goTradingBot/cdl"
	cryptosdb "goTradingBot/external/cryptos/db"
	"goTradingBot/predict"
	"goTradingBot/predict/portal"
	orderdb "goTradingBot/trading/db"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type apiResponse struct {
	Result any    `json:"result"`
	Error  string `json:"error"`
}

func getCryptoHandler(w http.ResponseWriter, r *http.Request) {
	res := new(apiResponse)
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	if !query.Has("s") {
		w.WriteHeader(http.StatusBadRequest)
		res.Error = "пропущен обязательный параметр запроса: s"
		data, _ := json.Marshal(res)
		w.Write(data)
		return
	}
	search := query.Get("s")
	crypto, err := cryptosdb.FindCrypto(search)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res.Error = fmt.Sprintf("по запросу '%s' ничего не найдено", search)
		json.NewEncoder(w).Encode(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res.Result = map[string]any{
		"id":     crypto.ID,
		"name":   crypto.Name,
		"symbol": crypto.Symbol,
	}
	json.NewEncoder(w).Encode(res)
}

func getCryptoLiteDetailHandler(w http.ResponseWriter, r *http.Request) {
	res := new(apiResponse)
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	if !query.Has("id") {
		w.WriteHeader(http.StatusBadRequest)
		res.Error = "пропущен обязательный параметр запроса: id"
		data, _ := json.Marshal(res)
		w.Write(data)
		return
	}
	id, err := strconv.ParseInt(query.Get("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	cryptoDetail, err := state.cryptos.GetCryptoLiteDetail(int(id))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res.Result = cryptoDetail
	json.NewEncoder(w).Encode(res)
}

func getCryptoList(w http.ResponseWriter, r *http.Request) {
	res := new(apiResponse)
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	limit := 100
	if query.Has("l") {
		if l, err := strconv.ParseInt(query.Get("l"), 10, 64); err == nil {
			limit = int(l)
		}
	}
	cryptoList, err := state.cryptos.GetCryptoList(limit)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res.Result = cryptoList
	json.NewEncoder(w).Encode(res)
}

func getCryptoFearAndGreed(w http.ResponseWriter, r *http.Request) {
	res := new(apiResponse)
	w.Header().Set("Content-Type", "application/json")
	fearAndGreed, err := state.cryptos.GetFearAndGreedMetrics()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res.Result = fearAndGreed
	json.NewEncoder(w).Encode(res)
}

func getCryptoDetailHandler(w http.ResponseWriter, r *http.Request) {
	res := new(apiResponse)
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	if !query.Has("id") {
		w.WriteHeader(http.StatusBadRequest)
		res.Error = "пропущен обязательный параметр запроса: id"
		data, _ := json.Marshal(res)
		w.Write(data)
		return
	}
	id, err := strconv.ParseInt(query.Get("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	cryptoDetail, err := state.cryptos.GetCryptoFullDetail(int(id))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res.Result = cryptoDetail
	json.NewEncoder(w).Encode(res)
}

func getCandlesHandler(w http.ResponseWriter, r *http.Request) {
	res := new(apiResponse)
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	if !query.Has("s") {
		w.WriteHeader(http.StatusBadRequest)
		res.Error = "пропущен обязательный параметр запроса: s"
		json.NewEncoder(w).Encode(res)
		return
	}
	if !query.Has("i") {
		w.WriteHeader(http.StatusBadRequest)
		res.Error = "пропущен обязательный параметр запроса: i"
		json.NewEncoder(w).Encode(res)
		return
	}
	coin := query.Get("s")
	interval, err := cdl.ParseInterval(query.Get("i"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	limit, err := strconv.ParseInt(query.Get("l"), 10, 64)
	if err != nil || limit > 999 {
		limit = 999
	}
	symbol := coin + "USDT"
	candles, err := state.cdlProvider.GetCandles(symbol, interval, int(limit)+1)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	if len(candles) == 0 {
		w.WriteHeader(http.StatusNotFound)
		res.Error = "пустой список свечей"
		json.NewEncoder(w).Encode(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res.Result = cdl.CandlesToRawData(candles[:len(candles)-1])
	json.NewEncoder(w).Encode(res)
}

func getCurrentCandleHandler(w http.ResponseWriter, r *http.Request) {
	res := new(apiResponse)
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	if !query.Has("s") {
		w.WriteHeader(http.StatusBadRequest)
		res.Error = "пропущен обязательный параметр запроса: s"
		json.NewEncoder(w).Encode(res)
		return
	}
	if !query.Has("i") {
		w.WriteHeader(http.StatusBadRequest)
		res.Error = "пропущен обязательный параметр запроса: i"
		data, _ := json.Marshal(res)
		w.Write(data)
		return
	}
	coin := query.Get("s")
	interval, err := cdl.ParseInterval(query.Get("i"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	symbol := coin + "USDT"
	candles, err := state.cdlProvider.GetCandles(symbol, interval, 1)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	if len(candles) == 0 {
		w.WriteHeader(http.StatusNotFound)
		res.Error = "пустой список свечей"
		json.NewEncoder(w).Encode(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res.Result = candles[len(candles)-1].AsArr()
	json.NewEncoder(w).Encode(res)
}

func getTrendPredictHandler(w http.ResponseWriter, r *http.Request) {
	res := new(apiResponse)
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Content-Type") != "application/json" {
		res.Error = "Неверный формат запроса: ожидается application/json"
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(res)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		res.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(res)
		return
	}
	defer r.Body.Close()
	var candlesRawData [][7]string
	if err := json.Unmarshal(body, &candlesRawData); err != nil {
		res.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(res)
		return
	}
	candles, err := cdl.CandlesFromRawData(candlesRawData)
	if err != nil {
		res.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(res)
		return
	}
	if len(candles) == 0 {
		w.WriteHeader(http.StatusNotFound)
		res.Error = "пустой список свечей"
		json.NewEncoder(w).Encode(res)
		return
	}
	n := len(candles)
	if n <= predict.FeatureOffset {
		res.Error = fmt.Sprintf("недостаточно данных для предсказания: %d <= %d", n, predict.FeatureOffset)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(res)
		return
	}
	features := state.fgModels[predict.A6N21P9].GenTranspose(candles, predict.FeatureOffset, -1)
	query := r.URL.Query()
	var markings []string
	if query.Has("m") {
		m := query.Get("m")
		markings = strings.Split(m, ",")
	}
	predict, err := portal.GetPrediction(features, markings...).Unwrap()
	if err != nil {
		res.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res.Result = predict
	json.NewEncoder(w).Encode(res)
}

func getOrderLog(w http.ResponseWriter, r *http.Request) {
	res := new(apiResponse)
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	if !query.Has("p") {
		w.WriteHeader(http.StatusBadRequest)
		res.Error = "пропущен обязательный параметр запроса: p"
		json.NewEncoder(w).Encode(res)
		return
	}
	periodSec, err := strconv.ParseInt(query.Get("p"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	orders, err := orderdb.GetOrderRequestsByPeriod(periodSec)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res.Result = orders
	json.NewEncoder(w).Encode(res)
}
