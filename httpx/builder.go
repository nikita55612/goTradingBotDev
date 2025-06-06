package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

// RequestBuilder интерфейс для построения HTTP-запросов с помощью паттерна "строитель"
type RequestBuilder interface {
	SetURL(url string) RequestBuilder
	WithContext(ctx context.Context) RequestBuilder
	WithTimeout(d time.Duration) RequestBuilder
	WithBody(r io.Reader) RequestBuilder
	WithJsonData(d any) RequestBuilder
	AddHeader(k, v string) RequestBuilder
	SetHeaderValue(k, v string) RequestBuilder
	DelHeaderKey(k string) RequestBuilder
	SetQueryParam(k, v string) RequestBuilder
	DelQueryParam(k string) RequestBuilder
	SetHeader(header http.Header) RequestBuilder
	SetQueryParams(params map[string]string) RequestBuilder
	SetClient(c *http.Client) RequestBuilder
	AddCookie(c *http.Cookie) RequestBuilder
	SetCookies(cookies []*http.Cookie) RequestBuilder
	WithProxy(proxyURL string) RequestBuilder
	Do() *Request
}

// NewRequestBuilder создает новый строитель запросов
// Иммутабельность строителя - каждый метод возвращает новый объект строителя
func NewRequestBuilder(method, url string) RequestBuilder {
	return &requestBuilder{
		method: method,
		url:    url,
	}
}

// requestBuilder реализация интерфейса RequestBuilder
type requestBuilder struct {
	method      string
	url         string
	ctx         context.Context
	timeout     time.Duration
	body        io.Reader
	jsonData    any
	header      http.Header
	queryParams map[string]string
	client      *http.Client
	cookies     []*http.Cookie
	proxyURL    string
}

// copyQueryParams создает копию query параметров
func (rb *requestBuilder) copyQueryParams() map[string]string {
	if rb.queryParams == nil {
		return nil
	}
	params := make(map[string]string, len(rb.queryParams))
	for k, v := range rb.queryParams {
		params[k] = v
	}
	return params
}

// WithTimeout устанавливает таймаут для запроса
func (rb *requestBuilder) SetURL(url string) RequestBuilder {
	newBuilder := *rb
	newBuilder.url = url
	return &newBuilder
}

// WithTimeout устанавливает таймаут для запроса
func (rb *requestBuilder) WithTimeout(d time.Duration) RequestBuilder {
	newBuilder := *rb
	newBuilder.timeout = d
	return &newBuilder
}

// WithContext добавляет контекст к запросу
func (rb *requestBuilder) WithContext(ctx context.Context) RequestBuilder {
	newBuilder := *rb
	newBuilder.ctx = ctx
	return &newBuilder
}

// WithBody добавляет тело запроса
func (rb *requestBuilder) WithBody(r io.Reader) RequestBuilder {
	newBuilder := *rb
	newBuilder.body = r
	return &newBuilder
}

// WithJsonData добавляет тело запроса
func (rb *requestBuilder) WithJsonData(data any) RequestBuilder {
	newBuilder := *rb
	newBuilder.jsonData = data
	return &newBuilder
}

// AddHeader добавляет заголовок к запросу
func (rb *requestBuilder) AddHeader(k, v string) RequestBuilder {
	newBuilder := *rb
	if newBuilder.header == nil {
		newBuilder.header = make(http.Header)
	} else {
		newBuilder.header = rb.header.Clone()
	}
	newBuilder.header.Add(k, v)
	return &newBuilder
}

// SetHeader устанавливает значение заголовока к запросу
func (rb *requestBuilder) SetHeaderValue(k, v string) RequestBuilder {
	newBuilder := *rb
	if newBuilder.header == nil {
		newBuilder.header = make(http.Header)
	} else {
		newBuilder.header = rb.header.Clone()
	}
	newBuilder.header.Set(k, v)
	return &newBuilder
}

// DelHeaderKey удаляет заголовок запроса по ключу
func (rb *requestBuilder) DelHeaderKey(k string) RequestBuilder {
	newBuilder := *rb
	if newBuilder.header == nil {
		newBuilder.header = make(http.Header)
	} else {
		newBuilder.header = rb.header.Clone()
	}
	newBuilder.header.Del(k)
	return &newBuilder
}

// SetQueryParam назначает query-параметр к URL
func (rb *requestBuilder) SetQueryParam(k, v string) RequestBuilder {
	newBuilder := *rb
	if newBuilder.queryParams == nil {
		newBuilder.queryParams = make(map[string]string)
	} else {
		newBuilder.queryParams = rb.copyQueryParams()
	}
	newBuilder.queryParams[k] = v
	return &newBuilder
}

// DelQueryParam удаляет query-параметр URL
func (rb *requestBuilder) DelQueryParam(k string) RequestBuilder {
	newBuilder := *rb
	if newBuilder.queryParams == nil {
		newBuilder.queryParams = make(map[string]string)
	} else {
		newBuilder.queryParams = rb.copyQueryParams()
	}
	delete(newBuilder.queryParams, k)
	return &newBuilder
}

// SetHeader назначает заголовок запроса
func (rb *requestBuilder) SetHeader(header http.Header) RequestBuilder {
	newBuilder := *rb
	newBuilder.header = header.Clone()
	return &newBuilder
}

// SetQueryParams назначает query-параметры к URL
func (rb *requestBuilder) SetQueryParams(params map[string]string) RequestBuilder {
	newBuilder := *rb
	newBuilder.queryParams = make(map[string]string, len(params))
	for k, v := range params {
		newBuilder.queryParams[k] = v
	}
	return &newBuilder
}

// SetClient устанавливает кастомный HTTP-клиент для выполнения запроса
func (rb *requestBuilder) SetClient(c *http.Client) RequestBuilder {
	newBuilder := *rb
	newBuilder.client = c
	return &newBuilder
}

// AddCookie добавляет cookie к запросу
func (rb *requestBuilder) AddCookie(c *http.Cookie) RequestBuilder {
	newBuilder := *rb
	newBuilder.cookies = append([]*http.Cookie{}, rb.cookies...)
	newBuilder.cookies = append(newBuilder.cookies, c)
	return &newBuilder
}

// SetCookies устанавливает cookies для запроса
func (rb *requestBuilder) SetCookies(cookies []*http.Cookie) RequestBuilder {
	newBuilder := *rb
	newBuilder.cookies = make([]*http.Cookie, len(cookies))
	copy(newBuilder.cookies, cookies)
	return &newBuilder
}

// WithProxy прокси для запроса
func (rb *requestBuilder) WithProxy(proxyURL string) RequestBuilder {
	newBuilder := *rb
	newBuilder.proxyURL = proxyURL
	return &newBuilder
}

// Do выполняет построение запроса и возвращает объект Request
// Ошибки при строительсве запроса, передаются в объект Request
func (rb *requestBuilder) Do() *Request {
	var req *http.Request
	var err error

	u, err := url.Parse(rb.url)
	if err != nil {
		return &Request{err: err}
	}
	if len(rb.queryParams) > 0 {
		query := u.Query()
		for k, v := range rb.queryParams {
			query.Set(k, v)
		}
		u.RawQuery = query.Encode()
	}
	client := rb.client
	if client == nil {
		if rb.proxyURL != "" {
			proxyURL, err := url.Parse(rb.proxyURL)
			if err != nil {
				return &Request{err: err}
			}
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
			client = &http.Client{
				Transport: transport,
			}
		} else {
			client = http.DefaultClient
		}
	} else if rb.proxyURL != "" {
		proxyURL, err := url.Parse(rb.proxyURL)
		if err != nil {
			return &Request{err: err}
		}
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport = transport.Clone()
			transport.Proxy = http.ProxyURL(proxyURL)
			client.Transport = transport
		} else {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}
	body := rb.body
	if rb.jsonData != nil {
		jsonData, err := json.Marshal(rb.jsonData)
		if err != nil {
			return &Request{err: err}
		}
		body = bytes.NewBuffer(jsonData)
	}
	ctx := rb.ctx
	var cancel context.CancelFunc
	if rb.timeout > 0 {
		if ctx == nil {
			ctx = context.Background()
		}
		ctx, cancel = context.WithTimeout(ctx, rb.timeout)
	}
	if ctx != nil {
		req, err = http.NewRequestWithContext(ctx, rb.method, u.String(), body)
	} else {
		req, err = http.NewRequest(rb.method, u.String(), body)
	}
	if err != nil {
		cancel()
		return &Request{err: err}
	}
	if rb.header != nil {
		req.Header = rb.header.Clone()
	}
	for _, cookie := range rb.cookies {
		req.AddCookie(cookie)
	}
	return &Request{
		req:    req,
		err:    err,
		client: client,
		cancel: cancel,
	}
}

// Методы для создания строителей различных HTTP-методов

// Get создает строитель для GET-запроса
func Get(url string) RequestBuilder {
	return NewRequestBuilder(http.MethodGet, url)
}

// Post создает строитель для POST-запроса
func Post(url string) RequestBuilder {
	return NewRequestBuilder(http.MethodPost, url)
}

// Put создает строитель для PUT-запроса
func Put(url string) RequestBuilder {
	return NewRequestBuilder(http.MethodPut, url)
}

// Delete создает строитель для DELETE-запроса
func Delete(url string) RequestBuilder {
	return NewRequestBuilder(http.MethodDelete, url)
}

// Patch создает строитель для PATCH-запроса
func Patch(url string) RequestBuilder {
	return NewRequestBuilder(http.MethodPatch, url)
}

// Head создает строитель для HEAD-запроса
func Head(url string) RequestBuilder {
	return NewRequestBuilder(http.MethodHead, url)
}

// Options создает строитель для OPTIONS-запроса
func Options(url string) RequestBuilder {
	return NewRequestBuilder(http.MethodOptions, url)
}
