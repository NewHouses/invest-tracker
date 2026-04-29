CREATE TABLE IF NOT EXISTS investments (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    type        TEXT    NOT NULL CHECK (type IN ('accion','indice','copy_trading','fondo')),
    name        TEXT    NOT NULL,
    amount_usd  REAL    NOT NULL,
    month       INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year        INTEGER NOT NULL,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS transactions (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    investment_id INTEGER NOT NULL REFERENCES investments(id),
    amount_usd    REAL    NOT NULL,
    month         INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year          INTEGER NOT NULL,
    created_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_transactions_investment ON transactions(investment_id);
