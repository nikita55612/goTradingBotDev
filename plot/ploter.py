import csv
import sys
import matplotlib.pyplot as plt
import argparse

def read_csv(csv_path):
    """
    Читает CSV файл и возвращает заголовки и данные.
    """
    with open(csv_path, 'r', newline='', encoding='utf-8') as csvfile:
        reader = csv.reader(csvfile)
        headers = next(reader)  # Получаем заголовки (первая строка)
        data = []
        for row in reader:
            # Пытаемся преобразовать строковые значения в числа, где возможно
            processed_row = []
            for item in row:
                try:
                    processed_row.append(float(item))
                except ValueError:
                    processed_row.append(item)
            data.append(processed_row)

    return headers, data

def plot_data(headers, data, csv_path):
    """
    Строит график по данным из CSV с темной темой.
    """
    # Транспонируем данные для доступа по колонкам
    columns = list(zip(*data))

    # Проверяем, все ли колонки содержат числовые данные
    numeric_columns = []
    numeric_headers = []

    for i, column in enumerate(columns):
        if all(isinstance(item, (int, float)) for item in column):
            numeric_columns.append(column)
            numeric_headers.append(headers[i])

    if not numeric_columns:
        print("Не найдено числовых колонок для построения графика!")
        return

    # Устанавливаем стиль темной темы
    plt.style.use('dark_background')

    # Создаём график
    fig, ax = plt.subplots(figsize=(14, 9))

    # Создаем интенсивные цвета для линий
    colors = ['#FFFFFF', '#FF5733', '#33FF57', '#3357FF', '#F033FF',
          '#FF33F0', '#33FFF0', '#F0FF33', '#FFC733', '#33FFAA', '#33C7FF']

    # Создаём X-координаты как последовательность целых чисел
    x_values = list(range(len(data)))

    # Строим линии для каждой колонки
    for i, column in enumerate(numeric_columns):
        # Используем цвет из списка или генерируем, если колонок больше чем цветов
        color_index = i % len(colors)
        plt.plot(x_values, column, color=colors[color_index], label=numeric_headers[i])

    # Добавляем подписи к осям
    plt.xlabel('Номер строки', fontsize=12)
    plt.ylabel('Значения', fontsize=12)
    plt.title(f'График данных из {csv_path}', fontsize=16)

    # Добавляем легенду с названиями колонок
    plt.legend(fontsize=10, loc='best', facecolor='#1A1A1A', edgecolor='#666666')

    # Улучшаем внешний вид графика
    plt.tight_layout()

    # Улучшаем стиль осей
    ax.spines['top'].set_visible(False)
    ax.spines['right'].set_visible(False)
    ax.spines['bottom'].set_color('#666666')
    ax.spines['left'].set_color('#666666')

    # Показываем график
    plt.show()

def main():
    # Получаем аргументы командной строки
    parser = argparse.ArgumentParser(description='Читает CSV файл и строит график по данным.')
    parser.add_argument('csv_path', help='Путь к CSV файлу')
    args = parser.parse_args()

    csv_path = args.csv_path

    try:
        headers, data = read_csv(csv_path)
        plot_data(headers, data, csv_path)
    except Exception as e:
        print(f"Ошибка: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
