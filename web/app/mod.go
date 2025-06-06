package app

import (
	"net/http"
)

func RunTerminal(addr string) error {
	initAppState()

	mux := http.NewServeMux()
	mux.HandleFunc("/", terminalRootHandler)
	mux.HandleFunc("/terminal", terminalRootHandler)
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/api/v1/crypto", getCryptoHandler)
	mux.HandleFunc("/api/v1/crypto/list", getCryptoList)
	mux.HandleFunc("/api/v1/crypto/fearAndGreed", getCryptoFearAndGreed)
	mux.HandleFunc("/api/v1/crypto/detail", getCryptoDetailHandler)
	mux.HandleFunc("/api/v1/crypto/detail/lite", getCryptoLiteDetailHandler)
	mux.HandleFunc("/api/v1/candles", getCandlesHandler)
	mux.HandleFunc("/api/v1/candle", getCurrentCandleHandler)
	mux.HandleFunc("/api/v1/predict/trend", getTrendPredictHandler)
	mux.HandleFunc("/static/img/crypto/", getCryptoImgHandler)

	assets := http.FileServer(http.Dir("./web/assets/"))
	mux.Handle("/assets/", http.StripPrefix("/assets", assets))

	return http.ListenAndServe(addr, mux)
}

func RunOrderLog(addr string) error {
	initAppState()

	mux := http.NewServeMux()
	mux.HandleFunc("/", orderLogRootHandler)
	mux.HandleFunc("/order-log", orderLogRootHandler)
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/api/v1/crypto", getCryptoHandler)
	mux.HandleFunc("/static/img/crypto/", getCryptoImgHandler)
	mux.HandleFunc("/api/v1/order-log", getOrderLog)

	assets := http.FileServer(http.Dir("./web/assets/"))
	mux.Handle("/assets/", http.StripPrefix("/assets", assets))

	return http.ListenAndServe(addr, mux)
}
