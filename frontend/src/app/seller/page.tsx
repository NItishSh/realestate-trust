'use client';

import React, { useEffect, useState } from 'react';
import {
  Building,
  CreditCard,
  CheckCircle,
  Clock,
  ArrowRight,
  ShieldCheck,
  Eye,
  FileText,
  AlertCircle,
  Plus,
  RefreshCw,
  TrendingUp,
  MapPin,
  Home
} from 'lucide-react';
import { api, Transaction, Property, User } from '@/lib/api';

export default function SellerDashboard() {
  const [user, setUser] = useState<User | null>(null);
  const [properties, setProperties] = useState<Property[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadData() {
      try {
        const currentUser = await api.getCurrentUser();
        setUser(currentUser);

        if (currentUser) {
          const allProps = await api.getProperties();
          const myProps = allProps.filter(p => p.ownerId === currentUser.id || p.ownerId === currentUser.email);
          setProperties(myProps);

          const allTx = await api.getTransactions();
          const myTx = allTx.filter(tx => tx.sellerId === currentUser.id);
          setTransactions(myTx);
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
  const securedInEscrow = activeTx.filter(tx => tx.status === 'ESCROW').reduce((acc, tx) => acc + tx.totalAmount, 0);
  const pendingClosing = activeTx.filter(tx => tx.status !== 'ESCROW').reduce((acc, tx) => acc + tx.totalAmount, 0);
  const totalProceeds = securedInEscrow + pendingClosing;

  return (
    <div className="flex-1 p-6 lg:p-8 max-w-7xl mx-auto w-full flex flex-col gap-8">
      {/* Header section */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-headline font-bold text-on-surface">Seller Dashboard</h1>
          <p className="text-on-surface-variant mt-1">Manage your property listings and incoming escrow funds.</p>
        </div>
        <button className="flex items-center gap-2 bg-primary text-white font-medium px-4 py-2 rounded-lg hover:opacity-90 transition-opacity shadow-sm">
          <Plus className="w-4 h-4" /> List New Property
        </button>
      </div>

      {/* Bento Grid Layout */}
      <div className="grid grid-cols-1 md:grid-cols-12 gap-6">

        {/* Total Proceeds Overview */}
        <div className="col-span-1 md:col-span-8 bg-white rounded-xl border border-outline-variant p-6 shadow-sm flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-headline font-semibold text-on-surface">Total Potential Proceeds</h3>
              <div className="bg-success-container p-1.5 rounded-lg text-success">
                <ShieldCheck className="w-5 h-5" />
              </div>
            </div>
            <p className="text-4xl font-display font-bold text-on-surface tracking-tight">
              ${Math.floor(totalProceeds).toLocaleString()}<span className="text-lg text-on-surface-variant font-medium">.{(totalProceeds % 1).toFixed(2).substring(2) || "00"}</span>
            </p>
            <p className="text-sm text-success flex items-center gap-1 mt-2 font-medium">
              <CheckCircle className="w-4 h-4" /> FDIC Insured Funds
            </p>
          </div>
          <div className="mt-6 pt-6 border-t border-outline-variant grid grid-cols-2 gap-4">
            <div className="flex flex-col">
              <span className="text-xs text-on-surface-variant font-medium uppercase tracking-wider">Secured in Escrow</span>
              <span className="text-xl font-semibold text-on-surface">${securedInEscrow.toLocaleString()}</span>
            </div>
            <div className="flex flex-col">
              <span className="text-xs text-on-surface-variant font-medium uppercase tracking-wider">Pending Closing</span>
              <span className="text-xl font-semibold text-on-surface">${pendingClosing.toLocaleString()}</span>
            </div>
          </div>
        </div>

        {/* Action Required Card */}
        <div className="col-span-1 md:col-span-4 bg-white rounded-xl border border-outline-variant p-6 shadow-sm flex flex-col">
          <h3 className="font-headline font-semibold text-on-surface mb-4 flex items-center gap-2">
            <AlertCircle className="w-5 h-5 text-warning" />
            Action Required
          </h3>

          {activeTx.length > 0 ? (
            <div className="flex flex-col flex-1 justify-between gap-4">
              <div className="bg-warning-container rounded-lg p-4 border border-[#ffcc80]">
                <div className="flex items-start gap-3">
                  <Clock className="w-5 h-5 text-on-warning-container shrink-0 mt-0.5" />
                  <div>
                    <h4 className="font-semibold text-on-warning-container">Review Inspection Report</h4>
                    <p className="text-sm text-on-warning-container opacity-90 mt-1">Pending buyer approval for TX {activeTx[0].id.substring(0,8)}</p>
                  </div>
                </div>
              </div>
              <button className="bg-white border border-outline-variant text-on-surface font-medium px-4 py-2.5 rounded-lg hover:bg-surface-variant transition-colors w-full mt-auto shadow-sm">
                Review & Respond
              </button>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center flex-1 text-center bg-surface-container-low rounded-lg p-6 border border-outline-variant border-dashed">
              <CheckCircle className="w-10 h-10 text-success opacity-50 mb-3" />
              <p className="font-medium text-on-surface-variant">All up to date!</p>
              <p className="text-sm text-on-surface-variant mt-1">No pending actions required.</p>
            </div>
          )}
        </div>

        {/* Active Properties List */}
        <div className="col-span-1 md:col-span-8 bg-white rounded-xl border border-outline-variant shadow-sm overflow-hidden flex flex-col">
          <div className="p-6 border-b border-outline-variant flex justify-between items-center">
            <h3 className="font-headline font-semibold text-on-surface flex items-center gap-2">
              <Building className="w-5 h-5 text-primary" /> Active Properties
            </h3>
            <span className="text-sm text-on-surface-variant font-medium">{properties.length} Listings</span>
          </div>

          <div className="flex flex-col divide-y divide-outline-variant">
            {properties.map(prop => {
              const tx = activeTx.find(t => t.propertyId === prop.id);

              return (
                <div key={prop.id} className="p-6 hover:bg-surface-container-lowest transition-colors flex flex-col sm:flex-row gap-6 items-start sm:items-center justify-between group">
                  <div className="flex items-start gap-4 w-full sm:w-auto flex-1">
                    <div className="w-20 h-20 rounded-xl bg-surface-container-high flex items-center justify-center shrink-0 overflow-hidden border border-outline-variant">
                      {prop.thumbnail ? (
                         // eslint-disable-next-line @next/next/no-img-element
                        <img src={prop.thumbnail} alt={prop.address} className="w-full h-full object-cover" />
                      ) : (
                        <Home className="w-8 h-8 text-outline-variant" />
                      )}
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <h4 className="font-semibold text-on-surface text-lg leading-tight">{prop.address}</h4>
                      </div>

                      <div className="flex flex-wrap items-center gap-3 mt-2">
                        {tx ? (
                          <span className={`inline-flex items-center text-[10px] font-bold px-2.5 py-1 rounded-full uppercase tracking-wider ${tx.status === 'ESCROW' ? 'bg-success-container text-on-success-container' : 'bg-warning-container text-on-warning-container'}`}>
                            {tx.status === 'ESCROW' ? 'In Escrow' : 'Pending'}
                          </span>
                        ) : (
                          <span className="inline-flex items-center text-[10px] font-bold px-2.5 py-1 rounded-full uppercase tracking-wider bg-primary-container text-on-primary-container">
                            Active Listing
                          </span>
                        )}
                        <span className="text-sm font-semibold text-on-surface flex items-center gap-1">
                          <CreditCard className="w-3.5 h-3.5 text-on-surface-variant" />
                          ${prop.value?.toLocaleString()}
                        </span>
                      </div>

                      {tx ? (
                        <div className="mt-4 max-w-sm">
                          <div className="flex justify-between text-xs mb-1.5">
                            <span className="font-medium text-on-surface-variant">Progress: Buyer Inspection Review</span>
                            <span className="font-bold text-primary">40%</span>
                          </div>
                          <div className="w-full bg-surface-container-high rounded-full h-1.5 overflow-hidden">
                            <div className="bg-primary h-1.5 rounded-full" style={{ width: '40%' }}></div>
                          </div>
                        </div>
                      ) : (
                        <div className="mt-4 flex gap-5 text-sm">
                          <div className="flex items-center gap-1.5">
                            <Eye className="w-4 h-4 text-on-surface-variant" />
                            <span className="text-on-surface-variant"><span className="font-semibold text-on-surface">124</span> views</span>
                          </div>
                          <div className="flex items-center gap-1.5">
                            <FileText className="w-4 h-4 text-on-surface-variant" />
                            <span className="text-on-surface-variant"><span className="font-semibold text-on-surface">2</span> offers</span>
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                  <button className="shrink-0 bg-white border border-outline-variant text-on-surface p-2.5 rounded-lg hover:bg-surface-variant transition-colors shadow-sm self-start sm:self-center">
                    <ArrowRight className="w-5 h-5" />
                  </button>
                </div>
              );
            })}

            {properties.length === 0 && (
              <div className="p-10 text-center text-on-surface-variant bg-surface-container-low border-dashed">
                You have no active property listings.
              </div>
            )}
          </div>
        </div>

        {/* Right Column: Escrow Activity */}
        <div className="col-span-1 md:col-span-4 flex flex-col gap-6">
          <div className="bg-white rounded-xl border border-outline-variant shadow-sm flex flex-col">
            <div className="p-6 border-b border-outline-variant">
              <h3 className="font-headline font-semibold text-on-surface flex items-center gap-2">
                <Clock className="w-5 h-5 text-primary" /> Recent Activity
              </h3>
            </div>

            <div className="flex flex-col divide-y divide-outline-variant p-4">
              {activeTx.length > 0 ? (
                <>
                  <div className="flex gap-4 p-3 hover:bg-surface-container-lowest transition-colors rounded-lg group cursor-pointer">
                    <div className="w-10 h-10 rounded-full bg-success-container flex items-center justify-center text-on-success-container shrink-0">
                      <CheckCircle className="w-5 h-5" />
                    </div>
                    <div>
                      <p className="font-semibold text-on-surface group-hover:text-primary transition-colors">Earnest Money Received</p>
                      <p className="text-xs text-on-surface-variant mt-1">+$250,000 to Escrow Vault</p>
                    </div>
                  </div>
                  <div className="flex gap-4 p-3 hover:bg-surface-container-lowest transition-colors rounded-lg group cursor-pointer">
                    <div className="w-10 h-10 rounded-full bg-primary-container flex items-center justify-center text-primary shrink-0">
                      <FileText className="w-5 h-5" />
                    </div>
                    <div>
                      <p className="font-semibold text-on-surface group-hover:text-primary transition-colors">Offer Accepted</p>
                      <p className="text-xs text-on-surface-variant mt-1">TX {activeTx[0].id.substring(0,8)}</p>
                    </div>
                  </div>
                </>
              ) : (
                <div className="p-8 text-sm text-on-surface-variant text-center border border-dashed border-outline-variant rounded-lg m-2">
                  No recent activity found.
                </div>
              )}
            </div>
            <div className="p-4 border-t border-outline-variant bg-surface">
              <button className="w-full py-2.5 bg-white border border-outline-variant rounded-lg text-sm font-medium text-on-surface hover:bg-surface-variant transition-colors shadow-sm">
                View Complete History
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
