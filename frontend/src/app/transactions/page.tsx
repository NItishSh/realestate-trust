'use client';

import React, { useState } from 'react';
import { useStore } from '../../lib/store';
import {
  Plus,
  Coins,
  ChevronRight,
  CheckCircle2,
  XCircle,
  HelpCircle
} from 'lucide-react';
import { api } from '../../lib/api';

export default function Transactions() {
  const { transactions, createTransaction, fundEscrow } = useStore();
  const [selectedTxId, setSelectedTxId] = useState<string | null>(transactions[0]?.id || null);

  // Form states
  const [propertyId, setPropertyId] = useState('prop-101');
  const [buyerId, setBuyerId] = useState('usr-1');
  const [sellerId, setSellerId] = useState('usr-2');
  const [totalAmount, setTotalAmount] = useState('500000');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [fundAmount, setFundAmount] = useState('');
  const [isFunding, setIsFunding] = useState(false);

  const selectedTx = transactions.find(t => t.id === selectedTxId) || transactions[0];

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      await createTransaction({
        propertyId,
        buyerId,
        sellerId,
        totalAmount: parseFloat(totalAmount),
        status: 'DRAFT'
      });
      // Select the newly created tx
      const updatedTxList = useStore.getState().transactions;
      if (updatedTxList.length > 0) {
        setSelectedTxId(updatedTxList[updatedTxList.length - 1].id);
      }
    } catch (e) {
      console.error(e);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleFund = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedTx) return;
    setIsFunding(true);
    try {
      await fundEscrow(selectedTx.id, parseFloat(fundAmount));
      setFundAmount('');
    } catch (e) {
      console.error(e);
    } finally {
      setIsFunding(false);
    }
  };

  const handleTransition = async (targetStatus: 'ESCROW' | 'FUNDED' | 'CLOSED' | 'CANCELLED') => {
    if (!selectedTx) return;
    // Autonomously update state store in Zustand and write to Ledger
    useStore.setState((state) => ({
      transactions: state.transactions.map(t => t.id === selectedTx.id ? { ...t, status: targetStatus } : t)
    }));
    const log = await api.writeLedgerLog(`Transaction ${selectedTx.id} State transition to ${targetStatus}`);
    useStore.setState((state) => ({ ledger: [...state.ledger, log] }));
  };

  const getTimelineSteps = (status: string) => {
    const steps = [
      { name: 'Draft Configured', key: 'DRAFT', desc: 'Escrow instructions verified' },
      { name: 'Escrow Initialized', key: 'ESCROW', desc: 'Virtual accounts assigned' },
      { name: 'Account Funded', key: 'FUNDED', desc: 'FDIC audit balance confirm' },
      { name: 'Payout Released', key: 'CLOSED', desc: 'Deed register close complete' }
    ];

    const currentIdx = steps.findIndex(s => s.key === status);

    return steps.map((step, idx) => {
      let state: 'completed' | 'active' | 'upcoming' = 'upcoming';
      if (idx < currentIdx) state = 'completed';
      else if (idx === currentIdx) state = 'active';

      // Special case for Cancelled
      if (status === 'CANCELLED') {
        state = 'upcoming';
      }

      return { ...step, state };
    });
  };

  return (
    <div>
      {/* Header */}
      <div className="flex justify-between items-center mb-8">
        <div>
          <h2 className="text-3xl font-extrabold tracking-tight mb-2">Escrow Transactions</h2>
          <p className="text-slate-600 text-sm">Create virtual accounts, track escrow states, and execute payouts.</p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left Side: Create & List */}
        <div className="flex flex-col gap-8 lg:col-span-1">
          {/* New Transaction Form */}
          <div className="bg-white p-6 rounded-2xl">
            <h3 className="text-md font-bold mb-4 flex items-center gap-2">
              <Plus className="w-5 h-5 text-primary" />
              Configure Escrow
            </h3>
            <form onSubmit={handleCreate} className="flex flex-col gap-4">
              <div>
                <label className="text-[10px] text-slate-500 font-mono block mb-1">PROPERTY ID</label>
                <input
                  type="text"
                  value={propertyId}
                  onChange={e => setPropertyId(e.target.value)}
                  className="w-full bg-white border border-slate-200 rounded-xl px-4 py-2 text-xs outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus-visible:border-accent-cyan font-mono"
                  required
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-[10px] text-slate-500 font-mono block mb-1">BUYER ID</label>
                  <input
                    type="text"
                    value={buyerId}
                    onChange={e => setBuyerId(e.target.value)}
                    className="w-full bg-white border border-slate-200 rounded-xl px-4 py-2 text-xs outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus-visible:border-accent-cyan font-mono"
                    required
                  />
                </div>
                <div>
                  <label className="text-[10px] text-slate-500 font-mono block mb-1">SELLER ID</label>
                  <input
                    type="text"
                    value={sellerId}
                    onChange={e => setSellerId(e.target.value)}
                    className="w-full bg-white border border-slate-200 rounded-xl px-4 py-2 text-xs outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus-visible:border-accent-cyan font-mono"
                    required
                  />
                </div>
              </div>
              <div>
                <label className="text-[10px] text-slate-500 font-mono block mb-1">ESCROW AMOUNT (INR)</label>
                <input
                  type="number"
                  value={totalAmount}
                  onChange={e => setTotalAmount(e.target.value)}
                  className="w-full bg-white border border-slate-200 rounded-xl px-4 py-2 text-xs outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus-visible:border-accent-cyan font-mono"
                  required
                />
              </div>

              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full btn-primary text-white font-bold text-xs py-3 rounded-xl hover:opacity-90 active:scale-95 transition-colors transition-transform transition-opacity flex items-center justify-center gap-2 cursor-pointer"
              >
                {isSubmitting ? 'Creating…' : 'Initialize Escrow'}
                <ChevronRight className="w-4 h-4" />
              </button>
            </form>
          </div>

          {/* List of Escrows */}
          <div className="bg-white p-6 rounded-2xl flex-1">
            <h3 className="text-md font-bold mb-4">Active Escrow Accounts</h3>
            <div className="flex flex-col gap-2 max-h-[300px] overflow-y-auto">
              {transactions.map((tx) => (
                <button
                  key={tx.id}
                  onClick={() => setSelectedTxId(tx.id)}
                  className={`w-full p-3 rounded-xl text-left border transition-colors transition-transform transition-opacity flex items-center justify-between ${
                    selectedTx?.id === tx.id
                      ? 'bg-slate-100/40 border-accent-cyan/40'
                      : 'bg-transparent border-slate-200 hover:bg-white shadow-sm'
                  }`}
                >
                  <div>
                    <h4 className="text-xs font-bold text-slate-800 font-mono">{tx.id}</h4>
                    <span className="text-[9px] text-slate-500 font-mono">Amt: ₹{tx.totalAmount.toLocaleString('en-IN')}</span>
                  </div>
                  <span className={`text-[9px] font-mono font-bold px-2 py-0.5 rounded-full ${
                    tx.status === 'CLOSED' ? 'bg-green-50 text-success' :
                    tx.status === 'FUNDED' ? 'bg-accent-blue/10 text-accent-blue' :
                    tx.status === 'ESCROW' ? 'bg-accent-cyan/10 text-primary' :
                    'bg-slate-100 text-slate-600'
                  }`}>
                    {tx.status}
                  </span>
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Right Side: Timeline & Details */}
        <div className="lg:col-span-2">
          {selectedTx ? (
            <div className="bg-white p-8 rounded-2xl h-full flex flex-col justify-between">
              {/* Top details */}
              <div>
                <div className="flex justify-between items-start border-b border-slate-200 pb-6 mb-6">
                  <div>
                    <span className="text-[10px] font-mono text-primary uppercase tracking-wider">Escrow Deal Sheet</span>
                    <h3 className="text-xl font-extrabold font-mono mt-1">{selectedTx.id}</h3>
                  </div>
                  <div className="text-right">
                    <span className="text-xs text-slate-600 font-sans block">Total Purchase Value</span>
                    <span className="text-xl font-bold font-mono text-slate-800">₹{selectedTx.totalAmount.toLocaleString('en-IN')}</span>
                  </div>
                </div>

                {/* Progress bar timeline */}
                <h4 className="text-xs font-bold text-slate-600 uppercase tracking-wide mb-6">Escrow Progression Stage</h4>

                <div className="grid grid-cols-4 gap-4 relative mb-12">
                  {getTimelineSteps(selectedTx.status).map((step, idx) => {
                    const isLast = idx === 3;
                    return (
                      <div key={step.name} className="flex flex-col relative">
                        {/* Circle and Line */}
                        <div className="flex items-center mb-3">
                          <div className={`w-8 h-8 rounded-full flex items-center justify-center border font-bold text-xs font-mono shrink-0 transition-colors duration-300 ${
                            step.state === 'completed' ? 'bg-accent-cyan/20 border-accent-cyan text-primary' :
                            step.state === 'active' ? 'bg-accent-cyan text-white border-accent-cyan shadow-lg shadow-accent-cyan/25' :
                            'bg-white border-slate-200 text-slate-500'
                          }`}>
                            {step.state === 'completed' ? <CheckCircle2 className="w-4 h-4 text-primary" /> : idx + 1}
                          </div>
                          {!isLast && (
                            <div className={`h-[2px] w-full mx-2 transition-colors duration-500 ${
                              step.state === 'completed' ? 'bg-accent-cyan' : 'bg-slate-100'
                            }`} />
                          )}
                        </div>

                        {/* Labels */}
                        <span className={`text-xs font-bold ${
                          step.state === 'active' ? 'text-primary' : 'text-slate-700'
                        }`}>{step.name}</span>
                        <span className="text-[10px] text-slate-500 mt-1 leading-tight">{step.desc}</span>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Action buttons */}
              <div className="border-t border-slate-200 pt-6 mt-6">
                <h4 className="text-xs font-bold text-slate-600 uppercase tracking-wide mb-4">Execute Escrow Actions</h4>

                <div className="flex gap-4">
                  {selectedTx.status === 'DRAFT' && (
                    <button
                      onClick={() => handleTransition('ESCROW')}
                      className="px-6 py-3 rounded-xl bg-accent-cyan text-white text-xs font-bold hover:opacity-90 transition-opacity flex items-center gap-2 cursor-pointer"
                    >
                      <Plus className="w-4 h-4" />
                      Initialize Escrow Account
                    </button>
                  )}
                  {selectedTx.status === 'ESCROW' && (
                    <form onSubmit={handleFund} className="flex gap-2">
                      <input
                        type="number"
                        value={fundAmount}
                        onChange={(e) => setFundAmount(e.target.value)}
                        placeholder="Amount to deposit"
                        className="bg-white border border-slate-200 rounded-xl px-4 py-2 text-xs outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus:border-accent-blue font-mono w-40"
                        required
                      />
                      <button
                        type="submit"
                        disabled={isFunding}
                        className="px-6 py-3 rounded-xl bg-accent-blue text-white text-xs font-bold hover:opacity-90 transition-opacity flex items-center gap-2 cursor-pointer disabled:opacity-50"
                      >
                        <Coins className="w-4 h-4" />
                        {isFunding ? 'Funding…' : 'Confirm Funding Deposit'}
                      </button>
                    </form>
                  )}
                  {selectedTx.status === 'FUNDED' && (
                    <button
                      onClick={() => handleTransition('CLOSED')}
                      className="px-6 py-3 rounded-xl bg-green-500 text-white text-xs font-bold hover:opacity-90 transition-opacity flex items-center gap-2 cursor-pointer"
                    >
                      <CheckCircle2 className="w-4 h-4" />
                      Release Payouts & Close
                    </button>
                  )}

                  {selectedTx.status !== 'CLOSED' && selectedTx.status !== 'CANCELLED' && (
                    <button
                      onClick={() => handleTransition('CANCELLED')}
                      className="px-6 py-3 rounded-xl border border-rose-500/20 text-rose-500 text-xs font-bold hover:bg-rose-500/5 transition-colors flex items-center gap-2 cursor-pointer"
                    >
                      <XCircle className="w-4 h-4" />
                      Cancel Transaction
                    </button>
                  )}

                  {selectedTx.status === 'CLOSED' && (
                    <div className="p-4 rounded-xl bg-green-500/5 border border-accent-emerald/10 text-success text-xs flex items-center gap-2 w-full font-mono">
                      <CheckCircle2 className="w-5 h-5 shrink-0" />
                      This escrow transaction has been closed and payouts are settled. The cryptographic block hashes are sealed in the ledger.
                    </div>
                  )}

                  {selectedTx.status === 'CANCELLED' && (
                    <div className="p-4 rounded-xl bg-rose-500/5 border border-rose-500/10 text-rose-500 text-xs flex items-center gap-2 w-full font-mono">
                      <XCircle className="w-5 h-5 shrink-0" />
                      This transaction has been cancelled. Virtual accounts are closed.
                    </div>
                  )}
                </div>
              </div>
            </div>
          ) : (
            <div className="bg-white p-8 rounded-2xl h-full flex flex-col items-center justify-center text-slate-500 gap-2">
              <HelpCircle className="w-8 h-8 text-slate-600" />
              <p className="text-sm font-mono">No active transactions found. Initialize an escrow first.</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
