class GoTradingClient {
	constructor(baseUrl = '') {
		// Используем текущий URL страницы как базовый
		this.baseUrl = baseUrl || window.location.origin;
	}

	async #makeRequest(endpoint, params = {}, method = 'GET', body = null) {
		const url = new URL(`${this.baseUrl}${endpoint}`);
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined) {
				url.searchParams.append(key, value);
			}
		});

		const options = {
			method,
			headers: {
				'Content-Type': 'application/json',
			},
		};

		if (body) {
			options.body = JSON.stringify(body);
		}

		const response = await fetch(url, options);
		// if (!response.ok) {
		// 	const errorData = await response.json().catch(() => ({}));
		// 	throw new Error(errorData.error || `Request failed with status ${response.status}`);
		// }

		return response.json();
	}

	// Basic endpoints
	ping() {
		return this.#makeRequest('/ping');
	}

	// Crypto endpoints
	searchCrypto(query) {
		if (!query) {
			return Promise.reject(new Error('Search query is required'));
		}
		return this.#makeRequest('/api/v1/crypto', { s: query });
	}

	getCryptoDetail(id) {
		if (!id) {
			return Promise.reject(new Error('ID is required'));
		}
		return this.#makeRequest('/api/v1/crypto/detail', { id });
	}

	getCryptoLiteDetail(id) {
		if (!id) {
			return Promise.reject(new Error('ID is required'));
		}
		return this.#makeRequest('/api/v1/crypto/detail/lite', { id });
	}

	getCryptoFearAndGreed() {
		return this.#makeRequest('/api/v1/crypto/fearAndGreed');
	}

	getCryptoList(limit) {
		if (!limit) {
			return Promise.reject(new Error('Limit is required'));
		}
		return this.#makeRequest('/api/v1/crypto/list', { l: limit });
	}

	getCryptoImage(symbol) {
		return `${this.baseUrl}/static/img/crypto/${symbol}.png`;
	}

	// Candles endpoints
	getCandles(symbol, interval, limit = 999) {
		if (!symbol || !interval) {
			return Promise.reject(new Error('Symbol and interval are required'));
		}
		return this.#makeRequest('/api/v1/candles', { s: symbol, i: interval, l: limit });
	}

	getCurrentCandle(symbol, interval) {
		if (!symbol || !interval) {
			return Promise.reject(new Error('Symbol and interval are required'));
		}
		return this.#makeRequest('/api/v1/candle', { s: symbol, i: interval });
	}

	getOrderLog(periodSec) {
		if (!periodSec) {
			return Promise.reject(new Error('periodSec is required'));
		}
		return this.#makeRequest('/api/v1/order-log', { p: periodSec });
	}

	// Prediction endpoint
	predictTrend(candles, markings = []) {
		if (!candles || !Array.isArray(candles)) {
			return Promise.reject(new Error('Candles data is required and must be an array'));
		}
		return this.#makeRequest(
			'/api/v1/predict/trend',
			{ m: markings.join(',') },
			'POST',
			candles
		);
	}
}

window.GoTradingClient = GoTradingClient;
