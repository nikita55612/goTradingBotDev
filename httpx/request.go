package httpx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync/atomic"
)

// Request представляет собой HTTP-запрос и методы для работы с ответом
// Выполнение запроса происходит автоматически при вызове любого из методов работы с ответом
// Запрос выполняется один раз (до успешной попытки)
// Запрос можно выполнить принудительно, используя функцию Execute
type Request struct {
	req       *http.Request
	err       error
	client    *http.Client
	resp      *http.Response
	cancel    context.CancelFunc
	cancelled atomic.Bool
}

// executeIfNeeded выполняет запрос, если он еще не был выполнен
func (r *Request) executeIfNeeded() {
	if r.err != nil || r.resp != nil {
		return
	}
	r.Execute()
}

// Execute выполняет запрос принудительно
func (r *Request) Execute() (*http.Response, error) {
	r.resp, r.err = r.client.Do(r.req)
	return r.resp, r.err
}

// StatusCode возвращает код статуса HTTP-ответа
func (r *Request) StatusCode() int {
	r.executeIfNeeded()
	if r.err != nil || r.resp == nil {
		return 0
	}
	return r.resp.StatusCode
}

// StatusText возвращает текстовое описание статуса HTTP-ответа
func (r *Request) StatusText() string {
	r.executeIfNeeded()
	if r.err != nil || r.resp == nil {
		return ""
	}
	return http.StatusText(r.resp.StatusCode)
}

// IsSuccess проверяет, был ли запрос успешным (статус 2xx)
func (r *Request) IsSuccess() bool {
	code := r.StatusCode()
	return code >= 200 && code < 300
}

// IsClientError проверяет, была ли ошибка на стороне клиента (статус 4xx)
func (r *Request) IsClientError() bool {
	code := r.StatusCode()
	return code >= 400 && code < 500
}

// IsServerError проверяет, была ли ошибка на стороне сервера (статус 5xx)
func (r *Request) IsServerError() bool {
	code := r.StatusCode()
	return code >= 500 && code < 600
}

// ReadBody читает тело ответа и возвращает его в виде байтов
func (r *Request) ReadBody() ([]byte, error) {
	r.executeIfNeeded()
	if r.err != nil {
		return nil, r.err
	}
	return io.ReadAll(r.resp.Body)
}

// UnmarshalBody читает и декодирует JSON-тело ответа в переданную переменную
func (r *Request) UnmarshalBody(v any) error {
	body, err := r.ReadBody()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// Error возвращает ошибку, возникшую при строительстве или выполнении запроса
func (r *Request) Error() error {
	r.executeIfNeeded()
	return r.err
}

// Response возвращает объект HTTP-ответа
func (r *Request) Response() *http.Response {
	r.executeIfNeeded()
	return r.resp
}

// Headers возвращает заголовки ответа
func (r *Request) Headers() http.Header {
	r.executeIfNeeded()
	if r.resp == nil {
		return nil
	}
	return r.resp.Header
}

// GetHeader возвращает значение конкретного заголовка ответа
func (r *Request) GetHeader(k string) string {
	r.executeIfNeeded()
	if r.resp == nil {
		return ""
	}
	return r.resp.Header.Get(k)
}

// Cookies возвращает cookies из ответа
func (r *Request) Cookies() []*http.Cookie {
	r.executeIfNeeded()
	if r.resp == nil {
		return nil
	}
	return r.resp.Cookies()
}

// GetCookie возвращает cookie по имени
func (r *Request) GetCookie(name string) *http.Cookie {
	for _, cookie := range r.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// Close закрывает контекст и тело ответа
func (r *Request) Close() {
	if r.cancel != nil && !r.cancelled.Swap(true) {
		r.cancel()
	}
	if r.resp != nil && r.resp.Body != nil {
		_ = r.resp.Body.Close()
	}
}
