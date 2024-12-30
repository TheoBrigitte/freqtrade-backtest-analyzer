package main

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"

	"gonum.org/v1/gonum/stat"
)

// String returns a string representation of the MinimalROISorted
// in the format "name:value name:value ..."
func (mr MinimalROISorted) String() string {
	var output []string
	for _, v := range mr {
		output = append(output, fmt.Sprintf("%s:%.3f", v.Name, v.Value))
	}

	return strings.Join(output, "  ")
}

// GetReportIndexByReason returns the ExitReasonReport with the given reason
func (ers *ExitReasonReports) GetReportIndexByReason(reason string) *ExitReasonReport {
	for i := range *ers {
		if (*ers)[i].Reason == reason {
			return &(*ers)[i]
		}
	}

	return nil
}

// AddTrade adds a trade to the ExitReasonReports
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

// Compute computes values for the ExitReasonReports
// such as total profit, average profit, average duration, and standard deviation of duration
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

// sortMinimalROI sorts the MinimalROI map by value
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

// StrategyReport returns a StrategyReport for the Strategy
func (s Strategy) StrategyReport() StrategyReport {
	var strategyReport StrategyReport

	strategyReport.ExitReasonReports = s.StrategyExitReasonReport()

	return strategyReport
}

// GetExitReasons returns the exit reasons for a trade
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

// StrategyExitReasonReport returns an ExitReasonReports for the Strategy
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
