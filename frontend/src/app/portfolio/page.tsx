'use client';

import React, { useState } from 'react';
import { useStore } from '../../lib/store';
import {
  PieChart,
  Coins,
  TrendingUp,
  Plus,
  Building2,
  Search,
  CircleDollarSign,
  Briefcase,
  AlertCircle
} from 'lucide-react';

export default function Portfolio() {
  const { pools, buyTokens, currentUser, createPool } = useStore();
  const [selectedPoolId, setSelectedPoolId] = useState<string | null>(pools[0]?.id || null);
  const [shareCount, setShareCount] = useState('10');
  const [buyLoading, setBuyLoading] = useState(false);

  // New Pool Form State
  const [propertyId, setPropertyId] = useState('prop-103');
  const [totalTokens, setTotalTokens] = useState('5000');
  const [tokenPrice, setTokenPrice] = useState('150.00');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const selectedPool = pools.find(p => p.id === selectedPoolId) || pools[0];

  const handleBuy = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedPool || !currentUser) return;
    if (currentUser.kycStatus !== 'APPROVED') {
      alert("Verification Required: Please approve KYC verification in the 'KYC Onboarding' page before purchasing fractional shares!");
      return;
    }
    setBuyLoading(true);
    try {
      await buyTokens(selectedPool.id, parseInt(shareCount));
      setShareCount('10');
    } catch (e) {
      console.error(e);
    } finally {
      setBuyLoading(false);
    }
  };

  const handleCreatePool = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      await createPool({
        propertyId,
        totalTokens: parseInt(totalTokens),
        tokenPrice: parseFloat(tokenPrice)
      });
      setPropertyId('prop-' + Math.floor(Math.random() * 1000));
    } catch (e) {
      console.error(e);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Portfolio aggregates
  const totalValueInvested = pools.reduce((sum, p) => sum + (p.tokensSold * p.tokenPrice), 0);
  const estimatedYield = totalValueInvested * 0.082; // 8.2% annual dividend yields

  return (
    <div>
      {/* Header */}
      <div className="mb-8">
        <h2 className="text-3xl font-extrabold tracking-tight mb-2">Fractional Portfolios</h2>
        <p className="text-slate-400 text-sm">Monitor property fractionalization pools, execute share acquisitions, and track yield dividends.</p>
      </div>

      {/* Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="glass-panel p-6 rounded-2xl flex items-center gap-4">
          <div className="p-3 rounded-xl bg-accent-cyan/10">
            <Briefcase className="w-6 h-6 text-accent-cyan" />
          </div>
          <div>
            <span className="text-[10px] font-semibold text-slate-500 tracking-wide uppercase">Holdings Portfolio Value</span>
            <h4 className="text-xl font-bold font-mono mt-0.5">${totalValueInvested.toLocaleString()}</h4>
          </div>
        </div>
        <div className="glass-panel p-6 rounded-2xl flex items-center gap-4">
          <div className="p-3 rounded-xl bg-accent-emerald/10">
            <TrendingUp className="w-6 h-6 text-accent-emerald" />
          </div>
          <div>
            <span className="text-[10px] font-semibold text-slate-500 tracking-wide uppercase">Est. Annualized Dividends (8.2%)</span>
            <h4 className="text-xl font-bold font-mono mt-0.5 text-accent-emerald">${estimatedYield.toLocaleString()}</h4>
          </div>
        </div>
        <div className="glass-panel p-6 rounded-2xl flex items-center gap-4">
          <div className="p-3 rounded-xl bg-accent-blue/10">
            <Coins className="w-6 h-6 text-accent-blue" />
          </div>
          <div>
            <span className="text-[10px] font-semibold text-slate-500 tracking-wide uppercase">Active Asset Pools</span>
            <h4 className="text-xl font-bold font-mono mt-0.5">{pools.length} Assets</h4>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left Side: Pool Creation & Selector */}
        <div className="flex flex-col gap-8 lg:col-span-1">
          {/* Create Pool Form */}
          <div className="glass-panel p-6 rounded-2xl">
            <h3 className="text-md font-bold mb-4 flex items-center gap-2">
              <Plus className="w-5 h-5 text-accent-cyan" />
              Configure Property Pool
            </h3>
            <form onSubmit={handleCreatePool} className="flex flex-col gap-4">
              <div>
                <label className="text-[10px] text-slate-500 font-mono block mb-1">PROPERTY ID</label>
                <input
                  type="text"
                  value={propertyId}
                  onChange={e => setPropertyId(e.target.value)}
                  className="w-full bg-slate-900 border border-card-border rounded-xl px-4 py-2 text-xs focus:outline-none focus:border-accent-cyan font-mono"
                  required
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-[10px] text-slate-500 font-mono block mb-1">TOTAL SHARES</label>
                  <input
                    type="number"
                    value={totalTokens}
                    onChange={e => setTotalTokens(e.target.value)}
                    className="w-full bg-slate-900 border border-card-border rounded-xl px-4 py-2 text-xs focus:outline-none focus:border-accent-cyan font-mono"
                    required
                  />
                </div>
                <div>
                  <label className="text-[10px] text-slate-500 font-mono block mb-1">SHARE PRICE ($)</label>
                  <input
                    type="text"
                    value={tokenPrice}
                    onChange={e => setTokenPrice(e.target.value)}
                    className="w-full bg-slate-900 border border-card-border rounded-xl px-4 py-2 text-xs focus:outline-none focus:border-accent-cyan font-mono"
                    required
                  />
                </div>
              </div>

              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full bg-gradient-to-r from-accent-cyan to-accent-blue text-bg-primary font-bold text-xs py-3 rounded-xl hover:opacity-90 active:scale-95 transition-all flex items-center justify-center gap-2 cursor-pointer"
              >
                {isSubmitting ? 'Creating...' : 'Initialize Pool'}
              </button>
            </form>
          </div>

          {/* List of Pools */}
          <div className="glass-panel p-6 rounded-2xl flex-1">
            <h3 className="text-md font-bold mb-4">Investable Pools</h3>
            <div className="flex flex-col gap-2 max-h-[300px] overflow-y-auto">
              {pools.map((pool) => {
                const percentSold = (pool.tokensSold / pool.totalTokens) * 100;
                return (
                  <button
                    key={pool.id}
                    onClick={() => setSelectedPoolId(pool.id)}
                    className={`w-full p-3 rounded-xl text-left border transition-all flex items-center justify-between ${
                      selectedPool?.id === pool.id
                        ? 'bg-slate-800/40 border-accent-cyan/40'
                        : 'bg-transparent border-card-border hover:bg-slate-900/40'
                    }`}
                  >
                    <div>
                      <h4 className="text-xs font-bold text-slate-200 font-mono flex items-center gap-1.5">
                        <Building2 className="w-3.5 h-3.5 text-accent-cyan" />
                        {pool.propertyId}
                      </h4>
                      <span className="text-[9px] text-slate-500 font-mono mt-1 block">Sold: {percentSold.toFixed(0)}%</span>
                    </div>
                    <span className="text-xs font-mono font-bold text-accent-cyan">${pool.tokenPrice}</span>
                  </button>
                );
              })}
            </div>
          </div>
        </div>

        {/* Right Side: Purchase Form and Details */}
        <div className="lg:col-span-2">
          {selectedPool ? (
            <div className="glass-panel p-8 rounded-2xl h-full flex flex-col justify-between">
              <div>
                <div className="flex justify-between items-start border-b border-card-border pb-6 mb-6">
                  <div>
                    <span className="text-[10px] font-mono text-accent-cyan uppercase">Fractional Real Estate Pool</span>
                    <h3 className="text-xl font-bold font-mono mt-1">{selectedPool.propertyId}</h3>
                  </div>
                  <div className="text-right">
                    <span className="text-xs text-slate-400 block">Unit Cost</span>
                    <span className="text-xl font-bold font-mono text-slate-200">${selectedPool.tokenPrice.toFixed(2)}</span>
                  </div>
                </div>

                {/* Progress bar */}
                <div className="mb-8">
                  <div className="flex justify-between text-xs text-slate-400 font-semibold mb-2">
                    <span>Pool Allocation Status</span>
                    <span>{((selectedPool.tokensSold / selectedPool.totalTokens) * 100).toFixed(1)}% Subscribed</span>
                  </div>
                  <div className="w-full bg-slate-900 h-3 rounded-full overflow-hidden border border-card-border">
                    <div
                      className="bg-gradient-to-r from-accent-cyan to-accent-blue h-full transition-all duration-500"
                      style={{ width: `${(selectedPool.tokensSold / selectedPool.totalTokens) * 100}%` }}
                    />
                  </div>
                  <div className="flex justify-between text-[10px] font-mono text-slate-500 mt-2">
                    <span>{selectedPool.tokensSold} Shares purchased</span>
                    <span>{selectedPool.totalTokens - selectedPool.tokensSold} Shares remaining</span>
                  </div>
                </div>
              </div>

              {/* Purchase interface */}
              <div className="border-t border-card-border pt-6 mt-6">
                <h4 className="text-xs font-bold text-slate-400 uppercase tracking-wide mb-4">Invest in Asset shares</h4>

                {currentUser && currentUser.kycStatus !== 'APPROVED' && (
                  <div className="p-4 rounded-xl bg-rose-500/5 border border-rose-500/10 text-rose-500 text-xs flex items-center gap-3 mb-4 leading-tight font-mono">
                    <AlertCircle className="w-5 h-5 shrink-0" />
                    Acquisition Locked: Please submit and approve your KYC profile before executing property share purchases.
                  </div>
                )}

                <form onSubmit={handleBuy} className="flex gap-4 items-end">
                  <div className="w-32">
                    <label className="text-[10px] text-slate-500 font-mono block mb-1">SHARES COUNT</label>
                    <input
                      type="number"
                      value={shareCount}
                      onChange={e => setShareCount(e.target.value)}
                      disabled={currentUser?.kycStatus !== 'APPROVED'}
                      className="w-full bg-slate-900 border border-card-border rounded-xl px-4 py-2.5 text-xs focus:outline-none focus:border-accent-cyan font-mono"
                      required
                    />
                  </div>

                  <div className="flex-1">
                    <button
                      type="submit"
                      disabled={buyLoading || currentUser?.kycStatus !== 'APPROVED'}
                      className={`w-full font-bold text-xs py-3 rounded-xl transition-all flex items-center justify-center gap-2 cursor-pointer ${
                        currentUser?.kycStatus !== 'APPROVED'
                          ? 'bg-slate-800 text-slate-600 border border-card-border cursor-not-allowed'
                          : 'bg-gradient-to-r from-accent-cyan to-accent-blue text-bg-primary hover:opacity-90 active:scale-95'
                      }`}
                    >
                      {buyLoading ? 'Processing...' : `Purchase shares for $${(parseInt(shareCount || '0') * selectedPool.tokenPrice).toLocaleString()}`}
                    </button>
                  </div>
                </form>
              </div>
            </div>
          ) : (
            <div className="glass-panel p-8 rounded-2xl h-full flex flex-col items-center justify-center text-slate-500 gap-2">
              <AlertCircle className="w-8 h-8 text-slate-600" />
              <p className="text-sm font-mono">No active investment pools. Initialize a property pool first.</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
