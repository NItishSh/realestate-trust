import { create } from 'zustand';
import { api, User, Transaction, Loan, FractionalPool, LedgerBlock } from './api';

interface State {
  users: User[];
  transactions: Transaction[];
  loans: Loan[];
  pools: FractionalPool[];
  ledger: LedgerBlock[];
  isLoading: boolean;
  currentUser: User | null;
  activeTxId: string | null;

  fetchInitialData: () => Promise<void>;
  setCurrentUser: (user: User) => void;
  registerUser: (user: Partial<User>) => Promise<void>;
  submitKYC: (userId: string, docType: string, docRef: string) => Promise<void>;
  createTransaction: (tx: Partial<Transaction>) => Promise<void>;
  fundEscrow: (txId: string, amount: number) => Promise<void>;
  applyLoan: (loan: Partial<Loan>) => Promise<void>;
  buyTokens: (poolId: string, count: number) => Promise<void>;
  createPool: (pool: Partial<FractionalPool>) => Promise<void>;
  setActiveTxId: (id: string | null) => void;
  login: (email: string, password?: string) => Promise<void>;
  logout: () => void;
  logAction: (message: string) => Promise<void>;
}

export const useStore = create<State>((set, get) => ({
  users: [],
  transactions: [],
  loans: [],
  pools: [],
  ledger: [],
  isLoading: false,
  currentUser: null,
  activeTxId: null,

  fetchInitialData: async () => {
    set({ isLoading: true });
    try {
      const [users, transactions, loans, pools, ledger] = await Promise.all([
        api.getUsers(),
        api.getTransactions(),
        api.getLoans(),
        api.getPools(),
        api.getLedger()
      ]);
      const currentEmail = typeof window !== 'undefined' ? localStorage.getItem('user_email') : null;
      const currentUser = users.find(u => u.email === currentEmail) || users[0] || null;
      set({
        users,
        transactions,
        loans,
        pools,
        ledger,
        currentUser,
        isLoading: false
      });
    } catch (e) {
      set({ isLoading: false });
    }
  },

  setCurrentUser: (user) => set({ currentUser: user }),

  registerUser: async (userForm) => {
    const newUser = await api.registerUser(userForm);
    set((state) => ({ users: [...state.users, newUser], currentUser: newUser }));
    if (typeof window !== 'undefined') localStorage.setItem('user_email', newUser.email);
  },

  logAction: async (message: string) => {
    const log = await api.writeLedgerLog(message);
    set((state) => ({ ledger: [...state.ledger, log] }));
  },

  submitKYC: async (userId, docType, docRef) => {
    await api.submitKYC(userId, { documentType: docType, documentReference: docRef });
    set((state) => ({
      users: state.users.map((u) => u.id === userId ? { ...u, kycStatus: 'PENDING', documentType: docType, documentReference: docRef } : u),
      currentUser: state.currentUser && state.currentUser.id === userId ?
        { ...state.currentUser, kycStatus: 'PENDING', documentType: docType, documentReference: docRef } : state.currentUser
    }));
    // Log to ledger
    const log = await api.writeLedgerLog(`KYC Submitted for User ${userId} (${docType})`);
    set((state) => ({ ledger: [...state.ledger, log] }));
  },

  createTransaction: async (txForm) => {
    const newTx = await api.createTransaction(txForm);
    set((state) => ({ transactions: [...state.transactions, newTx] }));
    const log = await api.writeLedgerLog(`Transaction Created: ${newTx.id} - ${newTx.totalAmount} INR`);
    set((state) => ({ ledger: [...state.ledger, log] }));
  },

  fundEscrow: async (txId, amount) => {
    await api.fundEscrow(txId, amount);
    set((state) => ({
      transactions: state.transactions.map(t =>
        t.id === txId ? { ...t, status: 'FUNDED' } : t
      )
    }));
    const log = await api.writeLedgerLog(`Escrow Funded: ${amount} INR for Transaction ${txId}`);
    set((state) => ({ ledger: [...state.ledger, log] }));
  },

  applyLoan: async (loanForm) => {
    const newLoan = await api.applyLoan(loanForm);
    set((state) => ({ loans: [...state.loans, newLoan] }));
    const log = await api.writeLedgerLog(`Loan Applied: ${newLoan.id} for Transaction ${newLoan.transactionId}`);
    set((state) => ({ ledger: [...state.ledger, log] }));
  },

  buyTokens: async (poolId, count) => {
    set((state) => ({
      pools: state.pools.map((p) => p.id === poolId ? { ...p, tokensSold: p.tokensSold + count } : p)
    }));
    const log = await api.writeLedgerLog(`Tokens Purchased: ${count} units in Pool ${poolId}`);
    set((state) => ({ ledger: [...state.ledger, log] }));
  },

  createPool: async (poolForm) => {
    const newPool = await api.createPool(poolForm);
    set((state) => ({ pools: [...state.pools, newPool] }));
    const log = await api.writeLedgerLog(`Fractional Asset Pool Created: ${newPool.id} (${newPool.propertyId})`);
    set((state) => ({ ledger: [...state.ledger, log] }));
  },

  setActiveTxId: (id) => set({ activeTxId: id }),

  login: async (email: string, password?: string) => {
    const data = await api.login(email, password);
    if (data && data.token) {
      if (typeof window !== 'undefined') {
        localStorage.setItem('user_email', email);
      }
      const users = await api.getUsers();
      const currentUser = users.find(u => u.email === email) || users[0] || null;
      set({ users, currentUser });
    } else {
      throw new Error("Invalid login response");
    }
  },

  logout: async () => {
    await api.logout();
    set({ currentUser: null });
  }
}));
