package main

// BacktestResult is the main structure that holds the backtest results
type BacktestResult struct {
	Strategy map[string]Strategy `json:"strategy"`
}

// Strategy represents the backtest results for a single strategy
type Strategy struct {
	// General purpose information
	BacktestStart CustomTime `json:"backtest_start"`
	BacktestEnd   CustomTime `json:"backtest_end"`
	BacktestDays  int        `json:"backtest_days"`

	// Finance informations
	TotalTrades         int     `json:"total_trades"`
	StartingBalance     float64 `json:"starting_balance"`
	FinalBalance        float64 `json:"final_balance"`
	ProfitTotalAbs      float64 `json:"profit_total_abs"`
	ProfitTotal         float64 `json:"profit_total"`
	ProfitMean          float64 `json:"profit_mean"`
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

	// Trading information
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

	Trades []Trade `json:"trades"`

	// Config
	MaxOpenTrades int                `json:"max_open_trades"`
	MinimalROI    map[string]float64 `json:"minimal_roi"`
	StakeCurrency string             `json:"stake_currency"`
	Stoploss      float64            `json:"stoploss"`

	MinimalROISorted MinimalROISorted
}

// Trade represents a single trade
type Trade struct {
	ExitReason    string  `json:"exit_reason"`
	ProfitAbs     float64 `json:"profit_abs"`
	ProfitRatio   float64 `json:"profit_ratio"`
	IsOpen        bool    `json:"is_open"`
	TradeDuration int     `json:"trade_duration"`
}

// MinimalROISorted is a slice of MinimalROI
type MinimalROISorted []MinimalROI

// MinimalROI represents minimal_roi settings
type MinimalROI struct {
	Name  string
	Value float64
}

// StrategyReport represents exit reason reports of a strategy
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
