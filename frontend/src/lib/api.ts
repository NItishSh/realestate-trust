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
  password?: string;
}

export interface Property {
  id: string;
  address: string;
  description: string;
  value: number;
  thumbnail: string;
  ownerId: string;
  documents?: string[];
  createdAt?: string;
  titleInsuranceStatus?: 'UNINSURED' | 'INSURED';
  titleInsurancePolicy?: string;
  titleInsuranceCompany?: string;
  titleInsuranceVerifiedAt?: string;
  sqft?: number;
  bedrooms?: number;
  bathrooms?: number;
  yearBuilt?: number;
  propertyType?: string;
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

export interface Feedback {
  id: string;
  userId: string;
  message: string;
  category: string;
  rating: number;
  createdAt?: string;
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

const mockFeedback: Feedback[] = [
  { id: "fb-1", userId: "usr-1", message: "Great application! Really easy to track real estate transactions.", category: "general", rating: 5, createdAt: "2026-07-14T12:00:00Z" }
];

const mockProperties: Property[] = [
  { id: "prop-101", address: "123 Ocean View Dr, Malibu", description: "Luxury beachfront villa", value: 4500000, thumbnail: "https://images.unsplash.com/photo-1512917774080-9991f1c4c750?w=800", ownerId: "usr-priya.sharma@realestate.in", titleInsuranceStatus: "UNINSURED", documents: ["Deed of Trust", "Property Inspection Report", "Title Insurance"], sqft: 3200, bedrooms: 4, bathrooms: 4, yearBuilt: 2020, propertyType: "Residential" }
];



const API_BASE = typeof window !== 'undefined' ? "" : "http://istio-ingress.istio-system.svc.cluster.local";

let isRefreshing = false;
let refreshSubscribers: ((token: string) => void)[] = [];

function subscribeTokenRefresh(cb: (token: string) => void) {
  refreshSubscribers.push(cb);
}

function onRefreshed(token: string) {
  refreshSubscribers.forEach(cb => cb(token));
  refreshSubscribers = [];
}

function getCorrelationID(): string {
  if (typeof window !== 'undefined' && window.crypto && window.crypto.randomUUID) {
    return window.crypto.randomUUID();
  }
  return Math.random().toString(36).substring(2, 15);
}

async function safeFetch<T>(url: string, options: RequestInit, fallback: T): Promise<T> {
  try {
    const token = typeof window !== 'undefined' ? localStorage.getItem('jwt_token') : null;
    const headers = new Headers(options.headers || {});
    if (token) {
      headers.set('Authorization', `Bearer ${token}`);
    }
    headers.set('X-Correlation-ID', getCorrelationID());

    let res = await fetch(url, {
      ...options,
      headers,
      signal: AbortSignal.timeout(1500)
    });

    if (res.status === 401) {
      console.warn("Access token expired. Attempting refresh...");
      const refreshToken = typeof window !== 'undefined' ? localStorage.getItem('refresh_token') : null;

      if (!refreshToken) {
        if (typeof window !== 'undefined' && window.location.pathname !== '/login' && window.location.pathname !== '/signup') {
          window.location.href = '/login';
        }
        throw new Error("Unauthorized");
      }

      if (!isRefreshing) {
        isRefreshing = true;
        try {
          const refreshHeaders = new Headers();
          refreshHeaders.set('Content-Type', 'application/json');
          refreshHeaders.set('X-Correlation-ID', getCorrelationID());

          const refreshRes = await fetch(`${API_BASE}/api/v1/refresh`, {
            method: "POST",
            headers: refreshHeaders,
            body: JSON.stringify({ refreshToken })
          });

          if (refreshRes.ok) {
            const data = await refreshRes.json();
            if (data.token && data.refreshToken) {
              localStorage.setItem('jwt_token', data.token);
              localStorage.setItem('refresh_token', data.refreshToken);
              isRefreshing = false;
              onRefreshed(data.token);
            } else {
              throw new Error("Invalid refresh response");
            }
          } else {
            throw new Error("Refresh failed");
          }
        } catch (err) {
          isRefreshing = false;
          localStorage.removeItem('jwt_token');
          localStorage.removeItem('refresh_token');
          localStorage.removeItem('user_email');
          if (typeof window !== 'undefined') {
            window.location.href = '/login';
          }
          throw new Error("Unauthorized");
        }
      }

      const newToken = await new Promise<string>((resolve) => {
        subscribeTokenRefresh(token => resolve(token));
      });

      headers.set('Authorization', `Bearer ${newToken}`);
      res = await fetch(url, {
        ...options,
        headers,
        signal: AbortSignal.timeout(1500)
      });
    }

    if (!res.ok) {
      throw new Error(`HTTP error! status: ${res.status}`);
    }
    return await res.json() as T;
  } catch (error) {
    console.error("API call failed, returning fallback mock data:", error);
    return fallback;
  }
}

export const api = {
  getCurrentUser: async (): Promise<User | null> => {
    if (typeof window === 'undefined') return null;
    const email = localStorage.getItem('user_email');
    if (!email) return null;
    try {
      const users = await api.getUsers();
      return users.find(u => u.email === email) || null;
    } catch {
      return null;
    }
  },

  // auth endpoint overrides
  login: async (email: string, password?: string) => {
    const headers = new Headers();
    headers.set('Content-Type', 'application/json');
    headers.set('X-Correlation-ID', getCorrelationID());

    const res = await fetch(`${API_BASE}/api/v1/login`, {
      method: "POST",
      headers,
      body: JSON.stringify({ email, password })
    });
    if (!res.ok) {
      throw new Error("Login failed");
    }
    return await res.json();
  },

  logout: async () => {
    try {
      const headers = new Headers();
      headers.set('Content-Type', 'application/json');
      headers.set('X-Correlation-ID', getCorrelationID());

      await fetch(`${API_BASE}/api/v1/logout`, {
        method: "POST",
        headers
      });
    } finally {
      if (typeof window !== 'undefined') {
        localStorage.removeItem('jwt_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('user_email');
        window.location.href = '/login';
      }
    }
  },

  // 1. Identity Service API mappings
  getUsers: () => safeFetch<User[]>(`${API_BASE}/api/v1/users`, { method: "GET" }, mockUsers),
  registerUser: (user: Partial<User>) => safeFetch<User>(`${API_BASE}/api/v1/users`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(user)
  }, { id: "usr-" + Math.random().toString(36).substr(2, 5), ...user, kycStatus: "PENDING" } as User),
  submitKYC: (userId: string, kyc: { documentType: string, documentReference: string }) => safeFetch<{ status: string }>(
    `${API_BASE}/api/v1/users/${userId}/kyc`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(kyc)
    },
    { status: "ACCEPTED" }
  ),

  // 2. Transaction Manager API mappings
  getTransactions: () => safeFetch<Transaction[]>(`${API_BASE}/api/v1/transactions`, { method: "GET" }, mockTransactions),
  createTransaction: (tx: Partial<Transaction>) => safeFetch<Transaction>(`${API_BASE}/api/v1/transactions`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(tx)
  }, { id: "tx-" + Math.random().toString(36).substr(2, 5), status: "DRAFT", ...tx } as Transaction),
  getEscrow: (txId: string) => safeFetch<EscrowAccount>(`${API_BASE}/api/v1/transactions/${txId}/escrow`, { method: "GET" }, mockEscrow[0]),
  fundEscrow: (txId: string, amount: number) => safeFetch<{ status: string }>(`${API_BASE}/api/v1/transactions/${txId}/escrow/fund`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ amount })
  }, { status: "funding_received" }),

  // 3. Financing Engine API mappings
  getLoans: () => safeFetch<Loan[]>(`${API_BASE}/api/v1/loans`, { method: "GET" }, mockLoans),
  applyLoan: (loan: Partial<Loan>) => safeFetch<Loan>(`${API_BASE}/api/v1/loans`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(loan)
  }, { id: "ln-" + Math.random().toString(36).substr(2, 5), status: "APPLIED", ...loan } as Loan),

  // 4. Tokenization Engine API mappings
  getPools: () => safeFetch<FractionalPool[]>(`${API_BASE}/api/v1/pools`, { method: "GET" }, mockPools),
  createPool: (pool: Partial<FractionalPool>) => safeFetch<FractionalPool>(`${API_BASE}/api/v1/pools`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(pool)
  }, { id: "pool-" + Math.random().toString(36).substr(2, 5), tokensSold: 0, ...pool } as FractionalPool),

  // 5. Ledger Service API mappings
  getLedger: () => safeFetch<LedgerBlock[]>(`${API_BASE}/api/v1/logs`, { method: "GET" }, mockBlocks),
  writeLedgerLog: (log: string) => safeFetch<LedgerBlock>(`${API_BASE}/api/v1/logs`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ payload: log })
  }, {
    index: mockBlocks.length,
    timestamp: new Date().toISOString(),
    payload: log,
    previousHash: mockBlocks[mockBlocks.length - 1].hash,
    hash: Math.random().toString(36).substr(2, 10) + "mockhash"
  }),

  // 6. Property Registry API mappings
  getProperties: () => safeFetch<Property[]>(`${API_BASE}/api/v1/properties`, { method: "GET" }, mockProperties),
  getProperty: (id: string) => safeFetch<Property>(`${API_BASE}/api/v1/properties/${id}`, { method: "GET" }, mockProperties[0]),
  verifyTitleInsurance: (id: string, insurance: { company: string; policy: string }) => safeFetch<Property>(`${API_BASE}/api/v1/properties/${id}/insurance/verify`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(insurance)
  }, { ...mockProperties[0], titleInsuranceStatus: "INSURED", titleInsuranceCompany: insurance.company, titleInsurancePolicy: insurance.policy, titleInsuranceVerifiedAt: new Date().toISOString() } as Property),

  updatePropertyDetails: (id: string, details: { sqft: number; bedrooms: number; bathrooms: number; yearBuilt: number; propertyType: string }) => safeFetch<Property>(`${API_BASE}/api/v1/properties/${id}/details`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(details)
  }, { ...mockProperties[0], ...details } as Property),


  // 7. Feedback Service API mappings
  submitFeedback: (feedback: { message: string; category: string; rating: number }) => safeFetch<Feedback>(`${API_BASE}/api/v1/feedback`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(feedback)
  }, { id: "fb-" + Math.random().toString(36).substr(2, 5), userId: "usr-1", ...feedback, createdAt: new Date().toISOString() } as Feedback),

  getFeedback: () => safeFetch<Feedback[]>(`${API_BASE}/api/v1/feedback`, { method: "GET" }, mockFeedback)
};
