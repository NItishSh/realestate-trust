CREATE TABLE fractional_pools (
    id VARCHAR(255) PRIMARY KEY,
    property_id VARCHAR(255) NOT NULL,
    total_tokens BIGINT NOT NULL,
    tokens_sold BIGINT NOT NULL DEFAULT 0,
    token_price DECIMAL(15, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE fractional_holdings (
    id VARCHAR(255) PRIMARY KEY,
    pool_id VARCHAR(255) NOT NULL REFERENCES fractional_pools(id) ON DELETE CASCADE,
    investor_id VARCHAR(255) NOT NULL,
    token_count BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
