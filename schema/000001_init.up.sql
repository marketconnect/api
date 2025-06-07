CREATE TABLE IF NOT EXISTS api_key_balances (
    api_key TEXT PRIMARY KEY,
    balance INT NOT NULL
);

CREATE TABLE IF NOT EXISTS token_costs (
    token_type TEXT PRIMARY KEY,
    cost INT NOT NULL
);
