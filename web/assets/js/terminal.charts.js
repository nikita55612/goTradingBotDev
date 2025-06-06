class ChartManager {
    constructor(client) {
        this.client = client;
        this.symbolChart = null;
        this.signalsChart = null;
        this.currentOhlcData = [];
        this.allOhlcData = {
            '1h': [],
            '15m': []
        };
        this.allSignalsData = {
            '1h': { pt4: [], npt9: [] },
            '15m': { pt4: [], npt9: [] }
        };
        this.lastSignals = {
            '1h': { pt4: null, npt9: null },
            '15m': { pt4: null, npt9: null }
        };
        this.currentInterval = '15m';
        this.currentSymbol = null;
        this.initIntervalSwitcher();
    }

    initIntervalSwitcher() {
        document.querySelectorAll('.interval-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const interval = btn.dataset.interval;
                this.switchInterval(interval);
            });
        });
    }

    switchInterval(interval) {
        if (interval === this.currentInterval) return;

        this.currentInterval = interval;
        document.querySelectorAll('.interval-btn').forEach(b => {
            b.classList.toggle('active', b.dataset.interval === interval);
        });
        this.currentOhlcData = this.allOhlcData[this.currentInterval] || [];
        this.initSymbolChart();
        this.updateSymbolChart();
        this.updateSignalsChart(true);
    }

    updateSignalsInfo() {
        // Получаем текущие и предыдущие значения для всех сигналов
        const h1pt4Current = this.allSignalsData['1h'].pt4.at(-1);
        const h1pt4Previous = this.allSignalsData['1h'].pt4.at(-2);

        const h1npt9Current = this.allSignalsData['1h'].npt9.at(-1);
        const h1npt9Previous = this.allSignalsData['1h'].npt9.at(-2);

        const m15pt4Current = this.allSignalsData['15m'].pt4.at(-1);
        const m15pt4Previous = this.allSignalsData['15m'].pt4.at(-2);

        const m15npt9Current = this.allSignalsData['15m'].npt9.at(-1);
        const m15npt9Previous = this.allSignalsData['15m'].npt9.at(-2);

        // Обновляем элементы для 1h-pt4
        this.updateSignalPair(
            document.getElementById('pv-1h-pt4'), // previous value
            document.getElementById('1h-pt4'),    // current value
            h1pt4Previous,
            h1pt4Current
        );

        // Обновляем элементы для 1h-npt9
        this.updateSignalPair(
            document.getElementById('pv-1h-npt9'),
            document.getElementById('1h-npt9'),
            h1npt9Previous,
            h1npt9Current
        );

        // Обновляем элементы для 15m-pt4
        this.updateSignalPair(
            document.getElementById('pv-15m-pt4'),
            document.getElementById('15m-pt4'),
            m15pt4Previous,
            m15pt4Current
        );

        // Обновляем элементы для 15m-npt9
        this.updateSignalPair(
            document.getElementById('pv-15m-npt9'),
            document.getElementById('15m-npt9'),
            m15npt9Previous,
            m15npt9Current
        );
    }

    updateSignalPair(prevElement, currElement, prevValue, currValue) {
        // Обновляем предыдущее значение
        if (prevValue !== undefined && prevValue !== null) {
            prevElement.textContent = prevValue.toFixed(2);
            const prevClass = prevValue < 0.5 ? 'negative' : 'positive';
            prevElement.className = prevElement.className.split(' ')
                .filter(c => !['positive', 'negative', 'neutral'].includes(c))
                .concat(prevClass)
                .join(' ');
        }

        // Обновляем текущее значение
        if (currValue !== undefined && currValue !== null) {
            currElement.textContent = currValue.toFixed(2);
            const currClass = currValue < 0.5 ? 'negative' : 'positive';
            currElement.className = currElement.className.split(' ')
                .filter(c => !['positive', 'negative', 'neutral'].includes(c))
                .concat(currClass)
                .join(' ');
        }
    }

    parseCandle(candleData) {
        if (!candleData) return null;

        if (Array.isArray(candleData[0])) {
            return candleData.map(candle => this.createCandleObject(candle));
        } else if (Array.isArray(candleData)) {
            return this.createCandleObject(candleData);
        }
        throw new Error("Invalid candle data format");
    }

    createCandleObject(candleArray) {
        if (!candleArray || candleArray.length < 7) return null;

        return {
            t: parseInt(candleArray[0]),
            o: parseFloat(candleArray[1]),
            h: parseFloat(candleArray[2]),
            l: parseFloat(candleArray[3]),
            c: parseFloat(candleArray[4]),
            v: parseFloat(candleArray[5]),
            to: parseFloat(candleArray[6])
        };
    }

    stringifyCandle(candleData) {
        if (!candleData) return null;

        if (Array.isArray(candleData)) {
            return candleData.map(candle => this.candleToArray(candle));
        } else if (typeof candleData === 'object') {
            return this.candleToArray(candleData);
        }
        throw new Error("Invalid candle object format");
    }

    candleToArray(candle) {
        if (!candle) return null;

        return [
            candle.t.toString(),
            candle.o.toString(),
            candle.h.toString(),
            candle.l.toString(),
            candle.c.toString(),
            candle.v.toString(),
            candle.to.toString()
        ];
    }

    initSignalsChart() {
        const canvas = document.getElementById('signals-chart');
        if (!canvas) return;

        if (this.signalsChart) {
            this.signalsChart.destroy();
        }

        const config = {
            type: 'line',
            options: {
                animation: false,
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        min: 0, max: 1,
                        ticks: { stepSize: 0.1 },
                        grid: {
                            color: ctx => ctx.tick.value === 0.5 ? '#f43841' : '#e0e0e020',
                            lineWidth: ctx => ctx.tick.value === 0.5 ? 2 : 1
                        }
                    },
                    x: { grid: { display: false } }
                },
                plugins: {
                    legend: { position: 'top' },
                    tooltip: {
                        callbacks: {
                            label: ctx => `${ctx.dataset.label}: ${ctx.raw?.toFixed(2) || 'N/A'}`
                        }
                    }
                }
            }
        };

        this.signalsChart = new Chart(canvas, config);
    }

    initSymbolChart() {
        const canvas = document.getElementById('symbol-chart');
        if (!canvas) return;

        if (this.symbolChart) {
            this.symbolChart.destroy();
        }

        const config = {
            type: 'line',
            options: this.getSymbolChartOptions(),
            plugins: [{
                id: 'candlestick',
                beforeDraw: chart => this.drawCandlesticks(chart)
            }]
        };

        this.symbolChart = new Chart(canvas, config);
    }

    getSymbolChartOptions() {
        return {
            animation: false,
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { display: false },
                tooltip: {
                    callbacks: {
                        label: ctx => {
                            const d = this.currentOhlcData[ctx.dataIndex];
                            if (!d) return [];
                            return [
                                `Open: ${d.o?.toFixed(2) || 'N/A'}`,
                                `High: ${d.h?.toFixed(2) || 'N/A'}`,
                                `Low: ${d.l?.toFixed(2) || 'N/A'}`,
                                `Close: ${d.c?.toFixed(2) || 'N/A'}`,
                                `Change: ${d.c && d.o ? ((d.c - d.o) / d.o * 100).toFixed(2) : 'N/A'}%`
                            ];
                        }
                    }
                }
            },
            scales: {
                y: { offset: true, grid: { display: false } },
                x: { offset: true, bounds: 'data', grid: { display: false } }
            }
        };
    }

    drawCandlesticks(chart) {
        if (!this.currentOhlcData?.length) return;

        const { ctx, scales: { x, y } } = chart;
        const candleWidth = (x.width / this.currentOhlcData.length) * 0.7;

        ctx.save();
        ctx.lineWidth = 1.5;

        this.currentOhlcData.forEach((d, i) => {
            if (!d) return;

            const xPos = x.getPixelForValue(i);
            const isBullish = d.c >= d.o;
            const color = isBullish ? '#73c936' : '#f43841';

            // Draw wick
            ctx.beginPath();
            ctx.moveTo(xPos, y.getPixelForValue(d.h));
            ctx.lineTo(xPos, y.getPixelForValue(d.l));
            ctx.strokeStyle = color;
            ctx.stroke();

            // Draw body
            const openY = y.getPixelForValue(d.o);
            const closeY = y.getPixelForValue(d.c);
            const height = Math.max(1, Math.abs(openY - closeY));

            ctx.fillStyle = color;
            ctx.fillRect(
                xPos - candleWidth / 2,
                Math.min(openY, closeY),
                candleWidth,
                height
            );

            // Draw body border
            ctx.strokeStyle = color;
            ctx.strokeRect(
                xPos - candleWidth / 2,
                Math.min(openY, closeY),
                candleWidth,
                height
            );
        });

        ctx.restore();
    }

    async loadDataForAllIntervals(symbol) {
        try {
            if (!symbol) return false;

            if (this.currentSymbol && this.currentSymbol === symbol) {
                return await this.updateExistingData(symbol);
            } else {
                return await this.loadNewData(symbol);
            }
        } catch (error) {
            console.error("Error loading data:", error);
            return false;
        }
    }

    async updateExistingData(symbol) {
        try {
            const [h1Candles, min15Candles, currH1Candle, currM15Candle] = await Promise.all([
                this.client.getCandles(symbol, 60, 2),
                this.client.getCandles(symbol, 15, 2),
                this.client.getCurrentCandle(symbol, 60),
                this.client.getCurrentCandle(symbol, 15)
            ]);

            const h1CandlesRes = this.parseCandle(h1Candles.result);
            const m15CandlesRes = this.parseCandle(min15Candles.result);
            const currH1CandleRes = this.parseCandle(currH1Candle.result);
            const currM15CandleRes = this.parseCandle(currM15Candle.result);

            if (!h1CandlesRes || !m15CandlesRes || !currH1CandleRes || !currM15CandleRes) {
                throw new Error("Failed to parse candle data");
            }

            await this.processIntervalUpdate('1h', h1CandlesRes, currH1CandleRes);
            await this.processIntervalUpdate('15m', m15CandlesRes, currM15CandleRes);

            return true;
        } catch (error) {
            console.error("Error updating existing data:", error);
            return false;
        }
    }

    async processIntervalUpdate(interval, historyData, currentData) {
        if (!this.allOhlcData[interval]?.length || !historyData?.length || !currentData) return;

        const lastCandle = this.allOhlcData[interval].at(-2);

        if (lastCandle?.t !== historyData.at(-1)?.t) {
            this.allOhlcData[interval].splice(-1, 1, historyData.at(-1));
            await this.updateSignalsData(interval);
            this.allOhlcData[interval].push(currentData);
            this.allOhlcData[interval].shift();
        } else {
            this.allOhlcData[interval].splice(-1, 1, currentData);
        }
    }

    async updateSignalsData(interval) {
        try {
            const intervalKey = interval === '1h' ? 'H1' : 'M15';
            const predict = await this.client.predictTrend(
                this.stringifyCandle(this.allOhlcData[interval]),
                [`xgb_linear-${intervalKey}_`,]
            );

            if (!predict?.result) throw new Error("No prediction result");

            this.allSignalsData[interval].pt4 =
                predict.result[`xgb_linear-${intervalKey}_PerfectTrend-p4`] || [];
            this.allSignalsData[interval].npt9 =
                predict.result[`xgb_linear-${intervalKey}_NextPerfectTrend-p9`] || [];
        } catch (error) {
            console.error(`Error updating signals for ${interval}:`, error);
        }
    }

    async loadNewData(symbol) {
        try {
            const [h1Candles, m15Candles, currH1Candle, currM15Candle] = await Promise.all([
                this.client.getCandles(symbol, 60, 100),
                this.client.getCandles(symbol, 15, 100),
                this.client.getCurrentCandle(symbol, 60),
                this.client.getCurrentCandle(symbol, 15)
            ]);

            this.allOhlcData['1h'] = this.parseCandle(h1Candles.result) || [];
            this.allOhlcData['15m'] = this.parseCandle(m15Candles.result) || [];

            const currH1 = this.parseCandle(currH1Candle.result);
            const currM15 = this.parseCandle(currM15Candle.result);

            if (currH1) this.allOhlcData['1h'].push(currH1);
            if (currM15) this.allOhlcData['15m'].push(currM15);

            this.currentSymbol = symbol;

            const [predict1h, predict15m] = await Promise.all([
                this.client.predictTrend(h1Candles.result, ['linear-', 'H1']),
                this.client.predictTrend(m15Candles.result, ['linear-', 'M15'])
            ]);

            if (predict1h?.result) {
                this.allSignalsData['1h'].pt4 = predict1h.result['xgb_linear-H1_PerfectTrend-p4'] || [];
                this.allSignalsData['1h'].npt9 = predict1h.result['xgb_linear-H1_NextPerfectTrend-p9'] || [];
            }

            if (predict15m?.result) {
                this.allSignalsData['15m'].pt4 = predict15m.result['xgb_linear-M15_PerfectTrend-p4'] || [];
                this.allSignalsData['15m'].npt9 = predict15m.result['xgb_linear-M15_NextPerfectTrend-p9'] || [];
            }

            return true;
        } catch (error) {
            console.error("Error loading new data:", error);
            return false;
        }
    }

    padZerosToStart(originalArray, targetLength = 100) {
        if (!originalArray) return Array(targetLength).fill(0);
        if (originalArray.length >= targetLength) return [...originalArray];
        return [...Array(targetLength - originalArray.length).fill(0), ...originalArray];
    }

    async renderChart(symbol) {
        try {
            if (!symbol) return;

            const success = await this.loadDataForAllIntervals(symbol);
            if (!success) return;

            this.currentOhlcData = this.allOhlcData[this.currentInterval] || [];

            this.initSymbolChart();
            this.updateSymbolChart();

            await this.updateSignalsChart();
        } catch (error) {
            console.error("Error rendering chart:", error);
        }
    }

    updateSymbolChart() {
        if (!this.symbolChart) return;

        this.symbolChart.data = {
            labels: this.currentOhlcData.map((_, i) => i),
            datasets: [{
                data: this.currentOhlcData.map(d => d?.c || 0),
                borderColor: 'rgba(0, 0, 0, 0)',
                pointRadius: 0,
                pointHoverRadius: 7,
            }]
        };
        this.symbolChart.update();
    }

    async updateSignalsChart(forced = false) {
        const interval = this.currentInterval;
        const lastSignal = this.lastSignals[interval]?.pt4;
        const currentSignal = this.allSignalsData[interval]?.pt4?.at(-1);

        if (forced || !lastSignal || lastSignal !== currentSignal) {
            this.initSignalsChart();
            this.lastSignals[interval].pt4 = currentSignal;

            const pt4 = this.padZerosToStart(this.allSignalsData[interval]?.pt4);
            const npt9 = this.padZerosToStart(this.allSignalsData[interval]?.npt9);
            pt4.push(pt4.at(-1));
            npt9.push(npt9.at(-1));

            this.signalsChart.data = {
                labels: Array.from({ length: pt4.length }, (_, i) => i),
                datasets: [
                    {
                        label: 'pt4',
                        data: pt4,
                        borderColor: '#ffdd33',
                        borderWidth: 1.5,
                        pointRadius: 0,
                        pointHoverRadius: 7,
                    },
                    {
                        label: 'npt9',
                        data: npt9,
                        borderColor: '#cc8c3c',
                        borderWidth: 1.5,
                        pointRadius: 0,
                        pointHoverRadius: 7,
                    }
                ]
            };
            this.signalsChart.update();
            this.updateSignalsInfo();
        }
    }
}

let chartManager;

function InitChartManager(client) {
    chartManager = new ChartManager(client);
    window.RenderChart = symbol => chartManager.renderChart(symbol);
}

window.InitChartManager = InitChartManager;
