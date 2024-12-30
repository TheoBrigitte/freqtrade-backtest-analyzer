package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func (br BacktestResult) Print() {
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
		columnConfig := []table.ColumnConfig{
			{Name: "Exits", Align: text.AlignRight},
			{Name: "Avg Profit %", Align: text.AlignRight, Transformer: numberTransformer},
			{Name: "Tot Profit", Align: text.AlignRight, Transformer: numberTransformer},
			{Name: "Tot Profit %", Align: text.AlignRight, Transformer: numberTransformer},
			{Name: "Avg Duration", Align: text.AlignRight, Transformer: minuteDurationTransformer},
			{Name: "StdDev Duration", Align: text.AlignRight, Transformer: minuteDurationTransformer},
		}

		tExits := table.NewWriter()
		tExits.SetOutputMirror(os.Stdout)
		tExits.AppendHeader(table.Row{"Exit Reason", "Exits", "Avg Profit %", "Tot Profit", "Tot Profit %", "Avg Duration", "StdDev Duration"})
		tExits.SetColumnConfigs(columnConfig)
		tExits.SortBy([]table.SortBy{
			{Name: "Exits", Mode: table.DscNumeric},
		})

		tROIExits := table.NewWriter()
		tROIExits.SetOutputMirror(os.Stdout)
		tROIExits.AppendHeader(table.Row{"ROI exit Reason", "Exits", "Avg Profit %", "Tot Profit", "Tot Profit %", "Avg Duration", "StdDev Duration"})
		tROIExits.SetColumnConfigs(columnConfig)
		tROIExits.SortBy([]table.SortBy{
			{Name: "Exits", Mode: table.DscNumeric},
		})

		appendRow(tExits, tROIExits, strategyReport.ExitReasonReports)

		tExits.Render()
		tROIExits.Render()

		// Win loss report
		tWinLoss := table.NewWriter()
		tWinLoss.SetOutputMirror(os.Stdout)
		tWinLoss.SetColumnConfigs([]table.ColumnConfig{
			{Name: "Entries", Align: text.AlignRight},
			{Name: "Avg Profit %", Align: text.AlignRight, Transformer: floatTransformer},
			{Name: "Cum Profit", Align: text.AlignRight, Transformer: floatTransformer},
			{Name: "Tot Profit USDT", Align: text.AlignRight, Transformer: priceTransformer},
			{Name: "Tot Profit %", Align: text.AlignRight, Transformer: percentageTransformer},
			{Name: "Avg Duration", Align: text.AlignRight, Transformer: secondDurationTransformer},
			{Name: "Wins", Align: text.AlignRight},
			{Name: "Draws", Align: text.AlignRight},
			{Name: "Loss", Align: text.AlignRight},
			{Name: "Win %", Align: text.AlignRight, Transformer: percentageTransformer},
		})
		tWinLoss.AppendHeader(table.Row{"TAG", "Entries", "Avg Profit %", "Cum Profit", "Tot Profit USDT", "Tot Profit %", "Avg Duration", "Win", "Draws", "Loss", "Win %"})
		tWinLoss.AppendRow([]interface{}{"TOTAL", s.TotalTrades, s.ProfitTotalAbs / float64(s.TotalTrades), 0.0, s.ProfitTotalAbs, s.ProfitTotal, s.HoldingAvgDuration, s.Wins, s.Draws, s.Losses, float64(s.Wins) / float64(s.TotalTrades)})
		tWinLoss.Render()

		// General metric report
		tMetrics := table.NewWriter()
		tMetrics.SetOutputMirror(os.Stdout)
		tMetrics.AppendHeader(table.Row{"Metric", "Value"})
		tMetrics.AppendRow([]interface{}{"Strategy", strategyName})
		tMetrics.AppendRow([]interface{}{"Minimal ROI", s.MinimalROISorted.String()})
		tMetrics.AppendRow([]interface{}{"Stoploss", fmt.Sprintf("%.4f", s.Stoploss)})
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
		tMetrics.AppendRow([]interface{}{"Avg profit %", percentageTransformer(s.ProfitMean)})
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
		tMetrics.AppendRow([]interface{}{"Score", s.Score()})
		tMetrics.Render()
	}
}

func appendRow(tExits, tROIExits table.Writer, reports ExitReasonReports) {
	for _, v := range reports {
		if len(v.Reason) > 3 && strings.HasPrefix(v.Reason, "roi") {
			tROIExits.AppendRow([]interface{}{v.Reason, v.Exits, v.AvgProfit, v.TotalProfit, v.TotalProfitPercentage, v.AvgDuration, v.StdDevDuration})
		} else {
			tExits.AppendRow([]interface{}{v.Reason, v.Exits, v.AvgProfit, v.TotalProfit, v.TotalProfitPercentage, v.AvgDuration, v.StdDevDuration})
		}
		if len(v.ExitReasonReports) > 0 {
			appendRow(tExits, tROIExits, v.ExitReasonReports)
		}
	}
}
