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
    # Создаем фигуру и оси (2 графика: свечи сверху, индикатор снизу)
    fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(
        12, 8), gridspec_kw={'height_ratios': [3, 1]})

    # Количество свечей
    n = len(data['open'])

    # Ширина свечей
    width = 0.6

    # График свечей (верхний)
    for i in range(n):
        # Определяем цвет свечи
        color = 'green' if data['close'][i] >= data['open'][i] else 'red'

        # Линия (тендшка) от high до low
        ax1.plot([i, i], [data['low'][i], data['high'][i]],
                 color=color, linewidth=1)

        # Прямоугольник свечи (тело)
        rect = patches.Rectangle(
            (i - width/2, min(data['open'][i], data['close'][i])),
            width,
            abs(data['open'][i] - data['close'][i]),
            facecolor=color,
            edgecolor=color
        )
        ax1.add_patch(rect)

    # Настройки верхнего графика
    ax1.set_title('Японские свечи')
    ax1.set_ylabel('Цена')

    # График индикатора (нижний)
    if 'ind' in data:
        ax2.plot(range(n), data['ind'], color='blue', linewidth=1)
        ax2.set_title('Индикатор')
        ax2.set_xlabel('Номер свечи')
        ax2.set_ylabel('Значение')

        ax1.set_xlim(-0.5, n-0.5)
        ax2.set_xlim(-0.5, n-0.5)
    else:
        print("В данных отсутствует столбец 'ind'")

    plt.tight_layout()
    plt.show()


data = read_csv("ind.csv")
plot_candlestick(data)
