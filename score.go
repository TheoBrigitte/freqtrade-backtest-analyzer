package main

import "math"

// Score returns the strategy score
func (s Strategy) Score() float64 {
	return compute_score(s.Expectancy, s.ProfitFactor, s.DrawdownRelative*100, s.ProfitMean*100, s.ProfitTotal*100)
}

// compute_score computes the score of a strategy
// based on its expectancy, profit factor, drawdown, average profit, and total profit.
func compute_score(expectancy, profitFactor, drawDown, avgProfit, totalProfit float64) float64 {
	// baseline represents values that are considered "good" for a strategy
	baseline := map[string]float64{
		"expectancy":    0.2,
		"profit_factor": 2,
		"drawdown":      2,
		"avg_profit":    1,
		"total_profit":  20,
	}

	// sensitivities represent how much each metric contributes to the final score
	sensitivities := map[string]float64{
		"expectancy":    2,
		"profit_factor": 1.5,
		"drawdown":      1.2,
		"avg_profit":    1,
		"total_profit":  1,
	}

	// weights represent the importance of each metric in the final score
	weights := map[string]float64{
		"expectancy":    0.3,
		"profit_factor": 0.25,
		"drawdown":      0.2,
		"avg_profit":    0.15,
		"total_profit":  0.10,
	}

	// compute uncapped scores for each metric
	expectancy_score := metric_score(expectancy, baseline["expectancy"], true, sensitivities["expectancy"])
	profit_factor_score := metric_score(profitFactor, baseline["profit_factor"], true, sensitivities["profit_factor"])
	drawdown_score := metric_score(drawDown, baseline["drawdown"], false, sensitivities["drawdown"])
	avg_profit_score := metric_score(avgProfit, baseline["avg_profit"], true, sensitivities["avg_profit"])
	total_profit_score := metric_score(totalProfit, baseline["total_profit"], true, sensitivities["total_profit"])

	// compute total score
	total_score := weights["expectancy"]*expectancy_score +
		weights["profit_factor"]*profit_factor_score +
		weights["drawdown"]*drawdown_score +
		weights["avg_profit"]*avg_profit_score +
		weights["total_profit"]*total_profit_score

	// handle extreme values
	if profitFactor < 1 || expectancy < 0 {
		total_score = -1
	}

	return total_score
}

// metric_score computes the score of a metric
// using logaritmic growth and exponential decay.
//
//   - Logarithmic Growth
//     log(1+value/baseline): Produces diminishing returns as values increase above the baseline.
//     This ensures very high values (e.g., Profit Factor = 5) contribute more but don't dominate excessively.
//
//   - Exponential Decay for Drawdown:
//     −log(1+value/baseline): Penalizes drawdowns proportionally.
//     Larger drawdowns have exponentially higher negative impacts.
func metric_score(value, baseline float64, positiveDirection bool, sensitivity float64) float64 {
	if positiveDirection {
		return sensitivity * math.Log(1+value/baseline)
	} else {
		return -sensitivity * math.Log(1+value/baseline)
	}
}
