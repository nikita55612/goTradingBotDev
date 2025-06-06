import csv
import numpy as np
import matplotlib.pyplot as plt
from matplotlib.colors import LinearSegmentedColormap
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


def plot_pred(data):
    # binary_values = data['nci']
    # binary_values_true = [0. if i < 0.5 else 1. for i in binary_values]
    # price = data['price']
    # predictions["pt3_signal"] = signals[0]
	# predictions["npt7_signal"] = signals[1]
	# predictions["tqz_signal"] = signals[2]
	# predictions["ntqz_signal"] = signals[3]
    sig = "TrendQualityZone1H"
    signals = data[sig]
    norm_signals = [0. if i < 0.5 else 1. for i in data[sig]]
    true_signals = data[f'tqz_signal']
    ss = 0
    for i in range(len(true_signals)):
        if true_signals[i] == norm_signals[i]:
            ss += 1
    print(ss / len(true_signals))


    fig, ax = plt.subplots(figsize=(12, 6))

    n = len(data['open'])

    # Ширина свечей
    width = 0.6

    for i in range(1, n):
        # Определяем цвет свечи
        color = 'green' if signals[i-0] > 0.5 else 'red'

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


# Читаем файл и строим график
data = read_csv('predictions.csv')
plot_pred(data)
