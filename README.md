# freqtrade backtest analyzer

CLI tool to analyze Freqtrade backtesting results.

see https://www.freqtrade.io/en/stable/backtesting/

## Example

```
$ go run -v ./main.go ../user_data/backtest_results/backtest-result-2023-02-09_21-32-52.json
2023/02/09 21:33:52 > start
2023/02/09 21:33:52 > opened ../user_data/backtest_results/backtest-result-2023-02-09_21-32-52.json
2023/02/09 21:33:52 > read 3942797 bytes from ../user_data/backtest_results/backtest-result-2023-02-09_21-32-52.json
2023/02/09 21:33:52 > loaded backtest result
2023/02/09 21:33:52 > processing 4335 trades
2023/02/09 21:33:52 > WARNING 2774 trades have duration=0
+--------------------+-------+--------------+------------+--------------+--------------+---------------------+
| EXIT REASON        | EXITS | AVG PROFIT % | TOT PROFIT | TOT PROFIT % | AVG DURATION |     STDDEV DURATION |
+--------------------+-------+--------------+------------+--------------+--------------+---------------------+
| roi                |  3221 |         2.44 |    7845.56 |        57.49 |      15h8m0s |  26h59m53.28816828s |
| roi 0:0.025        |  1740 |         2.46 |    4285.92 |        54.63 |     17h51m0s | 34h25m50.906770766s |
| roi art:0.02       |  1474 |         2.41 |    3547.99 |        45.22 |     12h45m0s |  18h25m9.537500487s |
| exit_signal        |  1067 |        -4.99 |   -5319.57 |       -38.98 |      30h9m0s |  31h13m4.990168847s |
| force_exit         |    31 |        -2.78 |     -86.15 |        -0.63 |     24h52m0s |  5h59m34.824037708s |
| trailing_stop_loss |    16 |       -24.75 |    -396.01 |        -2.90 |      26h3m0s |  32h8m47.308790967s |
| roi art:0.01       |     7 |         1.66 |      11.65 |         0.15 |     92h10m0s |   69h55m57.3117111s |
+--------------------+-------+--------------+------------+--------------+--------------+---------------------+
+------------------------+-------------------------------+
| METRIC                 | VALUE                         |
+------------------------+-------------------------------+
| Backtest from          | 2022-11-08 18:00:00 +0000 UTC |
| Backtest to            | 2023-02-06 19:00:00 +0000 UTC |
| Max open trades        | 117                           |
| Minimal ROI            | 0:0.025                       |
|                        |                               |
| Strategy               | FSupertrendStrategy           |
| Total/Daily Avg Trades | 4335 / 48.17                  |
| Starting balance       | 1000.000 USDT                 |
| Final balance          | 3043.831 USDT                 |
| Absolute profit        | 2043.831 USDT                 |
| Total profit %         | 204.38%                       |
| CAGR %                 | 9031.43%                      |
| Sortino                | 106.82                        |
| Sharpe                 | 110.40                        |
| Calmar                 | 160.76                        |
| Profit factor          | 1.35                          |
| Expectancy             | 0.09                          |
| Trades per day         | 48.17                         |
| Avg. daily profit %    | 2.27                          |
| Avg. stake amount      | 97.483 USDT                   |
| Total trade volume     | 422587.413 USDT               |
|                        |                               |
| Long / Short           | 2216 2119                     |
| Total profit Long %    | 82.52%                        |
| Total profit Short %   | 121.86%                       |
| Absolute profit Long   | 825.187 USDT                  |
| Absolute profit Short  | 1218.644 USDT                 |
+------------------------+-------------------------------+
```

This output try to mimic the native output from `freqtrade backtesting`. But some fields are missing yet.

NOTE: exit results with reason roi, are threated in a special way. any `roi *` section is considered a subsection of the main `roi`, meaning that reported values account against the main `roi` section. i.e. `roi 0:0.025` has `54.63%` total profit over all the `roi` exits. `roi art*` are artificial categories added for better visibility. This is helpfull to pin down which `roi` setting is relevant or not.
