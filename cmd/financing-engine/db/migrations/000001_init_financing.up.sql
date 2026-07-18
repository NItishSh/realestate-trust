CREATE TYPE loan_status AS ENUM ('APPLIED', 'UNDERWRITING', 'APPROVED', 'REJECTED', 'DISBURSED');
CREATE TYPE disbursement_status AS ENUM ('PENDING', 'COMPLETED', 'FAILED');

CREATE TABLE loans (
    id VARCHAR(255) PRIMARY KEY,
    transaction_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    lender_id VARCHAR(255),
    requested_amount DECIMAL(15, 2) NOT NULL,
    approved_amount DECIMAL(15, 2),
    status loan_status NOT NULL DEFAULT 'APPLIED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE disbursements (
    id VARCHAR(255) PRIMARY KEY,
    loan_id VARCHAR(255) NOT NULL REFERENCES loans(id) ON DELETE CASCADE,
    virtual_account_number VARCHAR(50) NOT NULL,
    disbursed_amount DECIMAL(15, 2) NOT NULL,
    transaction_reference VARCHAR(100) UNIQUE NOT NULL,
    status disbursement_status NOT NULL DEFAULT 'PENDING',
    disbursed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_loans_transaction ON loans(transaction_id);
CREATE INDEX idx_disbursements_loan ON disbursements(loan_id);
