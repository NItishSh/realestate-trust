'use client';

import React, { useEffect, useState } from 'react';
import {
  Building,
  CreditCard,
  CheckCircle,
  Clock,
  ArrowRight,
  ShieldCheck,
  AlertCircle,
  Download,
  TrendingUp,
  PenTool,
  Search,
  Key,
  RefreshCw
} from 'lucide-react';
import { api, Transaction, Property, EscrowAccount, User } from '@/lib/api';

export default function BuyerDashboard() {
  const [user, setUser] = useState<User | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [properties, setProperties] = useState<Record<string, Property>>({});
  const [escrows, setEscrows] = useState<Record<string, EscrowAccount>>({});
  const [loading, setLoading] = useState(true);

  // Modal states
  const [isStatementOpen, setIsStatementOpen] = useState(false);
  const [isTransferOpen, setIsTransferOpen] = useState(false);
  const [isReviewOpen, setIsReviewOpen] = useState(false);

  useEffect(() => {
    async function loadData() {
      try {
        const currentUser = await api.getCurrentUser();
        setUser(currentUser);

        if (currentUser) {
          const allTx = await api.getTransactions();
          const myTx = allTx.filter(tx => tx.buyerId === currentUser.id);
          setTransactions(myTx);

          const propMap: Record<string, Property> = {};
          const escrowMap: Record<string, EscrowAccount> = {};

          for (const tx of myTx) {
            try {
              const prop = await api.getProperty(tx.propertyId);
              propMap[tx.propertyId] = prop;
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
        }
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

  // Calculate total escrow value for this buyer's active transactions
  const totalEscrowValue = activeTx.reduce((sum, tx) => {
    const esc = escrows[tx.id];
    return sum + (esc?.balance || 0);
  }, 0);

  // Find a transaction that requires funding
  const actionTx = activeTx.find(tx => tx.status === 'ESCROW' || tx.status === 'DRAFT');
  const actionProp = actionTx ? properties[actionTx.propertyId] : null;

  return (
    <div className="flex-1 p-6 lg:p-8 max-w-7xl mx-auto w-full flex flex-col gap-8">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-3xl font-headline font-bold text-on-surface">Buyer Dashboard</h2>
          <p className="text-on-surface-variant mt-1">Welcome back, your escrow vault is secure.</p>
        </div>
        <button onClick={() => setIsStatementOpen(true)} className="flex items-center gap-2 bg-primary-container text-primary font-medium px-4 py-2 rounded-lg hover:bg-opacity-80 transition-opacity">
          <Download className="w-4 h-4" />
          Statement
        </button>
      </div>

      {/* Bento Grid Layout */}
      <div className="grid grid-cols-1 md:grid-cols-12 gap-6">

        {/* Active Escrow Overview */}
        <div className="col-span-1 md:col-span-4 bg-white rounded-xl border border-outline-variant p-6 shadow-sm flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-headline font-semibold text-on-surface">Total in Escrow</h3>
              <div className="bg-primary-container p-1.5 rounded-lg text-primary">
                <ShieldCheck className="w-5 h-5" />
              </div>
            </div>
            <p className="text-4xl font-display font-bold text-on-surface tracking-tight">
              ${Math.floor(totalEscrowValue).toLocaleString()}<span className="text-lg text-on-surface-variant font-medium">.{(totalEscrowValue % 1).toFixed(2).substring(2) || "00"}</span>
            </p>
            <p className="text-sm text-success flex items-center gap-1 mt-2 font-medium">
              <TrendingUp className="w-4 h-4" />
              +2.4% vs last month
            </p>
          </div>
          <div className="mt-6 pt-6 border-t border-outline-variant flex justify-between items-center">
            <div className="flex flex-col">
              <span className="text-xs text-on-surface-variant font-medium uppercase tracking-wider">Active Transactions</span>
              <span className="text-xl font-semibold text-on-surface">{activeTx.length}</span>
            </div>
            <div className="flex flex-col items-end">
              <span className="text-xs text-on-surface-variant font-medium uppercase tracking-wider">FDIC Insured</span>
              <span className="text-success font-medium text-sm flex items-center gap-1"><CheckCircle className="w-3 h-3" /> Verified</span>
            </div>
          </div>
        </div>

        {/* Pending Actions */}
        <div className="col-span-1 md:col-span-8 bg-white rounded-xl border border-outline-variant p-6 shadow-sm">
          <h3 className="font-headline font-semibold text-on-surface mb-4 flex items-center gap-2">
            <AlertCircle className="w-5 h-5 text-warning" />
            Action Required
          </h3>
          <div className="space-y-4">

            {actionTx && actionProp ? (
              <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center p-4 rounded-lg bg-error-container border border-[#f2b8b5]">
                <div className="flex items-start gap-4 mb-3 sm:mb-0">
                  <div className="bg-white p-2 rounded-full mt-1">
                    <CreditCard className="w-5 h-5 text-on-error-container" />
                  </div>
                  <div>
                    <h4 className="font-semibold text-on-error-container">Fund Escrow: {actionProp.address}</h4>
                    <p className="text-sm text-on-error-container opacity-90 mt-1">Initial deposit of ${actionTx.totalAmount.toLocaleString()} due.</p>
                  </div>
                </div>
                <button onClick={() => setIsTransferOpen(true)} className="bg-on-error-container text-white px-5 py-2 rounded-lg font-medium hover:opacity-90 transition-opacity w-full sm:w-auto">
                  Transfer Funds
                </button>
              </div>
            ) : (
              <div className="p-4 rounded-lg bg-success-container text-on-success-container border border-[#c4eed0]">
                All up to date! No pending actions required.
              </div>
            )}

            {/* Action 2: Mocked for layout consistency */}
            {activeTx.length > 0 && (
              <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center p-4 rounded-lg bg-surface-container-low border border-outline-variant">
                <div className="flex items-start gap-4 mb-3 sm:mb-0">
                  <div className="bg-white p-2 rounded-full mt-1 shadow-sm border border-outline-variant">
                    <PenTool className="w-5 h-5 text-on-surface-variant" />
                  </div>
                  <div>
                    <h4 className="font-semibold text-on-surface">Approve Title Transfer</h4>
                    <p className="text-sm text-on-surface-variant mt-1">Review and sign the preliminary title report.</p>
                  </div>
                </div>
                <button onClick={() => setIsReviewOpen(true)} className="bg-white border border-outline-variant text-on-surface px-5 py-2 rounded-lg font-medium hover:bg-surface-variant transition-colors w-full sm:w-auto">
                  Review Document
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Active Property Timeline */}
        {activeTx.map(tx => {
          const prop = properties[tx.propertyId];
          if (!prop) return null;

          return (
            <div key={tx.id} className="col-span-1 md:col-span-12 bg-white rounded-xl border border-outline-variant p-6 shadow-sm">
              <div className="flex justify-between items-end mb-6">
                <div>
                  <h3 className="font-headline font-semibold text-on-surface">Transaction Timeline</h3>
                  <p className="text-on-surface-variant mt-1">{prop.address}</p>
                </div>
                <span className="bg-primary-container text-primary text-xs font-bold px-2.5 py-1 rounded-full">{tx.status}</span>
              </div>
              <div className="relative pt-4 pb-2 overflow-x-auto">
                <div className="absolute top-10 left-8 right-8 h-1 bg-surface-container-high rounded-full -z-10 hidden sm:block"></div>
                <div className="absolute top-10 left-8 h-1 bg-primary rounded-full -z-10 hidden sm:block" style={{width: tx.status === 'ESCROW' ? '50%' : '25%'}}></div>
                <div className="flex flex-col sm:flex-row justify-between gap-6 min-w-max sm:min-w-0">

                  {/* Step 1: Complete */}
                  <div className="flex sm:flex-col items-center gap-4 sm:gap-2 w-full sm:w-1/5 relative">
                    <div className="w-8 h-8 rounded-full bg-primary text-white flex items-center justify-center shrink-0 z-10">
                      <CheckCircle className="w-4 h-4 font-bold" />
                    </div>
                    <div className="sm:text-center">
                      <h4 className="font-medium text-sm text-on-surface">Offer Accepted</h4>
                      <p className="text-xs text-on-surface-variant mt-0.5">Approved</p>
                    </div>
                    <div className="absolute left-4 top-8 bottom-[-24px] w-[2px] bg-primary -z-10 sm:hidden"></div>
                  </div>

                  {/* Step 2: Complete */}
                  <div className="flex sm:flex-col items-center gap-4 sm:gap-2 w-full sm:w-1/5 relative">
                    <div className="w-8 h-8 rounded-full bg-primary text-white flex items-center justify-center shrink-0 z-10">
                      <CheckCircle className="w-4 h-4 font-bold" />
                    </div>
                    <div className="sm:text-center">
                      <h4 className="font-medium text-sm text-on-surface">Inspection</h4>
                      <p className="text-xs text-on-surface-variant mt-0.5">Cleared</p>
                    </div>
                    <div className="absolute left-4 top-8 bottom-[-24px] w-[2px] bg-primary -z-10 sm:hidden"></div>
                  </div>

                  {/* Step 3: Current */}
                  <div className="flex sm:flex-col items-center gap-4 sm:gap-2 w-full sm:w-1/5 relative">
                    <div className={`w-8 h-8 rounded-full flex items-center justify-center shrink-0 z-10 ${tx.status === 'ESCROW' ? 'bg-primary text-white' : 'bg-white border-2 border-primary text-primary ring-4 ring-primary-container'}`}>
                      {tx.status === 'ESCROW' ? <CheckCircle className="w-4 h-4 font-bold" /> : <CreditCard className="w-4 h-4" />}
                    </div>
                    <div className="sm:text-center">
                      <h4 className={`font-bold text-sm ${tx.status === 'ESCROW' ? 'text-on-surface' : 'text-primary'}`}>Escrow Funded</h4>
                      <p className={`text-xs mt-0.5 ${tx.status === 'ESCROW' ? 'text-on-surface-variant' : 'text-primary font-medium'}`}>
                        {tx.status === 'ESCROW' ? 'Funded' : 'Pending Action'}
                      </p>
                    </div>
                    <div className={`absolute left-4 top-8 bottom-[-24px] w-[2px] ${tx.status === 'ESCROW' ? 'bg-primary' : 'bg-surface-container-high'} -z-10 sm:hidden`}></div>
                  </div>

                  {/* Step 4: Upcoming */}
                  <div className="flex sm:flex-col items-center gap-4 sm:gap-2 w-full sm:w-1/5 relative">
                    <div className="w-8 h-8 rounded-full bg-surface-container-high text-outline-variant flex items-center justify-center shrink-0 z-10 border border-outline-variant">
                      <Search className="w-4 h-4" />
                    </div>
                    <div className="sm:text-center opacity-50">
                      <h4 className="font-medium text-sm text-on-surface">Title Search</h4>
                      <p className="text-xs text-on-surface-variant mt-0.5">Upcoming</p>
                    </div>
                    <div className="absolute left-4 top-8 bottom-[-24px] w-[2px] bg-surface-container-high -z-10 sm:hidden"></div>
                  </div>

                  {/* Step 5: Upcoming */}
                  <div className="flex sm:flex-col items-center gap-4 sm:gap-2 w-full sm:w-1/5 relative">
                    <div className="w-8 h-8 rounded-full bg-surface-container-high text-outline-variant flex items-center justify-center shrink-0 z-10 border border-outline-variant">
                      <Key className="w-4 h-4" />
                    </div>
                    <div className="sm:text-center opacity-50">
                      <h4 className="font-medium text-sm text-on-surface">Closing</h4>
                      <p className="text-xs text-on-surface-variant mt-0.5">Est. {new Date().toLocaleDateString()}</p>
                    </div>
                  </div>

                </div>
              </div>
            </div>
          )
        })}

        {activeTx.length === 0 && (
          <div className="col-span-1 md:col-span-12 p-8 text-center text-gray-500 bg-white rounded-xl border border-outline-variant shadow-sm border-dashed">
            You have no active property purchases at this time.
          </div>
        )}

        {/* Recent Transactions */}
        <div className="col-span-1 md:col-span-12 bg-white rounded-xl border border-outline-variant shadow-sm overflow-hidden">
          <div className="p-6 border-b border-outline-variant flex justify-between items-center">
            <h3 className="font-headline font-semibold text-on-surface">Recent Vault Activity</h3>
            <a className="text-primary text-sm font-medium hover:underline" href="#">View All</a>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="bg-surface-container-low text-on-surface-variant text-xs uppercase tracking-wider font-semibold border-b border-outline-variant">
                  <th className="p-4 font-medium">Date</th>
                  <th className="p-4 font-medium">Description</th>
                  <th className="p-4 font-medium">Status</th>
                  <th className="p-4 font-medium text-right">Amount</th>
                </tr>
              </thead>
              <tbody className="text-sm divide-y divide-outline-variant">
                {activeTx.map(tx => {
                  const esc = escrows[tx.id];
                  const prop = properties[tx.propertyId];
                  if (!esc || esc.balance <= 0) return null;
                  return (
                    <tr key={"act-" + tx.id} className="hover:bg-surface-container-lowest transition-colors group">
                      <td className="p-4 text-on-surface-variant">{new Date().toLocaleDateString()}</td>
                      <td className="p-4 font-medium text-on-surface">Earnest Money Deposit - {prop?.address || tx.propertyId}</td>
                      <td className="p-4">
                        <span className="inline-flex items-center gap-1 bg-success-container text-on-success-container text-xs px-2 py-1 rounded-md font-medium">
                          <CheckCircle className="w-3 h-3" /> Verified
                        </span>
                      </td>
                      <td className="p-4 text-right font-medium text-on-surface">+${esc.balance.toLocaleString()}.00</td>
                    </tr>
                  )
                })}
                <tr className="hover:bg-surface-container-lowest transition-colors group">
                  <td className="p-4 text-on-surface-variant">{new Date().toLocaleDateString()}</td>
                  <td className="p-4 font-medium text-on-surface">System Verification</td>
                  <td className="p-4">
                    <span className="inline-flex items-center gap-1 bg-success-container text-on-success-container text-xs px-2 py-1 rounded-md font-medium">
                      <CheckCircle className="w-3 h-3" /> Cleared
                    </span>
                  </td>
                  <td className="p-4 text-right font-medium text-on-surface">--</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
      {/* Modals */}
      {isStatementOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6" id="statement-modal">
          <div className="absolute inset-0 bg-black opacity-50" onClick={() => setIsStatementOpen(false)}></div>
          <div className="relative bg-white w-full max-w-2xl rounded-xl shadow-sm overflow-hidden flex flex-col max-h-[90vh]">
            <div className="px-6 py-4 border-b border-outline-variant flex justify-between items-center bg-surface">
              <h3 className="text-xl font-headline font-bold text-on-surface">Escrow Account Statement</h3>
              <button onClick={() => setIsStatementOpen(false)} className="text-on-surface-variant hover:bg-surface-container-high p-1 rounded-full transition-colors">
                <AlertCircle className="w-5 h-5" />
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-6 space-y-6">
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div className="p-4 bg-primary-container rounded-lg">
                  <p className="text-xs text-on-primary-container font-medium uppercase tracking-wider mb-1">Total Balance</p>
                  <p className="text-xl font-bold text-on-primary-container">${Math.floor(totalEscrowValue).toLocaleString()}.00</p>
                </div>
                <div className="p-4 bg-success-container rounded-lg">
                  <p className="text-xs text-on-success-container font-medium uppercase tracking-wider mb-1">Deposits (30d)</p>
                  <p className="text-xl font-bold text-on-success-container">+${(totalEscrowValue * 0.1).toLocaleString()}</p>
                </div>
                <div className="p-4 bg-surface-container-high rounded-lg">
                  <p className="text-xs text-on-surface-variant font-medium uppercase tracking-wider mb-1">Withdrawals (30d)</p>
                  <p className="text-xl font-bold text-on-surface">-$0.00</p>
                </div>
              </div>
              <div>
                <h4 className="font-headline font-semibold text-on-surface mb-3">Detailed Transactions</h4>
                <div className="border border-outline-variant rounded-lg overflow-hidden">
                  <table className="w-full text-left border-collapse text-sm">
                    <thead>
                      <tr className="bg-surface-container-low text-on-surface-variant text-xs uppercase font-semibold border-b border-outline-variant">
                        <th className="p-3">Date</th>
                        <th className="p-3">Description</th>
                        <th className="p-3">Type</th>
                        <th className="p-3 text-right">Amount</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-outline-variant">
                      <tr>
                        <td className="p-3 text-on-surface-variant">{new Date().toLocaleDateString()}</td>
                        <td className="p-3 font-medium">Earnest Money Deposit</td>
                        <td className="p-3"><span className="text-success">Deposit</span></td>
                        <td className="p-3 text-right font-medium">+${Math.floor(totalEscrowValue).toLocaleString()}.00</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
            <div className="px-6 py-4 border-t border-outline-variant bg-surface flex flex-col sm:flex-row justify-end gap-3">
              <button onClick={() => setIsStatementOpen(false)} className="px-5 py-2 border border-outline-variant text-on-surface font-medium rounded-lg hover:bg-surface-variant transition-colors">
                Close
              </button>
              <button onClick={() => setIsStatementOpen(false)} className="px-5 py-2 bg-primary text-white font-medium rounded-lg hover:opacity-90 transition-opacity flex items-center justify-center gap-2">
                <Download className="w-4 h-4" />
                Download as PDF
              </button>
            </div>
          </div>
        </div>
      )}

      {isTransferOpen && actionProp && actionTx && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6" id="transfer-funds-modal">
          <div className="absolute inset-0 bg-black opacity-50" onClick={() => setIsTransferOpen(false)}></div>
          <div className="relative bg-white w-full max-w-lg rounded-xl shadow-sm overflow-hidden flex flex-col max-h-[90vh]">
            <div className="px-6 py-4 border-b border-outline-variant flex justify-between items-center bg-surface">
              <h3 className="text-xl font-headline font-bold text-on-surface">Transfer Funds to Escrow</h3>
              <button onClick={() => setIsTransferOpen(false)} className="text-on-surface-variant hover:bg-surface-container-high p-1 rounded-full transition-colors">
                <AlertCircle className="w-5 h-5" />
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-6 space-y-6">
              <div className="bg-primary-container p-4 rounded-lg mb-6">
                <div className="flex justify-between items-center mb-1">
                  <span className="text-xs text-on-primary-container font-medium uppercase tracking-wider">Required Deposit</span>
                  <span className="text-xs text-on-primary-container opacity-70">Due in 48h</span>
                </div>
                <p className="text-3xl font-display font-bold text-on-primary-container">${actionTx.totalAmount.toLocaleString()}.00</p>
                <p className="text-sm text-on-primary-container mt-2 flex items-center gap-1">
                  {actionProp.address}
                </p>
              </div>
              <div className="space-y-4">
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium text-on-surface-variant">Select Funding Source</label>
                  <select className="w-full bg-surface-container-low border-outline-variant rounded-lg px-4 py-2.5 text-sm outline-none">
                    <option>Chase Checking (...4829)</option>
                    <option>Wire Transfer</option>
                  </select>
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium text-on-surface-variant">Routing Number</label>
                  <input type="text" placeholder="Enter routing number" className="w-full border border-outline-variant rounded-lg px-4 py-2.5 text-sm outline-none" />
                </div>
              </div>
            </div>
            <div className="px-6 py-4 border-t border-outline-variant bg-surface-container-low flex justify-end gap-3">
              <button onClick={() => setIsTransferOpen(false)} className="px-6 py-2 border border-outline-variant text-on-surface font-medium rounded-lg hover:bg-surface-variant transition-colors">
                Cancel
              </button>
              <button onClick={() => setIsTransferOpen(false)} className="px-6 py-2 bg-primary text-white font-medium rounded-lg hover:opacity-90 transition-opacity">
                Transfer Now
              </button>
            </div>
          </div>
        </div>
      )}

      {isReviewOpen && actionProp && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center p-4 sm:p-6">
          <div className="absolute inset-0 bg-on-surface opacity-50" onClick={() => setIsReviewOpen(false)}></div>
          <div className="relative bg-white w-full max-w-3xl max-h-[90vh] rounded-xl shadow-2xl flex flex-col overflow-hidden">
            <div className="flex items-center justify-between px-6 py-4 border-b border-outline-variant">
              <h3 className="text-xl font-headline font-bold text-on-surface">Approve Title Transfer</h3>
              <button onClick={() => setIsReviewOpen(false)} className="text-on-surface-variant hover:bg-surface-variant p-2 rounded-full transition-colors">
                <AlertCircle className="w-5 h-5" />
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-6 space-y-6">
              <div className="bg-surface-container-low border border-outline-variant rounded-lg p-8 shadow-inner">
                <div className="max-w-2xl mx-auto bg-white p-10 shadow-sm border border-outline-variant min-h-[600px] flex flex-col gap-4">
                  <div className="flex justify-between items-start border-b pb-4">
                    <div className="space-y-1">
                      <h4 className="font-bold text-lg uppercase tracking-tight">Preliminary Title Report</h4>
                      <p className="text-xs text-on-surface-variant">Report No: TR-{actionProp.id.substring(0,6).toUpperCase()}-2023</p>
                    </div>
                    <div className="text-right text-xs text-on-surface-variant">
                      <p>Date: {new Date().toLocaleDateString()}</p>
                      <p>Property: {actionProp.address}</p>
                    </div>
                  </div>
                  <div className="space-y-4 mt-4">
                    <div className="h-4 bg-surface-container-high rounded w-3/4"></div>
                    <div className="h-4 bg-surface-container-high rounded w-full"></div>
                    <div className="h-4 bg-surface-container-high rounded w-5/6"></div>
                    <div className="h-4 bg-surface-container-high rounded w-full"></div>
                    <div className="pt-4">
                      <div className="h-6 bg-surface-container-high rounded w-1/3 mb-2"></div>
                      <div className="h-4 bg-surface-container-high rounded w-full"></div>
                      <div className="h-4 bg-surface-container-high rounded w-full"></div>
                    </div>
                    <div className="pt-4">
                      <div className="h-6 bg-surface-container-high rounded w-1/4 mb-2"></div>
                      <div className="h-4 bg-surface-container-high rounded w-full"></div>
                      <div className="h-4 bg-surface-container-high rounded w-2/3"></div>
                    </div>
                  </div>
                  <div className="mt-auto pt-8 border-t border-dashed border-outline-variant">
                    <p className="text-[10px] text-on-surface-variant italic text-center">This document is a preliminary report and is subject to final verification by TrustEstate legal counsel.</p>
                  </div>
                </div>
              </div>
              <div className="space-y-4">
                <h4 className="font-headline font-semibold text-on-surface">Digital Signature</h4>
                <div className="space-y-2">
                  <label className="text-sm font-medium text-on-surface-variant">Type your full name to sign</label>
                  <input className="w-full border border-outline-variant rounded-lg px-4 py-2.5 focus:ring-2 focus:ring-primary focus:border-primary outline-none" placeholder="John D. Smith" type="text" />
                </div>
                <label className="flex items-start gap-3 cursor-pointer group">
                  <input className="mt-1 rounded border-outline-variant text-primary focus:ring-primary" type="checkbox" />
                  <span className="text-sm text-on-surface-variant group-hover:text-on-surface transition-colors">I agree to the Electronic Record and Signature Disclosure and understand that my typed name constitutes a legal signature.</span>
                </label>
              </div>
            </div>
            <div className="px-6 py-4 border-t border-outline-variant bg-surface-container-low flex flex-col sm:flex-row justify-end gap-3">
              <button onClick={() => setIsReviewOpen(false)} className="px-6 py-2 border border-outline-variant text-on-surface font-medium rounded-lg hover:bg-surface-variant transition-colors">Cancel</button>
              <button onClick={() => setIsReviewOpen(false)} className="px-6 py-2 bg-primary text-white font-medium rounded-lg hover:opacity-90 transition-opacity">Approve & Sign</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
