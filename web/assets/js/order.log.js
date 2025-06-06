class OrderLogManager {
	constructor() {
		this.client = new GoTradingClient();
		this.tableBody = document.querySelector('#order-log-tb tbody');
		this.cryptoCache = {};
		this.orderLogData = {};
		this.currentSort = { key: null, asc: true };

		this.init();
	}

	init() {
		this.bindEvents();
		this.loadData();
	}

	bindEvents() {
		document.querySelectorAll('th[data-sort-key]').forEach(th => {
			th.addEventListener('click', () => this.handleSortClick(th));
		});
	}

	async loadData() {
		try {
			const { result: orders } = await this.client.getOrderLog(10000);
			this.renderOrders(orders);
		} catch (error) {
			this.showError(error.message);
			console.error('Error loading order data:', error);
		}
	}

	showError(error) {
		this.tableBody.innerHTML = `
      <tr>
        <td colspan="11" class="error-message">
          Error loading data: ${error}
        </td>
      </tr>
    `;
	}

	formatTimestamp(timestamp) {
		if (!timestamp) return '-';

		const d = new Date(timestamp);
		const pad = n => n.toString().padStart(2, '0');

		return `${pad(d.getMonth() + 1)}/${pad(d.getDate())}/${d.getFullYear().toString().slice(-2)}, ` +
			`${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
	}

	formatCryptoHTML(crypto) {
		const currencyPath = crypto.name.toLowerCase().replace(/\s+/g, '-');
		const coinmarketcapUrl = `https://coinmarketcap.com/currencies/${currencyPath}`;

		return `
        <a href="${coinmarketcapUrl}"
           target="_blank"
           rel="noopener noreferrer"
           style="text-decoration: none; color: inherit;">
            <img class="symbol-icon" src="/static/img/crypto/${crypto.id}" alt="${crypto.symbol} icon">
            <span class="symbol-lname">${crypto.name}</span>
            <span class="symbol-sname">${crypto.symbol}</span>
        </a>
    `;
	}

	async getCryptoData(symbol) {
		if (this.cryptoCache[symbol]) {
			return this.cryptoCache[symbol];
		}

		const coin = symbol.replace('USDT', '');
		const response = await this.client.searchCrypto(coin);

		const crypto = response.error ?
			{ id: 1, symbol: coin, name: '-' } :
			response.result;

		this.cryptoCache[symbol] = crypto;
		return crypto;
	}

	createCell(content, className = '') {
		const cell = document.createElement('td');
		if (className) cell.className = className;

		if (typeof content === 'string' || typeof content === 'number') {
			cell.textContent = content;
		} else {
			cell.appendChild(content);
		}

		return cell;
	}

	async renderOrders(orders) {
		this.tableBody.innerHTML = '';
		this.orderLogData = {};

		let rowNumber = orders.length;

		for (const order of orders) {
			rowNumber--;
			const row = await this.createOrderRow(order, rowNumber); // Добавляем await
			this.tableBody.appendChild(row);
		}
	}

	async createOrderRow(order, rowNumber) {
		const { id, symbol, qty, avgPrice, execQty, execValue, fee, createdAt, isClosed } = order.order;
		const isBuy = qty > 0;
		const isOrderEmpty = id === '';

		// Store order data for sorting
		this.orderLogData[order.linkId] = {
			rowNumber,
			side: isBuy ? 1 : 0,
			qty: Math.abs(qty),
			avgPrice,
			execQty: Math.abs(execQty),
			execValue: Math.abs(execValue),
			fee,
			createdAt,
			status: isClosed ? 1 : 0,
		};

		const row = document.createElement('tr');
		row.id = order.linkId;

		// Добавляем await для создания ячейки с символом
		const symbolCell = await this.createSymbolCell(symbol);

		// Add cells to row
		row.append(
			this.createCell(rowNumber),
			symbolCell, // Используем уже созданную ячейку
			this.createCell(order.tag || '-'),
			this.createSideCell(isBuy),
			this.createCell(qty || '-', 'order-qty'),
			this.createCell(avgPrice ?? '-', 'avg-price'),
			this.createCell(isOrderEmpty ? '-' : (execQty || '-'), 'exec-qty'),
			this.createCell(isOrderEmpty ? '-' : (execValue || '-'), 'exec-value'),
			this.createCell(fee || '-', 'order-fee'),
			this.createCell(this.formatTimestamp(createdAt)),
			this.createStatusCell(id, isClosed)
		);

		return row;
	}

	async createSymbolCell(symbol) {
		const crypto = await this.getCryptoData(symbol);
		const wrapper = document.createElement('div');
		wrapper.className = 'symbol-name-wrapper';
		wrapper.innerHTML = this.formatCryptoHTML(crypto);
		return this.createCell(wrapper);
	}

	createSideCell(isBuy) {
		const span = document.createElement('span');
		span.className = isBuy ? 'order-buy' : 'order-sell';
		span.textContent = isBuy ? 'Buy' : 'Sell';
		return this.createCell(span);
	}

	createStatusCell(id, isClosed) {
		const span = document.createElement('span');

		if (id === '') {
			span.className = 'order-status-rejected';
			span.textContent = 'rejected';
		} else {
			span.className = isClosed ? 'order-status-closed' : 'order-status-placed';
			span.textContent = isClosed ? 'closed' : 'placed';
		}

		return this.createCell(span);
	}

	handleSortClick(th) {
		const key = th.getAttribute('data-sort-key');

		if (this.currentSort.key === key) {
			this.currentSort.asc = !this.currentSort.asc;
		} else {
			this.currentSort.key = key;
			this.currentSort.asc = true;
		}

		this.sortAndRender();
	}

	sortAndRender() {
		const { key, asc } = this.currentSort;

		const sortedIds = Object.entries(this.orderLogData)
			.sort(([, a], [, b]) => asc ? a[key] - b[key] : b[key] - a[key])
			.map(([id]) => id);

		sortedIds.forEach(id => {
			const row = document.getElementById(id);
			if (row) this.tableBody.appendChild(row);
		});
	}
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => new OrderLogManager());
