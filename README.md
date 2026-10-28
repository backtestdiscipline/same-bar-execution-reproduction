# Same-Bar Execution Bias — Independent BTC Reproduction

A backtest can give a strategy credit for a price move before the information needed to make its decision was available. This repository reproduces the main Bitcoin example from a paper on that timing error. The independently written study compares the paper's corrected next-open ledger with its same-bar error, holding the signal and other assumptions fixed. It includes the code, frozen method, sealed result, and provenance records.

## Original research

Ronald Chan and Eric Lopes, [*Look-Ahead Bias Has No Deflator: Quantifying the Same-Bar Execution Error in Systematic Backtests*](https://ssrn.com/abstract=7654039), version dated October 21, 2026. The reproduced experiment is the BTC comparison in section 5.1, table 1, using ledgers A and F defined in section 3. This implementation is independent of the paper's authors; their participation or endorsement is not implied.

## Method and timing

The study uses 3,336 Binance BTCUSDT spot daily bars, August 17, 2017 through October 4, 2026 inclusive, in UTC. The long-or-flat signal is based on 20-day close momentum, normalized by a trailing 90-observation z-score using sample standard deviation. It is long when the z-score is strictly greater than 0.5 and flat during warmup. Both ledgers subtract 0.001 per absolute position change.

For a signal calculated using the close of day `t`:

```text
day t open             day t close             day t+1 open          day t+2 open
    |-----------------------|-------------------------|---------------------|
    F credits the position from here ---------------->|  (open t to open t+1)
                                                      A credits from here ->| (open t+1 to open t+2)
```

Ledger F credits the position with the open-to-open return starting before the close needed to calculate it. Ledger A first credits that position over the next open-to-open interval. The endpoint convention uses the first 3,335 open-to-open returns, a zero final return, and no forced liquidation. Sharpe includes warmup zeros and uses a 365.25-day annualization. The [frozen specification](records/same-bar-execution-reproduction/spec.json) defines the complete comparison and tolerances.

## Results

The sealed historical backtest results are:

| Metric | A: next open | F: same-bar error |
|---|---:|---:|
| Terminal wealth multiple | 14.4119× | 6,094.9403× |
| Annualized Sharpe | 1.0134 | 2.8110 |
| Maximum drawdown | 44.72% | 18.88% |
| Position changes | 232 | 232 |
| Frozen comparison tolerance met | Yes | Yes |

Parity means each wealth multiple is within 0.5% relative error and each Sharpe within 0.005 absolute error of the paper's rounded figures. These are simulated historical calculations, not realized returns or evidence that the same-bar performance is achievable. The difference illustrates temporal accounting in this example; it does not measure how often the defect occurs or establish a deployable strategy.

## Interpretation and limitations

The numerical comparison reproduces the two reported metrics under the frozen rules. The recorded semantic verification is deterministic replay: `study.Verify` uses the same core calculation as `study.Evaluate`, not an independent second implementation. Neither replay nor the recorded integrity check independently validates the paper's wider claims. The model does not establish real-world execution feasibility or cover all execution constraints.

The authors' original raw data and canonical replication package were unavailable. Binance history can change between retrieval dates. The exact source bytes used here are retained locally, not committed to this repository. The paper's results were known before the method was frozen; the charter was written after data acquisition, although the raw prices were reportedly not explored for returns before freezing. This is a frozen-method reproduction against known results, not an untouched prospective test. The generated handoff's broader “prospective or confirmatory study” label should be read with this qualification.

## Evidence and provenance

Run `20261027T104112Z-e234b6bd8583` records outcome `COMPLETE`, platform integrity `PASS`, and semantic replay `PASS` in the [handoff](handoffs/20261027T104112Z-e234b6bd8583.json). These are distinct claims; hashes establish byte identity and integrity, not scientific correctness.

- [Charter](charter.md) and [status](STATUS.md): scope, acquisition chronology, and interpretation.
- [Manifest](records/same-bar-execution-reproduction/manifest.json): source paths, acquisition details, sizes, and hashes.
- [Specification](records/same-bar-execution-reproduction/spec.json) and [freeze](records/same-bar-execution-reproduction/freeze.json): method, comparison rules, and frozen record bindings.
- [Sealed result](.research-platform/runs/20261027T104112Z-e234b6bd8583/artifacts/result.json) and [seal](.research-platform/runs/20261027T104112Z-e234b6bd8583/seal.json): recorded calculations and integrity bindings.

## Inspecting and verifying

The code and evidence metadata can be inspected directly on GitHub. Building and running tests additionally requires Go **1.27.2** and a local checkout of the separate Go module `research-platform`. The platform is not currently available as a public dependency. The original `go.mod` contains a machine-specific replacement and is part of the sealed study implementation identity; it is retained unchanged.

With compatible study and platform checkouts as siblings, create a temporary Go workspace *outside* the study tree to override that replacement without editing `go.mod`. From the study root:

```sh
workdir=$(mktemp -d)
export GOWORK="$workdir/go.work" GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off
go work init "$PWD" ../research-platform
go test ./...
go vet ./...
```

Full verification of the existing sealed run additionally needs the exact five source files listed in the manifest at their recorded `data/` paths, plus a compatible platform checkout. Their bytes must match the recorded sizes and SHA-256 hashes; a fresh Binance download may differ. These source files are not distributed here. With those prerequisites in place, run from the study root:

```sh
go run ./cmd/same-bar-execution-reproduction verify 20261027T104112Z-e234b6bd8583
```

This checks the existing seal and performs semantic replay; it does not create a new evaluation. The command reports whether the current clean study implementation matches the sealed identity. A locally available platform checkout at commit `b3cd893c916ef0e3ffef651b576462c9f232170e` has an embedded identity matching the seal and was used for a later successful verification. Availability of a compatible checkout and the exact sources is required for practical replay, so the public repository alone does not provide it.

## License and project

No software license is supplied for this repository. The paper's licensing does not automatically license this implementation.

This study is part of [Backtest Discipline](https://backtestdiscipline.com) research and reproducibility work.
