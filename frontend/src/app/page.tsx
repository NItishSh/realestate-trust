'use client';

import React from 'react';
import { useStore } from '../lib/store';
import {
  ShieldAlert,
  Coins,
  CircleDollarSign,
  Activity,
  FileCheck2,
  TrendingUp,
  Cpu
} from 'lucide-react';
import Link from 'next/link';

export default function Dashboard() {
  const { transactions, loans, pools, ledger, isLoading } = useStore();

  // Metrics calculations
  const totalEscrowLocked = transactions
    .filter(tx => tx.status === 'ESCROW' || tx.status === 'FUNDED')
    .reduce((sum, tx) => sum + tx.totalAmount, 0);

  const activeDeals = transactions.filter(tx => tx.status !== 'CLOSED' && tx.status !== 'CANCELLED').length;
  const approvedLoans = loans.filter(l => l.status === 'APPROVED' || l.status === 'DISBURSED').length;
  const totalPoolsSize = pools.length;

  const cardData = [
    { name: 'Escrow Volume Locked', val: `$${totalEscrowLocked.toLocaleString()}`, unit: 'USD', icon: Coins, color: 'text-accent-cyan', bg: 'bg-accent-cyan/10' },
    { name: 'Active Trust Transactions', val: activeDeals, unit: 'Deals', icon: CircleDollarSign, color: 'text-accent-blue', bg: 'bg-accent-blue/10' },
    { name: 'Underwritten Mortgages', val: approvedLoans, unit: 'Approved', icon: FileCheck2, color: 'text-accent-emerald', bg: 'bg-accent-emerald/10' },
    { name: 'Cryptographic Blocks', val: ledger.length, unit: 'Sealed', icon: Cpu, color: 'text-violet-400', bg: 'bg-violet-400/10' },
  ];

  if (isLoading) {
    return (
      <div className="flex h-[80vh] items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <RefreshCwIcon className="w-8 h-8 text-accent-cyan animate-spin" />
          <p className="text-slate-400 text-sm font-mono">Synchronizing workspace states...</p>
        </div>
      </div>
    );
  }

  return (
    <div>
      {/* Welcome Header */}
      <div className="mb-8">
        <h2 className="text-3xl font-extrabold tracking-tight mb-2">
          RealEstate <span className="gradient-text">Trust Operations</span>
        </h2>
        <p className="text-slate-400 text-sm">
          Platform-wide metrics monitoring real-time escrow, smart fractionalization, and immutable audit logs.
        </p>
      </div>

      {/* Metrics Cards Grid */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        {cardData.map((card) => {
          const Icon = card.icon;
          return (
            <div key={card.name} className="glass-panel p-6 rounded-2xl relative overflow-hidden group transition-all duration-300 hover:-translate-y-1">
              <div className="flex justify-between items-start mb-4">
                <span className="text-xs font-semibold text-slate-400 tracking-wide uppercase">{card.name}</span>
                <div className={`p-2 rounded-xl ${card.bg}`}>
                  <Icon className={`w-5 h-5 ${card.color}`} />
                </div>
              </div>
              <div className="flex items-baseline gap-2">
                <span className="text-2xl font-bold font-mono tracking-tight">{card.val}</span>
                <span className="text-xs text-slate-500 font-mono">{card.unit}</span>
              </div>
              {/* Card bottom glow */}
              <div className="absolute bottom-0 left-0 w-full h-[2px] bg-gradient-to-r from-transparent via-accent-cyan/20 to-transparent scale-x-0 group-hover:scale-x-100 transition-transform duration-300" />
            </div>
          );
        })}
      </div>

      {/* Grid: Activity Log and Pools Overview */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left Side: Cryptographic Audit Trail */}
        <div className="lg:col-span-2 glass-panel p-6 rounded-2xl">
          <div className="flex justify-between items-center mb-6">
            <div>
              <h3 className="text-md font-bold tracking-tight">Audit Trail Activity</h3>
              <p className="text-xs text-slate-400">Latest immutable logs retrieved from `ledger-service`.</p>
            </div>
            <Link href="/ledger" className="text-xs text-accent-cyan hover:underline font-mono">
              View block timeline →
            </Link>
          </div>

          <div className="flex flex-col gap-4">
            {ledger.slice().reverse().slice(0, 4).map((block) => (
              <div key={block.index} className="flex gap-4 p-3 rounded-xl hover:bg-slate-800/20 transition-colors">
                <div className="flex flex-col items-center">
                  <div className="w-8 h-8 rounded-lg bg-slate-800 flex items-center justify-center shrink-0 border border-card-border font-mono text-xs font-bold text-slate-400">
                    #{block.index}
                  </div>
                  <div className="w-[1px] h-full bg-slate-800 mt-2" />
                </div>
                <div className="flex-1 overflow-hidden">
                  <div className="flex items-center justify-between gap-4 mb-1">
                    <p className="text-xs font-semibold truncate text-slate-200">{block.payload}</p>
                    <span className="text-[10px] text-slate-500 font-mono shrink-0">
                      {new Date(block.timestamp).toLocaleTimeString()}
                    </span>
                  </div>
                  <div className="flex items-center gap-1.5 font-mono text-[9px] text-slate-500">
                    <span className="text-accent-cyan">HASH:</span>
                    <span className="truncate">{block.hash}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Right Side: Fractional Pools Preview */}
        <div className="glass-panel p-6 rounded-2xl flex flex-col justify-between">
          <div>
            <div className="flex justify-between items-center mb-6">
              <h3 className="text-md font-bold tracking-tight font-sans">Fractional Pools</h3>
              <Link href="/portfolio" className="text-xs text-accent-cyan hover:underline font-mono">
                Invest →
              </Link>
            </div>

            <div className="flex flex-col gap-4">
              {pools.slice(0, 2).map((pool) => {
                const percentSold = (pool.tokensSold / pool.totalTokens) * 100;
                return (
                  <div key={pool.id} className="p-4 rounded-xl bg-slate-900/40 border border-card-border">
                    <div className="flex justify-between items-center mb-2">
                      <span className="text-xs font-bold text-slate-200 uppercase">{pool.propertyId}</span>
                      <span className="text-xs font-mono text-accent-cyan">${pool.tokenPrice}/share</span>
                    </div>

                    <div className="w-full bg-slate-800 h-1.5 rounded-full overflow-hidden mb-2">
                      <div className="bg-gradient-to-r from-accent-cyan to-accent-blue h-full" style={{ width: `${percentSold}%` }} />
                    </div>

                    <div className="flex justify-between text-[10px] font-mono text-slate-500">
                      <span>{pool.tokensSold} / {pool.totalTokens} shares sold</span>
                      <span className="text-slate-300 font-semibold">{percentSold.toFixed(0)}%</span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          <div className="p-4 rounded-xl bg-accent-cyan/5 border border-accent-cyan/10 flex items-center gap-3 mt-6">
            <ShieldAlert className="w-5 h-5 text-accent-cyan shrink-0" />
            <p className="text-[10px] text-slate-400 leading-tight">
              Escrow integrations run under FDIC-insured partner ledgers using virtual accounts dynamically created per transaction.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

function RefreshCwIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
      <path d="M3 3v5h5" />
      <path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16" />
      <path d="M16 16h5v5" />
    </svg>
  );
}
