package portal

import (
	"context"
	"errors"
	"fmt"
	"goTradingBot/httpx"
	"goTradingBot/pyexec"
	"strings"
	"sync"
	"time"
)

var (
	portalAddr string = "localhost:8666"
	process    *pyexec.PyProcess
	mu         sync.Mutex
)

// SetAddr устанавливает адрес для сервера portal
// Формат addr: "host:port" (например, localhost:8083)
func SetAddr(addr string) {
	mu.Lock()
	defer mu.Unlock()
	portalAddr = addr
}

// Start запускает процесс portal с контекстом
func StartWithContext(ctx context.Context) error {
	if err := Start(); err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		Stop()
	}()
	return nil
}

// Start запускает процесс portal и проверяет его доступность
func Start() error {
	mu.Lock()
	defer mu.Unlock()

	if process != nil {
		return fmt.Errorf("процесс portal уже запущен")
	}
	parts := strings.Split(portalAddr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("неверный формат portalAddr: %s (ожидается 'host:port')", portalAddr)
	}
	host, port := parts[0], parts[1]
	p, err := pyexec.NewPyProcess(
		"neuralab",
		pyexec.WithScriptName("portal.py"),
		pyexec.WithArgs("-H", host, "-P", port),
	)
	if err != nil {
		return fmt.Errorf("не удалось создать процесс portal: %w", err)
	}
	process = p
	if err := process.Start(); err != nil {
		process = nil
		return fmt.Errorf("не удалось запустить процесс portal: %w", err)
	}
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(5 * time.Minute)
	for {
		select {
		case <-ticker.C:
			if ping() == "pong" {
				return nil
			}
		case <-deadline:
			process.Stop()
			process = nil
			return fmt.Errorf("превышено время ожидания запуска процесса portal (2 минуты)")
		}
	}
}

// Stop останавливает процесс portal
func Stop() {
	mu.Lock()
	defer mu.Unlock()

	if process != nil {
		process.Stop()
		process = nil
	}
}

// Restart перезапускает процесс portal
func Restart() error {
	mu.Lock()
	defer mu.Unlock()

	if process != nil {
		process.Stop()
		process = nil
	}
	return Start()
}

// Request - запрос к порталу для получения предсказаний
type Request struct {
	Features [][]float64 `json:"features"` // Массив признаков для предсказания
	Markings []string    `json:"markings"` // Список меток для предсказания
}

// Response - ответ от портала с предсказаниями
type Response struct {
	Predict map[string][]float64 `json:"predict"` // Результаты предсказаний
	Error   string               `json:"error"`   // Описание ошибки, если возникла
}

// Unwrap возвращает результаты предсказаний или ошибку, если она была
func (pr *Response) Unwrap() (map[string][]float64, error) {
	if pr.Error != "" {
		return nil, errors.New(pr.Error)
	}
	return pr.Predict, nil
}

// UnwrapPredict возвращает результат конкретного предсказаня по вхождению строки
func (pr *Response) UnwrapPredict(contains string) ([]float64, error) {
	if pr.Error != "" {
		return nil, errors.New(pr.Error)
	}
	for k, v := range pr.Predict {
		if strings.Contains(k, contains) {
			return v, nil
		}
	}
	return nil, fmt.Errorf("UnwrapPredict: не найдено вхождений: %q", contains)
}

// UnwrapSinglePredict возвращает первое предсказание
func (pr *Response) UnwrapSinglePredict() ([]float64, error) {
	if pr.Error != "" {
		return nil, errors.New(pr.Error)
	}
	for _, v := range pr.Predict {
		return v, nil
	}
	return nil, fmt.Errorf("UnwrapSinglePredict: нет предсказаний")
}

// ping проверяет доступность портала
// Возвращает "pong" если portal доступен, иначе пустую строку
func ping() string {
	fullURL := fmt.Sprintf("http://%s/ping", portalAddr)
	res := httpx.Get(fullURL).Do()
	defer res.Close()
	if res.Error() != nil {
		return ""
	}
	body, err := res.ReadBody()
	if err != nil {
		return ""
	}
	return string(body)
}

// GetPrediction получает предсказания от портала для переданных признаков и меток
// features - массив признаков для предсказания
// markings - список меток для выбора моделей
func GetPrediction(features [][]float64, markings ...string) *Response {
	if len(markings) == 0 {
		markings = []string{"+"}
	}
	portalRequestObj := &Request{
		Features: features,
		Markings: markings,
	}
	mu.Lock()
	fullURL := fmt.Sprintf("http://%s/predict", portalAddr)
	mu.Unlock()
	res := httpx.Post(fullURL).
		WithJsonData(portalRequestObj).
		AddHeader("Content-Type", "application/json").
		Do()
	defer res.Close()
	if err := res.Error(); err != nil {
		return &Response{
			Error: fmt.Sprintf("GetPrediction: не удалось выполнить запрос: %v", err),
		}
	}
	var portalResponse Response
	if err := res.UnmarshalBody(&portalResponse); err != nil {
		return &Response{
			Error: fmt.Sprintf("GetPrediction: не удалось разобрать ответ сервера: %v", err),
		}
	}
	return &portalResponse
}
