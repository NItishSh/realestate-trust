'use client';

import React, { useEffect, useState } from 'react';
import {
  Building,
  CreditCard,
  CheckCircle,
  AlertTriangle,
  ArrowRight,
  ShieldCheck,
  Search,
  Users,
  Filter,
  RefreshCw,
  MoreVertical,
  Activity,
  Clock
} from 'lucide-react';
import { api, Transaction, Property, EscrowAccount, User } from '@/lib/api';

export default function AdminDashboard() {
  const [user, setUser] = useState<User | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [properties, setProperties] = useState<Record<string, Property>>({});
  const [escrows, setEscrows] = useState<Record<string, EscrowAccount>>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadData() {
      try {
        const currentUser = await api.getCurrentUser();
        setUser(currentUser);

        const allTx = await api.getTransactions();
        setTransactions(allTx);

        const propMap: Record<string, Property> = {};
        const escrowMap: Record<string, EscrowAccount> = {};

        for (const tx of allTx) {
          try {
            if (!propMap[tx.propertyId]) {
              const prop = await api.getProperty(tx.propertyId);
              propMap[tx.propertyId] = prop;
            }
          } catch (e) {
            console.error("Failed to load property", tx.propertyId);
          }
          try {
            const escrow = await api.getEscrow(tx.id);
            escrowMap[tx.id] = escrow;
          } catch (e) {
            console.error("Failed to load escrow", tx.id);
          }
        }

        setProperties(propMap);
        setEscrows(escrowMap);
      } catch (e) {
        console.error("Error loading dashboard data", e);
      } finally {
        setLoading(false);
      }
    }
    loadData();
  }, []);

  if (loading) {
    return <div className="flex h-full items-center justify-center text-on-surface-variant"><RefreshCw className="animate-spin w-8 h-8 text-primary"/></div>;
  }

  const activeTx = transactions.filter(tx => tx.status !== 'CLOSED' && tx.status !== 'CANCELLED');
  const closingSoon = activeTx.filter(tx => tx.status === 'ESCROW').length;

  // Calculate total escrow volume dynamically
  const totalEscrowVolume = activeTx.reduce((sum, tx) => {
    const esc = escrows[tx.id];
    return sum + (esc?.balance || 0);
  }, 0);

  return (
    <div className="flex-1 p-6 lg:p-8 max-w-7xl mx-auto w-full flex flex-col gap-8">
      {/* Header section */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-headline font-bold text-on-surface">Escrow Manager Dashboard</h1>
          <p className="text-on-surface-variant mt-1">Centralized oversight for all active real estate transactions.</p>
        </div>
        <div className="flex gap-3">
          <button className="flex items-center gap-2 bg-white border border-outline-variant text-on-surface font-medium px-4 py-2 rounded-lg hover:bg-surface-variant transition-colors shadow-sm">
            <Filter className="w-4 h-4" /> Filter
          </button>
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <Search className="w-4 h-4 text-outline-variant" />
            </div>
            <input
              type="text"
              placeholder="Search by ID or Address"
              className="pl-9 pr-4 py-2 border border-outline-variant rounded-lg text-sm bg-white focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent w-64 shadow-sm"
            />
          </div>
        </div>
      </div>

      {/* Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="bg-white p-6 rounded-xl border border-outline-variant shadow-sm flex flex-col gap-2 relative overflow-hidden">
          <p className="text-sm font-semibold text-on-surface-variant uppercase tracking-wider">Total Escrow Volume</p>
          <h2 className="text-3xl font-display font-bold text-on-surface">
            ${Math.floor(totalEscrowVolume).toLocaleString()}<span className="text-lg text-on-surface-variant font-medium">.{(totalEscrowVolume % 1).toFixed(2).substring(2) || "00"}</span>
          </h2>
          <div className="flex items-center gap-1 text-xs text-success font-semibold mt-2">
            <ShieldCheck className="w-4 h-4" /> Fully Secured
          </div>
        </div>

        <div className="bg-white p-6 rounded-xl border border-outline-variant shadow-sm flex flex-col gap-2 relative overflow-hidden">
          <p className="text-sm font-semibold text-on-surface-variant uppercase tracking-wider">Active Deals</p>
          <h2 className="text-3xl font-display font-bold text-on-surface">{activeTx.length}</h2>
          <div className="flex items-center gap-1 text-xs text-on-surface-variant mt-2 font-medium">
            <Building className="w-4 h-4 text-primary" /> Properties in Escrow
          </div>
        </div>

        <div className="bg-white p-6 rounded-xl border border-outline-variant shadow-sm flex flex-col gap-2 relative overflow-hidden">
          <p className="text-sm font-semibold text-on-surface-variant uppercase tracking-wider">Closing Soon</p>
          <h2 className="text-3xl font-display font-bold text-on-surface">{closingSoon}</h2>
          <div className="flex items-center gap-1 text-xs text-on-surface-variant mt-2 font-medium">
            <Clock className="w-4 h-4 text-warning" /> Closing within 7 days
          </div>
        </div>

        <div className="bg-error-container p-6 rounded-xl border border-[#f2b8b5] shadow-sm flex flex-col gap-2">
          <p className="text-sm font-semibold text-on-error-container uppercase tracking-wider">Active Disputes</p>
          <h2 className="text-3xl font-display font-bold text-on-error-container">0</h2>
          <div className="flex items-center gap-1 text-xs text-on-error-container font-semibold mt-2">
            <AlertTriangle className="w-4 h-4" /> Requires Admin Review
          </div>
        </div>
      </div>

      {/* Transactions Table */}
      <div className="bg-white rounded-xl border border-outline-variant shadow-sm overflow-hidden flex flex-col">
        <div className="p-6 border-b border-outline-variant flex justify-between items-center">
          <h3 className="font-headline font-semibold text-on-surface flex items-center gap-2">
            <Activity className="w-5 h-5 text-primary" /> Active Transactions
          </h3>
          <button className="text-on-surface-variant hover:bg-surface-variant p-2 rounded-full transition-colors">
            <MoreVertical className="w-5 h-5" />
          </button>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-surface-container-low text-on-surface-variant text-xs uppercase tracking-wider font-semibold border-b border-outline-variant">
                <th className="p-4 font-medium">Property & ID</th>
                <th className="p-4 font-medium">Parties</th>
                <th className="p-4 font-medium">Escrow Amount</th>
                <th className="p-4 font-medium">Status</th>
                <th className="p-4 font-medium text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="text-sm divide-y divide-outline-variant">

              {activeTx.map(tx => {
                const prop = properties[tx.propertyId];
                const esc = escrows[tx.id];

                return (
                  <tr key={tx.id} className="hover:bg-surface-container-lowest transition-colors group">
                    <td className="p-4">
                      <p className="font-semibold text-on-surface">{prop?.address || tx.propertyId}</p>
                      <p className="text-xs text-on-surface-variant mt-1">ID: {tx.id.substring(0,8)}...</p>
                    </td>
                    <td className="p-4">
                      <div className="flex flex-col gap-1 text-xs">
                        <span className="text-on-surface"><span className="font-semibold text-on-surface-variant w-12 inline-block">Buyer:</span> {tx.buyerId}</span>
                        <span className="text-on-surface"><span className="font-semibold text-on-surface-variant w-12 inline-block">Seller:</span> {tx.sellerId}</span>
                      </div>
                    </td>
                    <td className="p-4">
                      <p className="font-semibold text-on-surface">${esc?.balance?.toLocaleString() || 0}</p>
                    </td>
                    <td className="p-4">
                      {tx.status === 'ESCROW' ? (
                        <span className="inline-flex items-center gap-1 bg-success-container text-on-success-container text-xs px-2.5 py-1 rounded-full font-bold uppercase tracking-wider">
                          <CheckCircle className="w-3 h-3" /> Ready for Release
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1 bg-surface-container-high text-on-surface-variant text-xs px-2.5 py-1 rounded-full font-bold uppercase tracking-wider">
                          {tx.status}
                        </span>
                      )}
                    </td>
                    <td className="p-4 text-right">
                      {tx.status === 'ESCROW' ? (
                        <button className="bg-primary text-white text-xs font-semibold py-2 px-4 rounded-lg hover:opacity-90 transition-opacity shadow-sm">
                          Release Funds
                        </button>
                      ) : (
                        <button className="bg-white border border-outline-variant text-on-surface text-xs font-semibold py-2 px-4 rounded-lg hover:bg-surface-variant transition-colors shadow-sm">
                          View Details
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })}

              {activeTx.length === 0 && (
                <tr>
                  <td colSpan={5} className="p-8 text-center text-on-surface-variant border-dashed">
                    No active transactions found.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        <div className="p-4 border-t border-outline-variant bg-surface flex justify-between items-center text-sm text-on-surface-variant">
          Showing {activeTx.length > 0 ? 1 : 0} to {activeTx.length} of {activeTx.length} transactions
          <div className="flex gap-2">
            <button className="px-3 py-1.5 border border-outline-variant rounded-lg bg-surface-container-low text-outline-variant cursor-not-allowed">Previous</button>
            <button className="px-3 py-1.5 border border-outline-variant rounded-lg bg-white text-on-surface hover:bg-surface-variant transition-colors">Next</button>
          </div>
        </div>
      </div>
    </div>
  );
}
