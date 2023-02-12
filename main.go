package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"gonum.org/v1/gonum/stat"
)

// Data structure
type BacktestResult struct {
	Strategy map[string]Strategy `json:"strategy"`
}

type Strategy struct {
	// General purpose information
	BacktestStart CustomTime `json:"backtest_start"`
	BacktestEnd   CustomTime `json:"backtest_end"`
	BacktestDays  int        `json:"backtest_days"`

	// Finance
	TotalTrades         int     `json:"total_trades"`
	StartingBalance     float64 `json:"starting_balance"`
	FinalBalance        float64 `json:"final_balance"`
	ProfitTotalAbs      float64 `json:"profit_total_abs"`
	ProfitTotal         float64 `json:"profit_total"`
	ProfitFactor        float64 `json:"profit_factor"`
	CAGR                float64 `json:"cagr"`
	Sortino             float64 `json:"sortino"`
	Sharpe              float64 `json:"sharpe"`
	Calmar              float64 `json:"calmar"`
	Expectancy          float64 `json:"expectancy"`
	TradesPerDay        float64 `json:"trades_per_day"`
	AvgStakeAmount      float64 `json:"avg_stake_amount"`
	TotalVolume         float64 `json:"total_volume"`
	TradeCountLong      int     `json:"trade_count_long"`
	TradeCountShort     int     `json:"trade_count_short"`
	ProfitTotalLong     float64 `json:"profit_total_long"`
	ProfitTotalShort    float64 `json:"profit_total_short"`
	ProfitTotalLongAbs  float64 `json:"profit_total_long_abs"`
	ProfitTotalShortAbs float64 `json:"profit_total_short_abs"`

	Wins   int `json:"wins"`
	Draws  int `json:"draws"`
	Losses int `json:"losses"`

	HoldingAvgDuration float64 `json:"holding_avg_s"`
	WinnderAvgDuration float64 `json:"winner_holding_avg_s"`
	LoserAvgDuration   float64 `json:"loser_holding_avg_s"`

	MinBalance         float64    `json:"csum_min"`
	MaxBalance         float64    `json:"csum_max"`
	DrawdownRelative   float64    `json:"max_relative_drawdown"`
	DrawdownAbsAccount float64    `json:"max_drawdown_account"`
	DrawdownAbs        float64    `json:"max_drawdown_abs"`
	DrawdownHigh       float64    `json:"max_drawdown_high"`
	DrawdownLow        float64    `json:"max_drawdown_low"`
	DrawdownStart      CustomTime `json:"drawdown_start"`
	DrawdownEnd        CustomTime `json:"drawdown_end"`
	MarketChange       float64    `json:"market_change"`

	// Trading information
	Trades []Trade `json:"trades"`

	// Config
	MaxOpenTrades int                `json:"max_open_trades"`
	MinimalROI    map[string]float64 `json:"minimal_roi"`
	StakeCurrency string             `json:"stake_currency"`

	// Custom fields
	MinimalROISorted MinimalROISorted
}

type Trade struct {
	ExitReason    string  `json:"exit_reason"`
	ProfitAbs     float64 `json:"profit_abs"`
	ProfitRatio   float64 `json:"profit_ratio"`
	IsOpen        bool    `json:"is_open"`
	TradeDuration int     `json:"trade_duration"`
}

// Custom structure
type MinimalROISorted []MinimalROI

type MinimalROI struct {
	Name  string
	Value float64
}

type CustomTime struct {
	time.Time
}

const dateTimeFormat = "2006-01-02 15:04:05"

func (t *CustomTime) UnmarshalJSON(b []byte) (err error) {
	value := strings.Trim(string(b), `"`)
	if value == "" || value == "null" {
		return nil
	}

	date, err := time.Parse(dateTimeFormat, value)
	if err != nil {
		return err
	}
	t.Time = date
	return
}

// Report structure
type StrategyReport struct {
	ExitReasonReports ExitReasonReports
}

type ExitReasonReports []ExitReasonReport

type ExitReasonReport struct {
	Reason                string
	Exits                 int
	ProfitAbs             []float64
	TradeDurations        []int
	AvgDuration           int
	StdDevDuration        float64
	AvgProfit             float64
	TotalProfit           float64
	TotalProfitPercentage float64
	ExitReasonReports     ExitReasonReports
}

// Methods
func (mr MinimalROISorted) String() string {
	var output []string
	for _, v := range mr {
		output = append(output, fmt.Sprintf("%s:%.3f", v.Name, v.Value))
	}

	return strings.Join(output, "  ")
}

func (ers *ExitReasonReports) GetReportIndexByReason(reason string) *ExitReasonReport {
	for i := range *ers {
		if (*ers)[i].Reason == reason {
			return &(*ers)[i]
		}
	}

	return nil
}

func (ers *ExitReasonReports) AddTrade(t Trade, reasons []string, reasonIndex int) int {
	var zeroDuration = 0
	reason := reasons[reasonIndex]
	if !t.IsOpen {
		var er *ExitReasonReport
		er = ers.GetReportIndexByReason(reason)
		if er == nil {
			*ers = append(*ers, ExitReasonReport{
				Reason: reasons[reasonIndex],
			})
			er = ers.GetReportIndexByReason(reason)
		}

		er.Exits++
		if t.TradeDuration > 0 {
			er.TradeDurations = append(er.TradeDurations, t.TradeDuration)
		} else {
			zeroDuration = zeroDuration + 1
		}
		er.ProfitAbs = append(er.ProfitAbs, t.ProfitAbs)
		if reason == "roi inf+" {
			//log.Printf("profit_abs: %.3f  profit_ratio: %.17f\n", t.ProfitAbs, t.ProfitRatio)
		}

		if len(reasons) > reasonIndex+1 {
			zd := er.ExitReasonReports.AddTrade(t, reasons, reasonIndex+1)
			zeroDuration = zeroDuration + zd
		}
	}
	return zeroDuration
}

func (ers *ExitReasonReports) Compute() {
	var absoluteTotal float64 = 0
	for k := range *ers {
		var er *ExitReasonReport
		er = &(*ers)[k]

		var totalProfit float64 = 0
		for _, value := range er.ProfitAbs {
			totalProfit = totalProfit + value
		}
		absoluteTotal = absoluteTotal + math.Abs(totalProfit)
		er.TotalProfit = totalProfit
		er.AvgProfit = totalProfit / float64(len(er.ProfitAbs))

		var totalDuration = 0
		for _, value := range er.TradeDurations {
			totalDuration = totalDuration + value
		}

		lenTradeDurations := len(er.TradeDurations)
		if lenTradeDurations == 0 {
			lenTradeDurations = 1
		}
		er.AvgDuration = totalDuration / lenTradeDurations

		tDurations := []float64{}
		for _, v := range er.TradeDurations {
			tDurations = append(tDurations, float64(v))
		}
		er.StdDevDuration = stat.StdDev(tDurations, nil)
	}

	for k := range *ers {
		er := &(*ers)[k]

		er.TotalProfitPercentage = er.TotalProfit / absoluteTotal * 100

		er.ExitReasonReports.Compute()
	}
}

func (s *Strategy) sortMinimalROI() {
	keys := make([]string, 0, len(s.MinimalROI))

	// get keys
	for key := range s.MinimalROI {
		keys = append(keys, key)
	}

	// sort keys using values
	sort.SliceStable(keys, func(i, j int) bool {
		return s.MinimalROI[keys[i]] > s.MinimalROI[keys[j]]
	})

	// rebuild slice
	var sortedMinimalROI MinimalROISorted
	for _, k := range keys {
		sortedMinimalROI = append(sortedMinimalROI, MinimalROI{
			Name:  k,
			Value: s.MinimalROI[k],
		})
	}

	s.MinimalROISorted = sortedMinimalROI
}

func (s Strategy) StrategyReport() StrategyReport {
	var strategyReport StrategyReport

	strategyReport.ExitReasonReports = s.StrategyExitReasonReport()

	return strategyReport
}

func (s Strategy) GetExitReasons(t Trade, id int) []string {
	//log.Printf("> minimalROI\n%v\n", s.MinimalROISorted)
	if t.ExitReason == "roi" {
		for _, mr := range s.MinimalROISorted {
			if t.ProfitRatio >= mr.Value {
				//fmt.Printf("> DEBUG\n  id: %d  roiName: %s  roiValue: %f  profitRatio: %f\n", id, mr.Name, mr.Value, t.ProfitRatio)
				return []string{"roi", fmt.Sprintf("roi %s:%.3f", mr.Name, mr.Value)}
			}
		}

		if t.ProfitRatio >= 0.02 {
			return []string{"roi", "roi art:0.02"}
		}
		if t.ProfitRatio >= 0.01 {
			return []string{"roi", "roi art:0.01"}
		}
		//fmt.Printf("> DEBUG\n  id: %d  reason: %s  profitRatio: %f\n", id, t.ExitReason, t.ProfitRatio)
		return []string{"roi", "roi inf+"}
	}

	//fmt.Printf("> DEBUG\n  id: %d  reason: %s  profitRatio: %f\n", id, t.ExitReason, t.ProfitRatio)
	return []string{t.ExitReason}
}

func (s Strategy) StrategyExitReasonReport() ExitReasonReports {
	var exitReasonReports ExitReasonReports

	log.Printf("> processing %d trades\n", len(s.Trades))
	zeroDuration := 0
	for id, t := range s.Trades {
		exitReasons := s.GetExitReasons(t, id)
		zd := exitReasonReports.AddTrade(t, exitReasons, 0)
		zeroDuration = zeroDuration + zd

	}
	log.Printf("> WARNING %d trades have duration=0\n", zeroDuration)

	exitReasonReports.Compute()

	return exitReasonReports
}

func appendRow(t table.Writer, reports ExitReasonReports) {
	for _, v := range reports {
		t.AppendRow([]interface{}{v.Reason, v.Exits, v.AvgProfit, v.TotalProfit, v.TotalProfitPercentage, v.AvgDuration, v.StdDevDuration})
		if len(v.ExitReasonReports) > 0 {
			appendRow(t, v.ExitReasonReports)
		}
	}
}
func (br BacktestResult) PrintExitReasonsAverage() {
	// Transformers used to display values
	numberTransformer := text.NewNumberTransformer("%.2f")
	minuteDurationTransformer := func(val interface{}) string {
		d, err := time.ParseDuration(fmt.Sprintf("%vm", val))
		if err != nil {
			return "error"
		}
		return d.String()
	}
	secondDurationTransformer := func(val interface{}) string {
		d, err := time.ParseDuration(fmt.Sprintf("%vs", val))
		if err != nil {
			return "error"
		}
		return d.String()
	}
	percentageTransformer := func(val interface{}) string {
		v, ok := val.(float64)
		if !ok {
			return "error"
		}
		return fmt.Sprintf("%.2f%%", v*100)
	}
	floatTransformer := func(val interface{}) string {
		return fmt.Sprintf("%.2f", val)
	}

	for strategyName, s := range br.Strategy {
		priceTransformer := func(val interface{}) string {
			return fmt.Sprintf("%.3f %s", val, s.StakeCurrency)
		}

		// Sort ROI by value, so we can break down ROI exit by value.
		// e.g. roi setting -> 0:0.1  60:0.02; 100 roi exits does not tell anything
		//      but 80 exit >= 0.1 20  exit >= 0.02 tells how many exits per ROI setting
		s.sortMinimalROI()

		// Compute report
		strategyReport := s.StrategyReport()

		// Exit signals report
		tExits := table.NewWriter()
		tExits.SetOutputMirror(os.Stdout)
		tExits.AppendHeader(table.Row{"Exit Reason", "Exits", "Avg Profit %", "Tot Profit", "Tot Profit %", "Avg Duration", "StdDev Duration"})

		appendRow(tExits, strategyReport.ExitReasonReports)

		tExits.SetColumnConfigs([]table.ColumnConfig{
			{Name: "Exits", Align: text.AlignRight},
			{Name: "Avg Profit %", Align: text.AlignRight, Transformer: numberTransformer},
			{Name: "Tot Profit", Align: text.AlignRight, Transformer: numberTransformer},
			{Name: "Tot Profit %", Align: text.AlignRight, Transformer: numberTransformer},
			{Name: "Avg Duration", Align: text.AlignRight, Transformer: minuteDurationTransformer},
			{Name: "StdDev Duration", Align: text.AlignRight, Transformer: minuteDurationTransformer},
		}

		tExits.SortBy([]table.SortBy{
			{Name: "Exits", Mode: table.DscNumeric},
		})

		// General metric report
		tMetrics := table.NewWriter()
		tMetrics.SetOutputMirror(os.Stdout)
		tMetrics.AppendHeader(table.Row{"Metric", "Value"})
		tMetrics.AppendRow([]interface{}{"Strategy", strategyName})
		tMetrics.AppendRow([]interface{}{"Minimal ROI", s.MinimalROISorted.String()})
		tMetrics.AppendRow([]interface{}{"", ""})
		tMetrics.AppendRow([]interface{}{"Backtest from", s.BacktestStart})
		tMetrics.AppendRow([]interface{}{"Backtest to", s.BacktestEnd})
		tMetrics.AppendRow([]interface{}{"Max open trades", s.MaxOpenTrades})
		tMetrics.AppendRow([]interface{}{"", ""})
		tMetrics.AppendRow([]interface{}{"Total/Daily Avg Trades", fmt.Sprintf("%d / %.2f", s.TotalTrades, float64(s.TotalTrades)/float64(s.BacktestDays))})
		tMetrics.AppendRow([]interface{}{"Starting balance", priceTransformer(s.StartingBalance)})
		tMetrics.AppendRow([]interface{}{"Final balance", priceTransformer(s.FinalBalance)})
		tMetrics.AppendRow([]interface{}{"Absolute profit", priceTransformer(s.ProfitTotalAbs)})
		tMetrics.AppendRow([]interface{}{"Total profit %", percentageTransformer(s.ProfitTotal)})
		tMetrics.AppendRow([]interface{}{"CAGR %", percentageTransformer(s.CAGR)})
		tMetrics.AppendRow([]interface{}{"Sortino", floatTransformer(s.Sortino)})
		tMetrics.AppendRow([]interface{}{"Sharpe", floatTransformer(s.Sharpe)})
		tMetrics.AppendRow([]interface{}{"Calmar", floatTransformer(s.Calmar)})
		tMetrics.AppendRow([]interface{}{"Profit factor", floatTransformer(s.ProfitFactor)})
		tMetrics.AppendRow([]interface{}{"Expectancy", floatTransformer(s.Expectancy)})
		tMetrics.AppendRow([]interface{}{"Trades per day", floatTransformer(s.TradesPerDay)})
		tMetrics.AppendRow([]interface{}{"Avg. daily profit %", fmt.Sprintf("%.2f", float64(s.ProfitTotal*100)/float64(s.BacktestDays))})
		tMetrics.AppendRow([]interface{}{"Avg. stake amount", priceTransformer(s.AvgStakeAmount)})
		tMetrics.AppendRow([]interface{}{"Total trade volume", priceTransformer(s.TotalVolume)})
		tMetrics.AppendRow([]interface{}{"", ""})
		tMetrics.AppendRow([]interface{}{"Long / Short", fmt.Sprintf("%d / %d", s.TradeCountLong, s.TradeCountShort)})
		tMetrics.AppendRow([]interface{}{"Total profit Long %", percentageTransformer(s.ProfitTotalLong)})
		tMetrics.AppendRow([]interface{}{"Total profit Short %", percentageTransformer(s.ProfitTotalShort)})
		tMetrics.AppendRow([]interface{}{"Absolute profit Long", priceTransformer(s.ProfitTotalLongAbs)})
		tMetrics.AppendRow([]interface{}{"Absolute profit Short", priceTransformer(s.ProfitTotalShortAbs)})
		tMetrics.AppendRow([]interface{}{"", ""})
		tMetrics.AppendRow([]interface{}{"Avg. Duration Winners", secondDurationTransformer(s.WinnderAvgDuration)})
		tMetrics.AppendRow([]interface{}{"Avg. Duration Loser", secondDurationTransformer(s.LoserAvgDuration)})
		tMetrics.AppendRow([]interface{}{"", ""})
		tMetrics.AppendRow([]interface{}{"Min balance", priceTransformer(s.MinBalance)})
		tMetrics.AppendRow([]interface{}{"Max balance", priceTransformer(s.MaxBalance)})
		tMetrics.AppendRow([]interface{}{"Max % of account underwater", percentageTransformer(s.DrawdownRelative)})
		tMetrics.AppendRow([]interface{}{"Absolute Drawdown (Account)", percentageTransformer(s.DrawdownAbsAccount)})
		tMetrics.AppendRow([]interface{}{"Absolute Drawdown", priceTransformer(s.DrawdownAbs)})
		tMetrics.AppendRow([]interface{}{"Drawdown high", priceTransformer(s.DrawdownHigh)})
		tMetrics.AppendRow([]interface{}{"Drawdown low", priceTransformer(s.DrawdownLow)})
		tMetrics.AppendRow([]interface{}{"Drawdown Start", s.DrawdownStart})
		tMetrics.AppendRow([]interface{}{"Drawdown End", s.DrawdownEnd})
		tMetrics.AppendRow([]interface{}{"Market change", percentageTransformer(s.MarketChange)})

		tExits.Render()
		tMetrics.Render()
	}

}

func loadBacktestResultFromFilename(filename string) (*BacktestResult, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	log.Printf("> opened %s\n", filename)

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	log.Printf("> read %d bytes from %s\n", len(data), filename)

	err = f.Close()
	if err != nil {
		log.Printf("> WARNING: failed to close file %s: %v\n", filename, err)
	}

	var backtestResult *BacktestResult
	err = json.Unmarshal(data, &backtestResult)
	if err != nil {
		return nil, err
	}

	return backtestResult, nil
}

func main() {
	log.Println("> start")

	if len(os.Args[1:]) < 1 {
		log.Fatalf("expecting 1 argument got %d\n", len(os.Args[1:]))
	}

	var err error

	var backtestResult *BacktestResult
	{
		filename := os.Args[1]
		backtestResult, err = loadBacktestResultFromFilename(filename)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("> loaded backtest result\n")
	}

	backtestResult.PrintExitReasonsAverage()
}
