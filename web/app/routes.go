package app

import (
	"fmt"
	cryptosdb "goTradingBot/external/cryptos/db"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func terminalRootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		content, err := os.ReadFile("web/pages/404.html")
		if err != nil {
			http.Error(w, "Не удалось загрузить страницу", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		w.Write(content)
		return
	}
	content, err := os.ReadFile("web/pages/terminal.html")
	if err != nil {
		http.Error(w, "Не удалось загрузить страницу", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func orderLogRootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		content, err := os.ReadFile("web/pages/404.html")
		if err != nil {
			http.Error(w, "Не удалось загрузить страницу", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		w.Write(content)
		return
	}
	content, err := os.ReadFile("web/pages/order.log.html")
	if err != nil {
		http.Error(w, "Не удалось загрузить страницу", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "pong")
}

func getCryptoImgHandler(w http.ResponseWriter, r *http.Request) {
	s := strings.TrimPrefix(r.URL.Path, "/static/img/crypto/")
	if s == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	crypto, err := cryptosdb.GetCryptoByID(int(id))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(crypto.Logo)
}
