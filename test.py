import csv

def calculate_profit(csv_file):
    # Инициализация переменных
    total_profit = 0.0
    position = None  # 'long' или 'short'
    entry_price = 0.0
    previous_signal = None

    # Статистические переменные
    trades_count = 0
    profitable_trades = 0
    losing_trades = 0
    max_profit = 0.0
    max_loss = 0.0
    long_trades = 0
    short_trades = 0
    total_long_profit = 0.0
    total_short_profit = 0.0

    with open(csv_file, 'r') as file:
        reader = csv.DictReader(file)

        for row in reader:
            price = float(row['price'])
            signal = float(row['npt7'])

            # Определяем действие на основе сигнала
            if signal > 0.5:
                current_action = 'buy'
            else:
                current_action = 'sell'

            # Если это первая строка, просто запоминаем сигнал
            if previous_signal is None:
                previous_signal = current_action
                continue

            # Если сигнал изменился
            if current_action != previous_signal:
                # Если у нас была открытая позиция, закрываем ее
                if position is not None:
                    trades_count += 1

                    if position == 'long':
                        profit = (price - entry_price) / entry_price * 100
                        long_trades += 1
                        total_long_profit += profit
                    else:  # short
                        profit = (entry_price - price) / entry_price * 100
                        short_trades += 1
                        total_short_profit += profit

                    # Обновляем статистику
                    total_profit += profit

                    if profit > 0:
                        profitable_trades += 1
                        if profit > max_profit:
                            max_profit = profit
                    else:
                        losing_trades += 1
                        if profit < max_loss:
                            max_loss = profit

                    print(f"Закрытие {position} позиции: вход {entry_price:.2f}, выход {price:.2f}, прибыль {profit:.2f}%")

                # Открываем новую позицию
                if current_action == 'buy':
                    position = 'long'
                else:
                    position = 'short'
                entry_price = price
                print(f"Открытие {position} позиции по цене {price:.2f}")

                previous_signal = current_action

    # Выводим общую статистику
    print("\n" + "="*50)
    print("ИТОГОВАЯ СТАТИСТИКА ПО ТОРГАМ")
    print("="*50)
    print(f"Общее количество сделок: {trades_count}")
    print(f"Прибыльных сделок: {profitable_trades} ({profitable_trades/trades_count*100:.1f}%)")
    print(f"Убыточных сделок: {losing_trades} ({losing_trades/trades_count*100:.1f}%)")
    print(f"Общая прибыль: {total_profit:.2f}%")
    print(f"Средняя прибыль на сделку: {total_profit/trades_count:.2f}%")
    print(f"Максимальная прибыль: {max_profit:.2f}%")
    print(f"Максимальный убыток: {max_loss:.2f}%")
    print("\nСтатистика по типам сделок:")
    print(f"Длинных сделок: {long_trades} ({long_trades/trades_count*100:.1f}%)")
    print(f"Средняя прибыль длинных сделок: {total_long_profit/long_trades if long_trades > 0 else 0:.2f}%")
    print(f"Коротких сделок: {short_trades} ({short_trades/trades_count*100:.1f}%)")
    print(f"Средняя прибыль коротких сделок: {total_short_profit/short_trades if short_trades > 0 else 0:.2f}%")

    return {
        'total_profit': total_profit,
        'trades_count': trades_count,
        'profitable_trades': profitable_trades,
        'losing_trades': losing_trades,
        'max_profit': max_profit,
        'max_loss': max_loss,
        'long_trades': long_trades,
        'short_trades': short_trades,
        'avg_profit_per_trade': total_profit/trades_count if trades_count > 0 else 0,
        'avg_long_profit': total_long_profit/long_trades if long_trades > 0 else 0,
        'avg_short_profit': total_short_profit/short_trades if short_trades > 0 else 0
    }

# Пример использования
csv_file = 'test.csv'
stats = calculate_profit(csv_file)
