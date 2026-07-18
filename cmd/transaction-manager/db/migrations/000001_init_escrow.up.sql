CREATE TYPE transaction_status AS ENUM ('DRAFT', 'ESCROW', 'FUNDED', 'CLOSED', 'CANCELLED');

CREATE TABLE transactions (
    id VARCHAR(255) PRIMARY KEY,
    property_id VARCHAR(255) NOT NULL,
    buyer_id VARCHAR(255) NOT NULL,
    seller_id VARCHAR(255) NOT NULL,
    total_amount DECIMAL(15, 2) NOT NULL,
    status transaction_status NOT NULL DEFAULT 'DRAFT',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE escrow_accounts (
    id VARCHAR(255) PRIMARY KEY,
    transaction_id VARCHAR(255) NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    virtual_account_number VARCHAR(50) UNIQUE NOT NULL,
    bank_partner VARCHAR(100) NOT NULL,
    balance DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_escrow_accounts_va ON escrow_accounts(virtual_account_number);
