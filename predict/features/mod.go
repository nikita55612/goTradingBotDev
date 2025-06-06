package features

import (
	"encoding/json"
	"fmt"
	"goTradingBot/cdl"
	"goTradingBot/ta"
	"goTradingBot/utils"
	"goTradingBot/utils/norm"
	"goTradingBot/utils/numeric"
	"os"
	"sync"
)

type Generator struct {
	absFeatures []*absFeature
}

func (fg *Generator) Save(path string) error {
	jsonData, err := json.MarshalIndent(fg.absFeatures, "", "    ")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func NewGeneratorFromFile(path string) (*Generator, error) {
	fileData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var absFeatures []*absFeature
	err = json.Unmarshal(fileData, &absFeatures)
	if err != nil {
		return nil, err
	}
	Generator := Generator{
		absFeatures: absFeatures,
	}
	return &Generator, nil
}

func (fg *Generator) Labels() []string {
	labels := make([]string, len(fg.absFeatures))
	for n, s := range fg.absFeatures {
		labels[n] = s.Label()
	}
	return labels
}

func (fg *Generator) GenTranspose(candles []cdl.Candle, start, end int) [][]float64 {
	return numeric.TransposeMatrix(fg.Gen(candles, start, end))
}

// Если end < 0, без обрезки с конца
func (fg *Generator) Gen(candles []cdl.Candle, start, end int) [][]float64 {
	// n := len(candles)
	// if n < 100 {
	// 	panic("GenFeatures: Недостаточно данных")
	// }
	if end < 0 {
		end = len(candles)
	}
	featuresList := make([][]float64, len(fg.absFeatures))
	var wg sync.WaitGroup
	for n, f := range fg.absFeatures {
		if f.IsShift || f.IsField {
			continue
		}
		wg.Add(1)
		go func(index int, feature *absFeature) {
			defer wg.Done()

			zScorePeriod := feature.Params["zScorePeriod"].(int)
			if feature.Type == argFT || feature.Type == ratioFT {
				var ind []float64
				if feature.Type == argFT {
					ind = cdl.ListOfCandleArg(candles, cdl.CandleArg(feature.Name))
				} else if feature.Type == ratioFT {
					ind = cdl.ListOfCandleRatio(candles, cdl.CandleRatio(feature.Name), 1)
				}
				if zScorePeriod <= 1 {
					panic("zScorePeriod <= 1")
				}
				features := norm.ZScoreNormalize(ind, zScorePeriod)
				for s := 0; s < feature.WinSize; s++ {
					featuresList[index+s] = features[start-s : end-s]
				}
				return
			}
			var features []float64
			switch feature.Name {
			case "MACD":
				arg := feature.Params["arg"].(cdl.CandleArg)
				fPeriod := feature.Params["fPeriod"].(int)
				sPeriod := feature.Params["sPeriod"].(int)
				dPeriod := feature.Params["dPeriod"].(int)
				macd := ta.NewMACD(candles, arg, fPeriod, sPeriod, dPeriod)
				totalFields := len(feature.Fields)
				for s := 0; s < feature.WinSize; s++ {
					for fn, field := range feature.Fields {
						fieldV, err := utils.GetField[[]float64](*macd, field)
						if err != nil {
							panic(err)
						}
						if zScorePeriod > 1 {
							features = norm.ZScoreNormalize(fieldV, zScorePeriod)
						} else {
							features = fieldV
						}
						featuresList[index+s*totalFields+fn] = features[start-s : end-s]
					}
				}
			case "RSI":
				arg := feature.Params["arg"].(cdl.CandleArg)
				period := feature.Params["period"].(int)
				rsi := ta.NewRSI(candles, arg, period).Res
				if zScorePeriod > 1 {
					features = norm.ZScoreNormalize(rsi, zScorePeriod)
				} else {
					features = rsi
				}
				for s := 0; s < feature.WinSize; s++ {
					featuresList[index+s] = features[start-s : end-s]
				}
			case "SMA", "EMA", "VWMA":
				arg := feature.Params["arg"].(cdl.CandleArg)
				period := feature.Params["period"].(int)
				maT := ta.MaType(feature.Name)
				ma := ta.NewMovingAverage(maT, candles, arg, period).MaRes()
				if zScorePeriod <= 1 {
					panic("zScorePeriod <= 1")
				}
				features = norm.ZScoreNormalize(ma, zScorePeriod)
				for s := 0; s < feature.WinSize; s++ {
					featuresList[index+s] = features[start-s : end-s]
				}
			}
		}(n, f)
		if n%8 == 0 {
			wg.Wait()
		}
	}
	wg.Wait()
	return featuresList
}

type GeneratorBuilder interface {
	AddCandleArgs(args []cdl.CandleArg, period, WinSize int) GeneratorBuilder
	AddCandleRatios(ratios []cdl.CandleRatio, period, WinSize int) GeneratorBuilder
	AddMACD(fields []string, arg cdl.CandleArg, fPeriod, sPeriod, dPeriod, zScorePeriod, winSize int) GeneratorBuilder
	AddRSI(arg cdl.CandleArg, period, zScorePeriod, winSize int) GeneratorBuilder
	AddMovingAverage(maT ta.MaType, arg cdl.CandleArg, period, zScorePeriod, winSize int) GeneratorBuilder
	Build() *Generator
}

func NewGeneratorBuilder() GeneratorBuilder {
	return &fGB{
		fg: &Generator{},
	}
}

type FeatureType string

const (
	argFT       FeatureType = "A"
	ratioFT     FeatureType = "R"
	indicatorFT FeatureType = "I"
)

type absFeature struct {
	Type        FeatureType    `json:"type"`
	Name        string         `json:"name"`
	Params      map[string]any `json:"params"`
	OrderParams []string       `json:"orderParams"`
	IsShift     bool           `json:"isShift"`
	WinSize     int            `json:"winSize"`
	Fields      []string       `json:"fields"`
	IsField     bool           `json:"isField"`
}

func (f *absFeature) Label() string {
	label := fmt.Sprintf("%s-%s", f.Type, f.Name)
	for _, k := range f.OrderParams {
		label = fmt.Sprintf("%s-%s%v", label, string(k[0]), f.Params[k])
	}
	return label
}

type fGB struct {
	fg *Generator
}

func (fgb *fGB) AddCandleArgs(args []cdl.CandleArg, zScorePeriod, winSize int) GeneratorBuilder {
	for _, a := range args {
		for shift := 0; shift < winSize; shift++ {
			f := &absFeature{
				Type: argFT,
				Name: string(a),
				Params: map[string]any{
					"zScorePeriod": zScorePeriod,
					"shift":        shift,
				},
				OrderParams: []string{"zScorePeriod", "shift"},
				IsShift:     shift > 0,
				WinSize:     winSize,
				IsField:     false,
			}
			fgb.fg.absFeatures = append(fgb.fg.absFeatures, f)
		}

	}
	return fgb
}

func (fgb *fGB) AddCandleRatios(ratios []cdl.CandleRatio, zScorePeriod, winSize int) GeneratorBuilder {
	for _, r := range ratios {
		for shift := 0; shift < winSize; shift++ {
			f := &absFeature{
				Type: ratioFT,
				Name: string(r),
				Params: map[string]any{
					"zScorePeriod": zScorePeriod,
					"shift":        shift,
				},
				OrderParams: []string{"zScorePeriod", "shift"},
				IsShift:     shift > 0,
				WinSize:     winSize,
				IsField:     false,
			}
			fgb.fg.absFeatures = append(fgb.fg.absFeatures, f)
		}
	}
	return fgb
}

func (fgb *fGB) AddMACD(fields []string, arg cdl.CandleArg, fPeriod, sPeriod, dPeriod, zScorePeriod, winSize int) GeneratorBuilder {
	for shift := 0; shift < winSize; shift++ {
		for fn, field := range fields {
			var absFeatureFields []string
			if fn == 0 {
				absFeatureFields = fields
			}
			f := &absFeature{
				Type: indicatorFT,
				Name: "MACD",
				Params: map[string]any{
					"field":        field,
					"arg":          arg,
					"fPeriod":      fPeriod,
					"sPeriod":      sPeriod,
					"dPeriod":      dPeriod,
					"zScorePeriod": zScorePeriod,
					"shift":        shift,
				},
				OrderParams: []string{"field", "arg", "fPeriod", "sPeriod", "dPeriod", "zScorePeriod", "shift"},
				IsShift:     shift > 0,
				WinSize:     winSize,
				Fields:      absFeatureFields,
				IsField:     fn > 0,
			}
			fgb.fg.absFeatures = append(fgb.fg.absFeatures, f)
		}
	}
	return fgb
}

func (fgb *fGB) AddRSI(arg cdl.CandleArg, period, zScorePeriod, winSize int) GeneratorBuilder {
	for shift := 0; shift < winSize; shift++ {
		f := &absFeature{
			Type: indicatorFT,
			Name: "RSI",
			Params: map[string]any{
				"arg":          arg,
				"period":       period,
				"zScorePeriod": zScorePeriod,
				"shift":        shift,
			},
			OrderParams: []string{"arg", "period", "zScorePeriod", "shift"},
			IsShift:     shift > 0,
			WinSize:     winSize,
			IsField:     false,
		}
		fgb.fg.absFeatures = append(fgb.fg.absFeatures, f)
	}
	return fgb
}

func (fgb *fGB) AddMovingAverage(maT ta.MaType, arg cdl.CandleArg, period, zScorePeriod, winSize int) GeneratorBuilder {
	for shift := 0; shift < winSize; shift++ {
		f := &absFeature{
			Type: indicatorFT,
			Name: string(maT),
			Params: map[string]any{
				"arg":          arg,
				"period":       period,
				"zScorePeriod": zScorePeriod,
				"shift":        shift,
			},
			OrderParams: []string{"arg", "period", "zScorePeriod", "shift"},
			IsShift:     shift > 0,
			WinSize:     winSize,
			IsField:     false,
		}
		fgb.fg.absFeatures = append(fgb.fg.absFeatures, f)
	}
	return fgb
}

func (fgb *fGB) Build() *Generator {
	return fgb.fg
}
