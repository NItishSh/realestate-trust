import { describe, it, expect, beforeEach } from 'vitest';
import { useStore } from '../lib/store';
import { User } from '../lib/api';

describe('useStore', () => {
  beforeEach(() => {
    useStore.setState({
      users: [],
      transactions: [],
      loans: [],
      pools: [],
      ledger: [],
      isLoading: false,
      currentUser: null,
      activeTxId: null,
    });
  });

  it('initializes with default state', () => {
    const state = useStore.getState();
    expect(state.currentUser).toBeNull();
    expect(state.activeTxId).toBeNull();
    expect(state.users).toEqual([]);
    expect(state.transactions).toEqual([]);
    expect(state.isLoading).toBe(false);
  });

  it('sets current user correctly', () => {
    const mockUser: User = {
      id: 'usr-1',
      email: 'buyer@realestate-trust.local',
      fullName: 'Alice Buyer',
      role: 'BUYER',
      kycStatus: 'APPROVED',
      created_at: '2026-09-01T00:00:00Z',
    };

    useStore.getState().setCurrentUser(mockUser);
    expect(useStore.getState().currentUser).toEqual(mockUser);
  });

  it('sets active transaction ID correctly', () => {
    useStore.getState().setActiveTxId('tx-12345');
    expect(useStore.getState().activeTxId).toBe('tx-12345');

    useStore.getState().setActiveTxId(null);
    expect(useStore.getState().activeTxId).toBeNull();
  });
});
