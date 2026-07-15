'use client';

import React, { useState } from 'react';
import { useStore } from '../lib/store';
import {
  ShieldAlert,
  Coins,
  CircleDollarSign,
  Activity,
  FileCheck2,
  TrendingUp,
  Cpu,
  ArrowRight,
  BookOpen
} from 'lucide-react';
import Link from 'next/link';

export default function Dashboard() {
  const { transactions, loans, pools, ledger, isLoading } = useStore();
  const [activeJourney, setActiveJourney] = useState<'buyer' | 'investor' | 'auditor'>('buyer');

  // Metrics calculations
  const totalEscrowLocked = transactions
    .filter(tx => tx.status === 'ESCROW' || tx.status === 'FUNDED')
    .reduce((sum, tx) => sum + tx.totalAmount, 0);

  const activeDeals = transactions.filter(tx => tx.status !== 'CLOSED' && tx.status !== 'CANCELLED').length;
  const approvedLoans = loans.filter(l => l.status === 'APPROVED' || l.status === 'DISBURSED').length;
  const totalPoolsSize = pools.length;

  const cardData = [
    { name: 'Escrow Volume Locked', val: `₹${totalEscrowLocked.toLocaleString('en-IN')}`, unit: 'INR', icon: Coins, color: 'text-accent-cyan', bg: 'bg-accent-cyan/10' },
    { name: 'Active Trust Transactions', val: activeDeals, unit: 'Deals', icon: CircleDollarSign, color: 'text-accent-blue', bg: 'bg-accent-blue/10' },
    { name: 'Underwritten Mortgages', val: approvedLoans, unit: 'Approved', icon: FileCheck2, color: 'text-accent-emerald', bg: 'bg-accent-emerald/10' },
    { name: 'Cryptographic Blocks', val: ledger.length, unit: 'Sealed', icon: Cpu, color: 'text-violet-400', bg: 'bg-violet-400/10' },
  ];

  if (isLoading) {
    return (
      <div className="flex h-[80vh] items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <RefreshCwIcon className="w-8 h-8 text-accent-cyan animate-spin" />
          <p className="text-slate-400 text-sm font-mono">Synchronizing workspace states…</p>
        </div>
      </div>
    );
  }

  return (
    <div>
      {/* Welcome Header */}
      <div className="mb-8 border-b border-card-border pb-6">
        <h2 className="text-4xl font-bold tracking-wide mb-2 text-slate-100 font-serif uppercase">
          RealEstate <span className="gradient-text">Trust Operations</span>
        </h2>
        <p className="text-slate-400 text-sm font-sans tracking-wide">
          Platform-wide metrics monitoring real-time escrow, smart fractionalization, and immutable audit logs.
        </p>
      </div>

      {/* Metrics Cards Grid */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        {cardData.map((card) => {
          const Icon = card.icon;
          return (
            <div key={card.name} className="glass-panel p-6 rounded-2xl relative overflow-hidden group transition-all duration-300 hover:-translate-y-1 hover:border-accent-gold/20 cursor-pointer">
              <div className="flex justify-between items-start mb-4">
                <span className="text-[10px] font-bold text-slate-400 tracking-wider uppercase font-mono">{card.name}</span>
                <div className={`p-2 rounded-xl ${card.bg}`}>
                  <Icon className={`w-5 h-5 ${card.color}`} />
                </div>
              </div>
              <div className="flex items-baseline gap-2">
                <span className="text-2xl font-bold font-mono tracking-tight">{card.val}</span>
                <span className="text-xs text-slate-500 font-mono">{card.unit}</span>
              </div>
              {/* Card bottom glow */}
              <div className="absolute bottom-0 left-0 w-full h-[2px] bg-gradient-to-r from-transparent via-accent-gold/30 to-transparent scale-x-0 group-hover:scale-x-100 transition-transform duration-300" />
            </div>
          );
        })}
      </div>

      {/* Guided Demo Journeys */}
      <div className="glass-panel p-6 rounded-2xl mb-8">
        <div className="flex justify-between items-center mb-6">
          <div>
            <h3 className="text-md font-bold tracking-wide uppercase text-slate-200">Interactive Guided Journeys</h3>
            <p className="text-xs text-slate-400">Step-by-step walkthroughs to test the integrated microservices.</p>
          </div>
          <Link
            href="file:///Users/nitishshanchinagoudra/workspace/me/realestate-trust/docs/user_journeys.md"
            target="_blank"
            className="flex items-center gap-1.5 text-xs text-accent-gold hover:underline font-mono"
          >
            <BookOpen className="w-3.5 h-3.5" />
            Read Spec Manual
          </Link>
        </div>

        <div className="flex gap-4 border-b border-card-border pb-4 mb-6 flex-wrap">
          <button
            onClick={() => setActiveJourney('buyer')}
            className={`px-4 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer ${
              activeJourney === 'buyer' ? 'bg-accent-gold/15 text-accent-gold border border-accent-gold/20' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Buyer Escrow Flow
          </button>
          <button
            onClick={() => setActiveJourney('investor')}
            className={`px-4 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer ${
              activeJourney === 'investor' ? 'bg-accent-cyan/15 text-accent-cyan border border-accent-cyan/20' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Fractional Investor Flow
          </button>
          <button
            onClick={() => setActiveJourney('auditor')}
            className={`px-4 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer ${
              activeJourney === 'auditor' ? 'bg-accent-emerald/15 text-accent-emerald border border-accent-emerald/20' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            System Auditor Flow
          </button>
        </div>

        {activeJourney === 'buyer' && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <Link href="/kyc" className="p-4 rounded-xl bg-slate-900/40 border border-card-border hover:border-accent-cyan/20 transition-colors transition-transform transition-opacity group">
              <div className="text-accent-cyan font-bold text-xs mb-1 group-hover:underline flex items-center gap-1">
                Step 1: Onboard Buyer <ArrowRight className="w-3 h-3 group-hover:translate-x-0.5 transition-transform" />
              </div>
              <p className="text-[10px] text-slate-500">Go to KYC tab, register account as Buyer, and simulate KYC approval.</p>
            </Link>
            <Link href="/transactions" className="p-4 rounded-xl bg-slate-900/40 border border-card-border hover:border-accent-cyan/20 transition-colors transition-transform transition-opacity group">
              <div className="text-accent-cyan font-bold text-xs mb-1 group-hover:underline flex items-center gap-1">
                Step 2: Init Escrow <ArrowRight className="w-3 h-3 group-hover:translate-x-0.5 transition-transform" />
              </div>
              <p className="text-[10px] text-slate-500">Go to Escrow tab and create a new escrow transaction deal.</p>
            </Link>
            <Link href="/transactions" className="p-4 rounded-xl bg-slate-900/40 border border-card-border hover:border-accent-cyan/20 transition-colors transition-transform transition-opacity group">
              <div className="text-accent-cyan font-bold text-xs mb-1 group-hover:underline flex items-center gap-1">
                Step 3: Fund Virtual Acct <ArrowRight className="w-3 h-3 group-hover:translate-x-0.5 transition-transform" />
              </div>
              <p className="text-[10px] text-slate-500">Select transaction, initialize virtual account, and confirm deposit.</p>
            </Link>
            <Link href="/ledger" className="p-4 rounded-xl bg-slate-900/40 border border-card-border hover:border-accent-cyan/20 transition-colors transition-transform transition-opacity group">
              <div className="text-accent-cyan font-bold text-xs mb-1 group-hover:underline flex items-center gap-1">
                Step 4: Release & Seal <ArrowRight className="w-3 h-3 group-hover:translate-x-0.5 transition-transform" />
              </div>
              <p className="text-[10px] text-slate-500">Close the escrow to release payout. Verify the sealed hash in Ledger logs.</p>
            </Link>
          </div>
        )}

        {activeJourney === 'investor' && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Link href="/kyc" className="p-4 rounded-xl bg-slate-900/40 border border-card-border hover:border-accent-blue/20 transition-colors transition-transform transition-opacity group">
              <div className="text-accent-blue font-bold text-xs mb-1 group-hover:underline flex items-center gap-1">
                Step 1: Verify KYC <ArrowRight className="w-3 h-3 group-hover:translate-x-0.5 transition-transform" />
              </div>
              <p className="text-[10px] text-slate-500">Ensure your active user status is KYC APPROVED to unlock pools purchase.</p>
            </Link>
            <Link href="/portfolio" className="p-4 rounded-xl bg-slate-900/40 border border-card-border hover:border-accent-blue/20 transition-colors transition-transform transition-opacity group">
              <div className="text-accent-blue font-bold text-xs mb-1 group-hover:underline flex items-center gap-1">
                Step 2: Browse Pools <ArrowRight className="w-3 h-3 group-hover:translate-x-0.5 transition-transform" />
              </div>
              <p className="text-[10px] text-slate-500">Explore real estate listings (e.g. PROP-101) and share costs.</p>
            </Link>
            <Link href="/portfolio" className="p-4 rounded-xl bg-slate-900/40 border border-card-border hover:border-accent-blue/20 transition-colors transition-transform transition-opacity group">
              <div className="text-accent-blue font-bold text-xs mb-1 group-hover:underline flex items-center gap-1">
                Step 3: Buy & Track Yield <ArrowRight className="w-3 h-3 group-hover:translate-x-0.5 transition-transform" />
              </div>
              <p className="text-[10px] text-slate-500">Purchase shares, observe dividend increments, and inspect logs.</p>
            </Link>
          </div>
        )}

        {activeJourney === 'auditor' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Link href="/ledger" className="p-4 rounded-xl bg-slate-900/40 border border-card-border hover:border-violet-400/20 transition-colors transition-transform transition-opacity group">
              <div className="text-violet-400 font-bold text-xs mb-1 group-hover:underline flex items-center gap-1">
                Step 1: Write Custom Block <ArrowRight className="w-3 h-3 group-hover:translate-x-0.5 transition-transform" />
              </div>
              <p className="text-[10px] text-slate-500">Enter custom log details and publish a sealed compliance record block.</p>
            </Link>
            <Link href="/ledger" className="p-4 rounded-xl bg-slate-900/40 border border-card-border hover:border-violet-400/20 transition-colors transition-transform transition-opacity group">
              <div className="text-violet-400 font-bold text-xs mb-1 group-hover:underline flex items-center gap-1">
                Step 2: Verify Hashes Chaining <ArrowRight className="w-3 h-3 group-hover:translate-x-0.5 transition-transform" />
              </div>
              <p className="text-[10px] text-slate-500">Analyze the SHA256 chain. Verify that Block Hash matches previous block links.</p>
            </Link>
          </div>
        )}
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
                      <span className="text-xs font-mono text-accent-cyan">₹{pool.tokenPrice.toLocaleString('en-IN')}/share</span>
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
