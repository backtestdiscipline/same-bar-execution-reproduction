package study

import (
	"math"
	"testing"
)

func TestSummarize(t *testing.T) {
	m := summarize([]float64{0.1, -0.1, 0}, 2, 0.99, 0)
	if math.Abs(m.Wealth-0.99) > 1e-12 || math.Abs(m.MaxDrawdown-0.1) > 1e-12 || math.Abs(m.Sharpe) > 1e-12 || !m.Parity {
		t.Fatalf("unexpected metrics: %+v", m)
	}
}
func TestWarmupAndTiming(t *testing.T) {
	bars := make([]bar, 112)
	for i := range bars {
		bars[i] = bar{int64(i), 100, 100}
	}
	bars[109].close = 120
	bars[111].open = 110
	r, err := calculate(bars)
	if err != nil {
		t.Fatal(err)
	}
	if r.ValidSignals != 3 || r.SameBar.Flips != 2 || r.Clean.Flips != 2 {
		t.Fatalf("unexpected signals: %+v", r)
	}
	if math.Abs(r.SameBar.Wealth-0.998001) > 1e-12 || math.Abs(r.Clean.Wealth-1.097901) > 1e-12 {
		t.Fatalf("unexpected timing: %+v", r)
	}
}
