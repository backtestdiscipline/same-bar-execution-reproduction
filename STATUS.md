# Status

- Study: same-bar-execution-reproduction; frozen reproduction complete.
- Source bytes: user-supplied PDF and four Binance API pages acquired 2026-10-27, retained under ignored `data/`; local SHA-256 identities are in the manifest. The authors' original data and replication package were not available. The paper was read and reported results known before freeze; raw price results were not explored before freeze. Charter was written after acquisition.
- Run `20261027T104112Z-e234b6bd8583`: COMPLETE; integrity PASS and study semantic replay PASS. Semantic replay repeats the implementation and is not an independent implementation.
- BTC 3,336 bars: A wealth 14.41189155, Sharpe 1.01344745, drawdown 44.7236%; F wealth 6094.94030863, Sharpe 2.81101172, drawdown 18.8839%. Both meet frozen comparison tolerances. Each has 232 position changes.
- A's signal from close t is first credited from open t+1; F credits the same close-dependent signal from open t. This establishes the logical timing defect, not execution feasibility or the frequency of this defect in practice.
- No platform extension was attempted; timestamps and availability are scientific semantics owned by the study. Backup of this new repository and its ignored sources is outstanding; existing platform backup script covers a fixed list of other studies.
