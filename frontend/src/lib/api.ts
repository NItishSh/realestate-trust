// Client API interface connecting to Go microservices.
// Falls back to mock memory data if microservices are offline.

export interface User {
  id: string;
  email: string;
  fullName: string;
  role: 'BUYER' | 'SELLER' | 'BROKER' | 'OFFICER' | 'ADMIN';
  created_at?: string;
  kycStatus?: 'PENDING' | 'APPROVED' | 'REJECTED';
  documentType?: string;
  documentReference?: string;
}

export interface Transaction {
  id: string;
  propertyId: string;
  buyerId: string;
  sellerId: string;
  totalAmount: number;
  status: 'DRAFT' | 'ESCROW' | 'FUNDED' | 'CLOSED' | 'CANCELLED';
  created_at?: string;
}

export interface EscrowAccount {
  id: string;
  transactionId: string;
  virtualAccountNumber: string;
  bankPartner: string;
  balance: number;
}

export interface Loan {
  id: string;
  transactionId: string;
  userId: string;
  lenderId?: string;
  requestedAmount: number;
  approvedAmount?: number;
  status: 'APPLIED' | 'UNDERWRITING' | 'APPROVED' | 'REJECTED' | 'DISBURSED';
}

export interface FractionalPool {
  id: string;
  propertyId: string;
  totalTokens: number;
  tokenPrice: number;
  tokensSold: number;
}

export interface LedgerBlock {
  index: number;
  timestamp: string;
  payload: string;
  previousHash: string;
  hash: string;
}

// Fallback Mock State
const mockUsers: User[] = [
  { id: "usr-1", email: "investor@gmail.com", fullName: "John Doe", role: "BUYER", kycStatus: "APPROVED", documentType: "PASSPORT", documentReference: "P74928" },
  { id: "usr-2", email: "seller@realestate.com", fullName: "Jane Smith", role: "SELLER", kycStatus: "APPROVED", documentType: "PASSPORT", documentReference: "P93011" },
];

const mockTransactions: Transaction[] = [
  { id: "tx-1", propertyId: "prop-101", buyerId: "usr-1", sellerId: "usr-2", totalAmount: 450000.00, status: "ESCROW" },
  { id: "tx-2", propertyId: "prop-102", buyerId: "usr-1", sellerId: "usr-2", totalAmount: 850000.00, status: "DRAFT" },
];

const mockEscrow: EscrowAccount[] = [
  { id: "esc-1", transactionId: "tx-1", virtualAccountNumber: "VA-TRUST-092182", bankPartner: "TRUSTEE BANK PACIFIC", balance: 150000.00 },
];

const mockLoans: Loan[] = [
  { id: "ln-1", transactionId: "tx-1", userId: "usr-1", requestedAmount: 300000.00, approvedAmount: 300000.00, status: "APPROVED" },
];

const mockPools: FractionalPool[] = [
  { id: "pool-1", propertyId: "prop-101", totalTokens: 1000, tokenPrice: 450.00, tokensSold: 320 },
  { id: "pool-2", propertyId: "prop-103", totalTokens: 5000, tokenPrice: 170.00, tokensSold: 4800 },
];

const mockBlocks: LedgerBlock[] = [
  { index: 0, timestamp: "2026-07-12T05:00:00Z", payload: "Genesis Block", previousHash: "0", hash: "8523ef33a921d7494e09f61b9a5286c4ad4b8efc8f38d38ca5d762957eef53ab" },
  { index: 1, timestamp: "2026-07-12T05:05:00Z", payload: "Transaction Created: tx-1", previousHash: "8523ef33a921d7494e09f61b9a5286c4ad4b8efc8f38d38ca5d762957eef53ab", hash: "9a2f7c03ba88e7b1a20b08051a2d5dbd392fe8e90632ef3914a2a11b643a6d71" },
];

const API_BASE = "http://localhost";
const ports = {
  transactions: 8080,
  identity: 8081,
  financing: 8082,
  tokenization: 8083,
  ledger: 8084
};

async function safeFetch<T>(url: string, options: RequestInit, fallback: T): Promise<T> {
  try {
    // Inject Authorization header if token exists
    const token = typeof window !== 'undefined' ? localStorage.getItem('jwt_token') : null;
    const headers = new Headers(options.headers || {});
    if (token) {
      headers.set('Authorization', `Bearer ${token}`);
    }

    const res = await fetch(url, {
      ...options,
      headers,
      signal: AbortSignal.timeout(1500)
    });

    if (res.status === 401) {
       console.error("Unauthorized request");
       // Could redirect to login here
    }

    if (!res.ok) throw new Error("API error status");
    return await res.json() as T;
  } catch (e) {
    // Graceful fallback to mock data
    return fallback;
  }
}

export const api = {
  // Auth
  login: async (email: string) => {
    try {
      const res = await fetch(`${API_BASE}:${ports.identity}/api/v1/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email })
      });
      if (res.ok) {
        const data = await res.json();
        if (data.token && typeof window !== 'undefined') {
          localStorage.setItem('jwt_token', data.token);
        }
        return data;
      }
    } catch (e) {
      console.error("Login failed", e);
    }
  },

  // 1. Identity Service API mappings
  getUsers: () => safeFetch<User[]>(`${API_BASE}:${ports.identity}/api/v1/users`, { method: "GET" }, mockUsers),
  registerUser: (user: Partial<User>) => safeFetch<User>(`${API_BASE}:${ports.identity}/api/v1/users`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(user)
  }, { id: "usr-" + Math.random().toString(36).substr(2, 5), ...user, kycStatus: "PENDING" } as User),
  submitKYC: (userId: string, kyc: { documentType: string, documentReference: string }) => safeFetch<{ status: string }>(
    `${API_BASE}:${ports.identity}/api/v1/users/${userId}/kyc`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(kyc)
    },
    { status: "ACCEPTED" }
  ),

  // 2. Transaction Manager API mappings
  getTransactions: () => safeFetch<Transaction[]>(`${API_BASE}:${ports.transactions}/api/v1/transactions`, { method: "GET" }, mockTransactions),
  createTransaction: (tx: Partial<Transaction>) => safeFetch<Transaction>(`${API_BASE}:${ports.transactions}/api/v1/transactions`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(tx)
  }, { id: "tx-" + Math.random().toString(36).substr(2, 5), status: "DRAFT", ...tx } as Transaction),
  getEscrow: (txId: string) => safeFetch<EscrowAccount>(`${API_BASE}:${ports.transactions}/api/v1/transactions/${txId}/escrow`, { method: "GET" }, mockEscrow[0]),
  fundEscrow: (txId: string, amount: number) => safeFetch<{ status: string }>(`${API_BASE}:${ports.transactions}/api/v1/transactions/${txId}/escrow/fund`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ amount })
  }, { status: "funding_received" }),

  // 3. Financing Engine API mappings
  getLoans: () => safeFetch<Loan[]>(`${API_BASE}:${ports.financing}/api/v1/loans`, { method: "GET" }, mockLoans),
  applyLoan: (loan: Partial<Loan>) => safeFetch<Loan>(`${API_BASE}:${ports.financing}/api/v1/loans`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(loan)
  }, { id: "ln-" + Math.random().toString(36).substr(2, 5), status: "APPLIED", ...loan } as Loan),

  // 4. Tokenization Engine API mappings
  getPools: () => safeFetch<FractionalPool[]>(`${API_BASE}:${ports.tokenization}/api/v1/pools`, { method: "GET" }, mockPools),
  createPool: (pool: Partial<FractionalPool>) => safeFetch<FractionalPool>(`${API_BASE}:${ports.tokenization}/api/v1/pools`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(pool)
  }, { id: "pool-" + Math.random().toString(36).substr(2, 5), tokensSold: 0, ...pool } as FractionalPool),

  // 5. Ledger Service API mappings
  getLedger: () => safeFetch<LedgerBlock[]>(`${API_BASE}:${ports.ledger}/api/v1/logs`, { method: "GET" }, mockBlocks),
  writeLedgerLog: (log: string) => safeFetch<LedgerBlock>(`${API_BASE}:${ports.ledger}/api/v1/logs`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ payload: log })
  }, {
    index: mockBlocks.length,
    timestamp: new Date().toISOString(),
    payload: log,
    previousHash: mockBlocks[mockBlocks.length - 1].hash,
    hash: Math.random().toString(36).substr(2, 10) + "mockhash"
  })
};
