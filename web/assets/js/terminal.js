class CryptoViewer {
    constructor() {
        this.client = new GoTradingClient();
        InitChartManager(this.client);
        this.symbolData = null;
        this.intervals = {
            cryptoList: null,
            fearAndGreed: null,
            symbolDetail: null
        };

        this.initElements();
        this.initEventListeners();
    }

    initElements() {
        this.symbolInput = document.getElementById('symbol-input');
        this.symbolInfo = document.getElementById('symbol-info');
        this.cryptoList = document.getElementById('crypto-list');
        this.fearAndGreedBlock = document.getElementById('fear-and-greed');
    }

    initEventListeners() {
        this.symbolInput.addEventListener('change', this.handleSymbolChange.bind(this));
        window.addEventListener('beforeunload', this.cleanup.bind(this));
    }

    showError(errorMessage) {
        alert(`Ошибка: ${errorMessage}`);
    }

    updateSymbolInfo() {
        if (!this.symbolData) return;

        this.symbolInfo.classList.remove('hidden');

        const elements = {
            'symbol-icon': { attr: 'src', value: `./static/img/crypto/${this.symbolData.id}` },
            'symbol-name': { text: this.symbolData.name },
            'symbol-ticker': { text: this.symbolData.symbol },
            'symbol-rank': { text: `#${this.symbolData.statistics.rank}` }
        };

        Object.entries(elements).forEach(([id, { attr, text, value }]) => {
            const element = document.getElementById(id);
            if (!element) return;

            if (attr) element[attr] = value;
            if (text) element.textContent = text;
        });

        this.updateChangeElements();
    }

    updateChangeElements() {
        const changes = {
            'change-1h': this.symbolData.statistics.priceChangePercentage1h,
            'change-24h': this.symbolData.statistics.priceChangePercentage24h,
            'change-7d': this.symbolData.statistics.priceChangePercentage7d,
            'change-yesterday': this.symbolData.statistics.priceChangePercentageYesterday
        };

        Object.entries(changes).forEach(([id, value]) => {
            this.updateChangeElement(id, value);
        });
    }

    updateChangeElement(elementId, value) {
        const element = document.getElementById(elementId);
        if (!element) return;

        element.textContent = `${value > 0 ? '+' : ''}${value.toFixed(2)}%`;

        const className = value > 0 ? 'positive' :
            value < 0 ? 'negative' : 'neutral';

        element.className = element.className.split(' ')
            .filter(c => !['positive', 'negative', 'neutral'].includes(c))
            .concat(className)
            .join(' ');
    }

    updateCryptoList(cryptoData) {
        const itemsContainer = this.cryptoList.querySelector('.crypto-list-items');
        itemsContainer.innerHTML = '';

        cryptoData.forEach((crypto, index) => {
            const item = document.createElement('div');
            item.className = 'crypto-item';
            item.innerHTML = this.createCryptoItemHTML(crypto, index);
            item.addEventListener('click', () => this.selectCrypto(crypto.symbol));
            itemsContainer.appendChild(item);
        });

        this.cryptoList.classList.remove('hidden');
    }

    createCryptoItemHTML(crypto, index) {
        return `
            <span class="crypto-index">${index + 1}</span>
            <span class="crypto-name">${crypto.name}</span>
            <span class="crypto-symbol">${crypto.symbol}</span>
            <span class="crypto-change ${this.getChangeClass(crypto.percentChange24h)}">
                ${this.formatChangeValue(crypto.percentChange24h)}
            </span>
            <span class="crypto-change ${this.getChangeClass(crypto.percentChange7d)}">
                ${this.formatChangeValue(crypto.percentChange7d)}
            </span>
        `;
    }

    selectCrypto(symbol) {
        this.symbolInput.value = symbol;
        this.symbolInput.dispatchEvent(new Event('change'));
    }

    formatChangeValue(value) {
        return `${value > 0 ? '+' : ''}${value.toFixed(2)}%`;
    }

    getChangeClass(value) {
        if (value > 0) return 'positive';
        if (value < 0) return 'negative';
        return 'neutral';
    }

    async fetchCryptoList() {
        try {
            const data = await this.client.getCryptoList(100);
            if (data.error) return this.showError(data.error);
            this.updateCryptoList(data.result);
        } catch (error) {
            this.showError(error.message);
        }
    }

    async fetchSymbolDetail(symbol, mod = 0) {
        if (!symbol) return;

        try {
            const searchData = await this.client.searchCrypto(symbol);
            if (searchData.error) return this.showError(searchData.error);

            RenderChart(symbol).catch(() => { });

            if (mod === 1 && Math.random() > 0.1) return;

            const detailData = await this.client.getCryptoDetail(searchData.result.id);
            if (detailData.error) return this.showError(detailData.error);

            this.symbolData = detailData.result;
            this.updateSymbolInfo();
        } catch (error) {
            this.showError(error.message);
        }
    }

    updateFearAndGreed(data) {
        if (!this.fearAndGreedBlock) return;

        const { nowScore: score, percentChange24h: change24h } = data;
        const indicator = this.fearAndGreedBlock.querySelector('.fg-indicator');
        const scoreElement = this.fearAndGreedBlock.querySelector('.fg-score');
        const changeElement = this.fearAndGreedBlock.querySelector('.fg-change-value');

        indicator.style.left = `${score}%`;
        scoreElement.textContent = score;

        changeElement.textContent = this.formatChangeValue(change24h);
        changeElement.className = 'fg-change-value ' + this.getChangeClass(change24h);

        this.fearAndGreedBlock.classList.remove('hidden');
    }

    async fetchFearAndGreed() {
        try {
            const data = await this.client.getCryptoFearAndGreed();
            if (data.error) return this.showError(data.error);
            this.updateFearAndGreed(data.result);
        } catch (error) {
            this.showError(error.message);
        }
    }

    handleSymbolChange() {
        const symbol = this.symbolInput.value.trim().toUpperCase();

        if (this.intervals.symbolDetail) {
            clearInterval(this.intervals.symbolDetail);
            this.intervals.symbolDetail = null;
        }

        if (symbol) {
            this.fetchSymbolDetail(symbol);
            this.intervals.symbolDetail = setInterval(() => {
                this.fetchSymbolDetail(symbol, 1);
            }, 4000);
        }
    }

    init() {
        this.fetchCryptoList();
        this.fetchFearAndGreed();

        this.intervals.cryptoList = setInterval(
            () => this.fetchCryptoList(),
            100000
        );

        this.intervals.fearAndGreed = setInterval(
            () => this.fetchFearAndGreed(),
            200000
        );
    }

    cleanup() {
        Object.values(this.intervals).forEach(interval => {
            if (interval) clearInterval(interval);
        });
        this.symbolInput.removeEventListener('change', this.handleSymbolChange);
    }
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    const cryptoViewer = new CryptoViewer();
    cryptoViewer.init();
});
