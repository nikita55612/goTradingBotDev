import csv
import matplotlib.pyplot as plt
import matplotlib.patches as patches

def read_csv(filename):
    with open(filename, newline='') as csvfile:
        reader = csv.reader(csvfile)
        headers = next(reader)
        data = {col: [] for col in headers}

        for row in reader:
            if row:
                for col, value in zip(headers, row):
                    data[col].append(float(value))

    return data

def plot_candlestick(data):
    # Создаем фигуру и оси
    fig, ax = plt.subplots(figsize=(12, 6))

    # Количество свечей
    n = len(data['open'])
    signal = data['signal']

    # Ширина свечей
    width = 0.6

    for i in range(n-1):
        # Определяем цвет свечи
        color = 'green' if signal[i] > 0.5 else 'red'

        # Линия (тендшка) от high до low
        ax.plot([i, i], [data['low'][i], data['high'][i]], color=color, linewidth=1)

        # Прямоугольник свечи (тело)
        rect = patches.Rectangle(
            (i - width/2, min(data['open'][i], data['close'][i])),
            width,
            abs(data['open'][i] - data['close'][i]),
            facecolor=color,
            edgecolor=color
        )
        ax.add_patch(rect)

    # Настройки графика
    ax.set_title('Японские свечи')
    ax.set_xlabel('Номер свечи')
    ax.set_ylabel('Цена')

    plt.tight_layout()
    plt.show()

data = read_csv("signals.csv")
plot_candlestick(data)
