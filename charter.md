# Study charter

- Question: Can the paper's main BTC momentum comparison and timing explanation be independently reproduced from Binance daily spot data?
- Kind: reproduction.
- Target: Section 5.1, ledgers A and F, including wealth, Sharpe, and the position/return timing relationship.
- Non-goals: Trading advice, prevalence estimates, ETH, equities, parameter grids, the authors' gate, and claims of exact execution realism.
- Sources: the user-supplied paper and public Binance API daily BTCUSDT klines. Retain source bytes locally, do not commit them. No source bytes or excerpts are to be sent to another AI service.
- Acquisition: API data were acquired on 2026-10-27 after initial paper review and before this charter was written. The raw bars were not inspected for returns before method freeze; acquisition counts and dates alone were checked. The paper's published results are already known, so this is not an independent prospective test.
