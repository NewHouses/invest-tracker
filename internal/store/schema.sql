CREATE TABLE IF NOT EXISTS assets (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    type        TEXT    NOT NULL CHECK (type IN ('accion','indice','copy_trading','fondo')),
    name        TEXT    NOT NULL,
    amount_usd  REAL    NOT NULL,
    month       INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year        INTEGER NOT NULL,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS transactions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id   INTEGER NOT NULL REFERENCES assets(id),
    amount_usd REAL    NOT NULL,
    month      INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year       INTEGER NOT NULL,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_transactions_asset ON transactions(asset_id);

CREATE TABLE IF NOT EXISTS monthly_results (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id   INTEGER NOT NULL REFERENCES assets(id),
    result_usd REAL    NOT NULL,
    month      INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year       INTEGER NOT NULL,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_monthly_results_asset ON monthly_results(asset_id);

CREATE TABLE IF NOT EXISTS dividends (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    amount_usd REAL    NOT NULL,
    month      INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year       INTEGER NOT NULL,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);
