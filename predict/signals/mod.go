package signals

import (
	"encoding/json"
	"fmt"
	"goTradingBot/cdl"
	"goTradingBot/utils/numeric"
	"os"
	"sync"
)

// Generator генерирует и управляет торговыми сигналами на основе свечных данных
type Generator struct {
	absSignals []*absSignal
}

// Save сохраняет сгенерированные сигналы в JSON файл по указанному пути
func (sg *Generator) Save(path string) error {
	jsonData, err := json.MarshalIndent(sg.absSignals, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonData, 0644)
}

// NewGeneratorFromFile создает новый Generator из JSON файла
func NewGeneratorFromFile(path string) (*Generator, error) {
	fileData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var absSignals []*absSignal
	if err = json.Unmarshal(fileData, &absSignals); err != nil {
		return nil, err
	}
	return &Generator{absSignals: absSignals}, nil
}

// Labels возвращает список меток для всех сигналов
func (sg *Generator) Labels() []string {
	labels := make([]string, len(sg.absSignals))
	for n, s := range sg.absSignals {
		labels[n] = s.Label()
	}
	return labels
}

// GenTranspose генерирует матрицу сигналов и транспонирует ее
func (sg *Generator) GenTranspose(candles []cdl.Candle, start, end int) [][]float64 {
	return numeric.TransposeMatrix(sg.Gen(candles, start, end))
}

// Gen генерирует матрицу сигналов для заданного диапазона свечей
// Использует параллельную обработку для повышения производительности
func (sg *Generator) Gen(candles []cdl.Candle, start, end int) [][]float64 {
	signalsList := make([][]float64, len(sg.absSignals))
	var wg sync.WaitGroup
	for n, s := range sg.absSignals {
		wg.Add(1)
		go func(index int, signal *absSignal) {
			defer wg.Done()

			var signals []float64
			switch signal.Name {
			case "PerfectTrend":
				period := signal.Params["period"].(int)
				signals = PerfectTrend(candles, period)[start:end]
			case "NextPerfectTrend":
				period := signal.Params["period"].(int)
				signals = PerfectTrend(candles, period)[start+1 : end+1]
			// case "TrendQualityZone":
			// 	signals = TrendQualityZone(candles)[start:end]
			// case "NextTrendQualityZone":
			// 	signals = TrendQualityZone(candles)[start+1 : end+1]
			case "NextDirSignal":
				signals = NextDirSignal(candles)[start:end]
			case "NextBodyWiderSignal":
				signals = NextBodyWiderSignal(candles)[start:end]
			case "NextCandleOutsideSignal":
				signals = NextCandleOutsideSignal(candles)[start:end]
			}
			signalsList[index] = signals
		}(n, s)
		if n%8 == 0 {
			wg.Wait()
		}
	}
	wg.Wait()
	return signalsList
}

// GeneratorBuilder интерфейс для построения Generator
type GeneratorBuilder interface {
	AddPerfectTrend(period int) GeneratorBuilder
	AddNextPerfectTrend(period int) GeneratorBuilder
	// AddTrendQualityZone() GeneratorBuilder
	// AddNextTrendQualityZone() GeneratorBuilder
	AddNextDirSignal() GeneratorBuilder
	AddNextBodyWiderSignal() GeneratorBuilder
	AddNextCandleOutsideSignal() GeneratorBuilder
	Build() *Generator
}

// NewGeneratorBuilder создает новый билдер для Generator
func NewGeneratorBuilder() GeneratorBuilder {
	return &sGB{sg: &Generator{}}
}

// absSignal представляет абстрактный торговый сигнал с параметрами
type absSignal struct {
	Name   string         `json:"name"`
	Params map[string]any `json:"params"`
}

// Label генерирует читаемую метку для сигнала
func (s *absSignal) Label() string {
	label := s.Name
	for k, p := range s.Params {
		label = fmt.Sprintf("%s-%s%v", label, string(k[0]), p)
	}
	return label
}

// sGB реализация билдера для Generator
type sGB struct {
	sg *Generator
}

func (sgb *sGB) AddPerfectTrend(period int) GeneratorBuilder {
	sgb.sg.absSignals = append(sgb.sg.absSignals, &absSignal{
		Name: "PerfectTrend",
		Params: map[string]any{
			"period": period,
		},
	})
	return sgb
}

func (sgb *sGB) AddNextPerfectTrend(period int) GeneratorBuilder {
	sgb.sg.absSignals = append(sgb.sg.absSignals, &absSignal{
		Name: "NextPerfectTrend",
		Params: map[string]any{
			"period": period,
		},
	})
	return sgb
}

// func (sgb *sGB) AddTrendQualityZone() GeneratorBuilder {
// 	sgb.sg.absSignals = append(sgb.sg.absSignals, &absSignal{Name: "TrendQualityZone"})
// 	return sgb
// }

// func (sgb *sGB) AddNextTrendQualityZone() GeneratorBuilder {
// 	sgb.sg.absSignals = append(sgb.sg.absSignals, &absSignal{Name: "NextTrendQualityZone"})
// 	return sgb
// }

func (sgb *sGB) AddNextDirSignal() GeneratorBuilder {
	sgb.sg.absSignals = append(sgb.sg.absSignals, &absSignal{Name: "NextDirSignal"})
	return sgb
}

func (sgb *sGB) AddNextBodyWiderSignal() GeneratorBuilder {
	sgb.sg.absSignals = append(sgb.sg.absSignals, &absSignal{Name: "NextBodyWiderSignal"})
	return sgb
}

func (sgb *sGB) AddNextCandleOutsideSignal() GeneratorBuilder {
	sgb.sg.absSignals = append(sgb.sg.absSignals, &absSignal{Name: "NextCandleOutsideSignal"})
	return sgb
}

func (sgb *sGB) Build() *Generator {
	return sgb.sg
}
