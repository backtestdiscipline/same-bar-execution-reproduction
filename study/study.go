// Package study evaluates the paper's two BTC execution ledgers.
package study

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"research-platform/seal"
	"sort"
	"strconv"
	"strings"
)

type Evaluation struct {
	Outcome seal.TerminalOutcome
	Payload []byte
}
type Verification struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}
type bar struct {
	t           int64
	open, close float64
}
type metrics struct {
	Wealth              float64 `json:"wealth"`
	Sharpe              float64 `json:"sharpe"`
	MaxDrawdown         float64 `json:"max_drawdown"`
	Flips               int     `json:"flips"`
	WealthRelativeError float64 `json:"wealth_relative_error"`
	SharpeAbsoluteError float64 `json:"sharpe_absolute_error"`
	Parity              bool    `json:"parity"`
}
type result struct {
	Bars         int     `json:"bars"`
	ValidSignals int     `json:"valid_signals"`
	Clean        metrics `json:"clean_a"`
	SameBar      metrics `json:"same_bar_f"`
	Explanation  string  `json:"timing_explanation"`
}

func Evaluate(in Inputs) (Evaluation, error) {
	b, err := load(in)
	if err != nil {
		return Evaluation{}, err
	}
	r, err := calculate(b)
	if err != nil {
		return Evaluation{}, err
	}
	out, err := json.Marshal(r)
	return Evaluation{seal.Complete, out}, err
}
func Verify(in Inputs, payload []byte) (Verification, error) {
	b, err := load(in)
	if err != nil {
		return Verification{}, err
	}
	r, err := calculate(b)
	if err != nil {
		return Verification{}, err
	}
	expected, err := json.Marshal(r)
	if err != nil {
		return Verification{}, err
	}
	if !bytes.Equal(expected, payload) {
		return Verification{"FAIL", "Result does not match deterministic replay."}, nil
	}
	return Verification{"PASS", "Source validation and deterministic replay match; this does not independently establish the paper's claim."}, nil
}
func load(in Inputs) ([]bar, error) {
	var refs []seal.ArtifactRef
	for _, r := range in.Sources {
		if strings.Contains(r.Path, "btcusdt-daily-page-") {
			refs = append(refs, r)
		}
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].Path < refs[j].Path })
	if len(refs) != 4 {
		return nil, fmt.Errorf("expected four source pages")
	}
	var bars []bar
	for _, ref := range refs {
		f, err := in.OpenSource(ref)
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(f, 1<<20))
		ce := f.Close()
		if err != nil {
			return nil, err
		}
		if ce != nil {
			return nil, ce
		}
		var rows [][]json.RawMessage
		if err = json.Unmarshal(data, &rows); err != nil {
			return nil, err
		}
		for _, row := range rows {
			if len(row) != 12 {
				return nil, fmt.Errorf("invalid kline")
			}
			var t int64
			var o, c string
			if err = json.Unmarshal(row[0], &t); err != nil {
				return nil, err
			}
			if err = json.Unmarshal(row[1], &o); err != nil {
				return nil, err
			}
			if err = json.Unmarshal(row[4], &c); err != nil {
				return nil, err
			}
			op, e := strconv.ParseFloat(o, 64)
			if e != nil {
				return nil, e
			}
			cl, e := strconv.ParseFloat(c, 64)
			if e != nil {
				return nil, e
			}
			if op <= 0 || cl <= 0 || math.IsNaN(op) || math.IsInf(op, 0) || math.IsNaN(cl) || math.IsInf(cl, 0) {
				return nil, fmt.Errorf("invalid price")
			}
			bars = append(bars, bar{t, op, cl})
		}
	}
	if len(bars) != 3336 {
		return nil, fmt.Errorf("got %d bars", len(bars))
	}
	for i, b := range bars {
		if b.t != 1502928000000+int64(i)*86400000 {
			return nil, fmt.Errorf("invalid timestamp at %d", i)
		}
	}
	return bars, nil
}
func calculate(b []bar) (result, error) {
	if len(b) < 112 {
		return result{}, fmt.Errorf("insufficient bars")
	}
	n := len(b)
	p := make([]float64, n)
	mom := make([]float64, n)
	valid := 0
	for i := 20; i < n; i++ {
		mom[i] = b[i].close/b[i-20].close - 1
		if i < 109 {
			continue
		}
		mean := 0.0
		for j := i - 89; j <= i; j++ {
			mean += mom[j]
		}
		mean /= 90
		variance := 0.0
		for j := i - 89; j <= i; j++ {
			d := mom[j] - mean
			variance += d * d
		}
		sd := math.Sqrt(variance / 89)
		if sd > 0 && (mom[i]-mean)/sd > 0.5 {
			p[i] = 1
		}
		valid++
	}
	a := make([]float64, n)
	f := make([]float64, n)
	flipsA, flipsF := 0, 0
	for i := 0; i < n; i++ {
		ret := 0.0
		if i+1 < n {
			ret = b[i+1].open/b[i].open - 1
		}
		prev, prev2 := 0.0, 0.0
		if i >= 1 {
			prev = p[i-1]
		}
		if i >= 2 {
			prev2 = p[i-2]
		}
		da, df := math.Abs(prev-prev2), math.Abs(p[i]-prev)
		if da > 0 {
			flipsA++
		}
		if df > 0 {
			flipsF++
		}
		a[i] = prev*ret - 0.001*da
		f[i] = p[i]*ret - 0.001*df
	}
	return result{n, valid, summarize(a, flipsA, 14.41, 1.01), summarize(f, flipsF, 6094.94, 2.81), "Signal p[t] depends on close[t], available no earlier than that close. F credits p[t] from open[t], before that close; A credits p[t] from open[t+1]. Close availability does not prove executable fill at that instant."}, nil
}
func summarize(r []float64, flips int, targetW, targetS float64) metrics {
	wealth, peak, dd, sum := 1.0, 1.0, 0.0, 0.0
	for _, v := range r {
		wealth *= 1 + v
		if wealth > peak {
			peak = wealth
		}
		if d := 1 - wealth/peak; d > dd {
			dd = d
		}
		sum += v
	}
	mean := sum / float64(len(r))
	variance := 0.0
	for _, v := range r {
		d := v - mean
		variance += d * d
	}
	sharpe := mean / math.Sqrt(variance/float64(len(r)-1)) * math.Sqrt(365.25)
	we := math.Abs(wealth/targetW - 1)
	se := math.Abs(sharpe - targetS)
	return metrics{wealth, sharpe, dd, flips, we, se, we <= 0.005 && se <= 0.005}
}
