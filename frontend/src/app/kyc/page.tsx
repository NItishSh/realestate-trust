'use client';

import React, { useState } from 'react';
import { useStore } from '../../lib/store';
import {
  FileCheck2,
  UserPlus,
  Upload,
  CheckCircle2,
  Clock,
  XCircle,
  Fingerprint,
  UserCheck
} from 'lucide-react';

export default function KYCPage() {
  const { currentUser, registerUser, submitKYC } = useStore();

  // Registration Form State
  const [email, setEmail] = useState('');
  const [fullName, setFullName] = useState('');
  const [role, setRole] = useState<'BUYER' | 'SELLER' | 'BROKER'>('BUYER');

  // KYC Form State
  const [docType, setDocType] = useState('PASSPORT');
  const [docRef, setDocRef] = useState('');

  const [regLoading, setRegLoading] = useState(false);
  const [kycLoading, setKycLoading] = useState(false);

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email || !fullName) return;
    setRegLoading(true);
    try {
      await registerUser({ email, fullName, role });
      setEmail('');
      setFullName('');
    } catch (e) {
      console.error(e);
    } finally {
      setRegLoading(false);
    }
  };

  const handleKYC = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentUser || !docRef) return;
    setKycLoading(true);
    try {
      await submitKYC(currentUser.id, docType, docRef);
      setDocRef('');
    } catch (e) {
      console.error(e);
    } finally {
      setKycLoading(false);
    }
  };

  const checklistItems = [
    { name: 'Profile Registration', desc: 'Account initialized', done: !!currentUser },
    { name: 'Identity Document Submission', desc: 'Secure ID/Passport details verified', done: !!(currentUser && currentUser.documentReference) },
    { name: 'KYC Verification Status', desc: 'Loki KYC engines consent checks passed', done: !!(currentUser && currentUser.kycStatus === 'APPROVED') },
    { name: 'Compliance Escrow Permission', desc: 'Virtual accounts and payment rails unlocked', done: !!(currentUser && currentUser.kycStatus === 'APPROVED') }
  ];

  return (
    <div>
      {/* Header */}
      <div className="mb-8">
        <h2 className="text-3xl font-extrabold tracking-tight mb-2">Compliance & Onboarding</h2>
        <p className="text-slate-400 text-sm">Register user identities and submit document verifications to fulfill compliance rules.</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left Grid: Forms */}
        <div className="flex flex-col gap-8 lg:col-span-1">
          {/* User Registration Form */}
          <div className="glass-panel p-6 rounded-2xl">
            <h3 className="text-md font-bold mb-4 flex items-center gap-2">
              <UserPlus className="w-5 h-5 text-accent-cyan" />
              Register User Account
            </h3>
            <form onSubmit={handleRegister} className="flex flex-col gap-4">
              <div>
                <label className="text-[10px] text-slate-500 font-mono block mb-1">FULL NAME</label>
                <input
                  type="text"
                  value={fullName}
                  onChange={e => setFullName(e.target.value)}
                  className="w-full bg-slate-900 border border-card-border rounded-xl px-4 py-2 text-xs focus:outline-none focus:border-accent-cyan font-sans"
                  placeholder="e.g. John Doe"
                  required
                />
              </div>
              <div>
                <label className="text-[10px] text-slate-500 font-mono block mb-1">EMAIL ADDRESS</label>
                <input
                  type="email"
                  value={email}
                  onChange={e => setEmail(e.target.value)}
                  className="w-full bg-slate-900 border border-card-border rounded-xl px-4 py-2 text-xs focus:outline-none focus:border-accent-cyan font-mono"
                  placeholder="name@domain.com"
                  required
                />
              </div>
              <div>
                <label className="text-[10px] text-slate-500 font-mono block mb-1">ACCOUNT ROLE</label>
                <select
                  value={role}
                  onChange={e => setRole(e.target.value as any)}
                  className="w-full bg-slate-900 border border-card-border rounded-xl px-4 py-2 text-xs focus:outline-none focus:border-accent-cyan cursor-pointer"
                >
                  <option value="BUYER">BUYER (Investor / Homeowner)</option>
                  <option value="SELLER">SELLER (Builder / Owner)</option>
                  <option value="BROKER">BROKER (Agent / Facilitator)</option>
                </select>
              </div>

              <button
                type="submit"
                disabled={regLoading}
                className="w-full bg-gradient-to-r from-accent-cyan to-accent-blue text-bg-primary font-bold text-xs py-3 rounded-xl hover:opacity-90 active:scale-95 transition-all flex items-center justify-center gap-2 cursor-pointer"
              >
                {regLoading ? 'Registering...' : 'Register Account'}
              </button>
            </form>
          </div>

          {/* KYC Upload Form */}
          {currentUser && (
            <div className="glass-panel p-6 rounded-2xl">
              <h3 className="text-md font-bold mb-4 flex items-center gap-2">
                <Upload className="w-5 h-5 text-accent-cyan" />
                Submit KYC Document
              </h3>
              <form onSubmit={handleKYC} className="flex flex-col gap-4">
                <div>
                  <label className="text-[10px] text-slate-500 font-mono block mb-1">DOCUMENT TYPE</label>
                  <select
                    value={docType}
                    onChange={e => setDocType(e.target.value)}
                    className="w-full bg-slate-900 border border-card-border rounded-xl px-4 py-2 text-xs focus:outline-none focus:border-accent-cyan cursor-pointer"
                  >
                    <option value="PASSPORT">PASSPORT</option>
                    <option value="DRIVERS_LICENSE">DRIVER'S LICENSE</option>
                    <option value="PAN_CARD">PAN CARD</option>
                  </select>
                </div>
                <div>
                  <label className="text-[10px] text-slate-500 font-mono block mb-1">DOCUMENT REFERENCE / NUMBER</label>
                  <input
                    type="text"
                    value={docRef}
                    onChange={e => setDocRef(e.target.value)}
                    className="w-full bg-slate-900 border border-card-border rounded-xl px-4 py-2 text-xs focus:outline-none focus:border-accent-cyan font-mono"
                    placeholder="e.g. P120389A"
                    required
                  />
                </div>

                <button
                  type="submit"
                  disabled={kycLoading || !!currentUser.documentReference}
                  className={`w-full font-bold text-xs py-3 rounded-xl transition-all flex items-center justify-center gap-2 cursor-pointer ${
                    currentUser.documentReference
                      ? 'bg-slate-800 text-slate-500 border border-card-border cursor-not-allowed'
                      : 'bg-gradient-to-r from-accent-cyan to-accent-blue text-bg-primary hover:opacity-90 active:scale-95'
                  }`}
                >
                  {kycLoading ? 'Submitting...' : currentUser.documentReference ? 'KYC Already Submitted' : 'Submit Verification'}
                </button>
              </form>
            </div>
          )}
        </div>

        {/* Right Grid: Verification Checklist */}
        <div className="lg:col-span-2">
          {currentUser ? (
            <div className="glass-panel p-8 rounded-2xl h-full flex flex-col justify-between">
              <div>
                <div className="flex justify-between items-start border-b border-card-border pb-6 mb-6">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-xl bg-slate-850 flex items-center justify-center border border-card-border">
                      <Fingerprint className="w-5 h-5 text-accent-cyan" />
                    </div>
                    <div>
                      <span className="text-[10px] font-mono text-slate-500 uppercase">Verification Check</span>
                      <h3 className="text-lg font-bold mt-0.5">{currentUser.fullName}</h3>
                    </div>
                  </div>
                  <div>
                    <span className="text-[10px] text-slate-500 font-mono block text-right">KYC STATUS</span>
                    <span className={`text-xs font-mono font-bold px-3 py-1 rounded-full inline-block mt-1 ${
                      currentUser.kycStatus === 'APPROVED' ? 'bg-accent-emerald/10 text-accent-emerald' :
                      currentUser.kycStatus === 'PENDING' ? 'bg-accent-blue/10 text-accent-blue' :
                      'bg-rose-500/10 text-rose-500'
                    }`}>
                      {currentUser.kycStatus || 'NOT_SUBMITTED'}
                    </span>
                  </div>
                </div>

                <h4 className="text-xs font-bold text-slate-400 uppercase tracking-wide mb-6">Onboarding Checkpoint Checklist</h4>

                <div className="flex flex-col gap-6">
                  {checklistItems.map((item, idx) => (
                    <div key={item.name} className="flex gap-4 items-start">
                      <div className={`w-6 h-6 rounded-full flex items-center justify-center shrink-0 border mt-0.5 ${
                        item.done
                          ? 'bg-accent-emerald/20 border-accent-emerald text-accent-emerald'
                          : 'bg-slate-900 border-card-border text-slate-600'
                      }`}>
                        {item.done ? <CheckCircle2 className="w-4 h-4 text-accent-emerald" /> : idx + 1}
                      </div>
                      <div>
                        <p className={`text-sm font-bold ${
                          item.done ? 'text-slate-200' : 'text-slate-500'
                        }`}>{item.name}</p>
                        <span className="text-xs text-slate-500 leading-tight mt-0.5 block">{item.desc}</span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {currentUser.kycStatus === 'PENDING' && (
                <div className="border-t border-card-border pt-6 mt-6 flex justify-between items-center gap-4">
                  <div className="flex items-center gap-2 text-accent-blue text-xs font-mono">
                    <Clock className="w-5 h-5 animate-pulse shrink-0" />
                    Review Pending: Compliance Officer verification check in progress.
                  </div>
                  <button
                    onClick={() => {
                      // Autonomously approve user for local demo ease
                      useStore.setState((state) => ({
                        users: state.users.map(u => u.id === currentUser.id ? { ...u, kycStatus: 'APPROVED' } : u),
                        currentUser: state.currentUser && state.currentUser.id === currentUser.id ?
                          { ...state.currentUser, kycStatus: 'APPROVED' } : state.currentUser
                      }));
                    }}
                    className="px-4 py-2 rounded-lg bg-accent-emerald/10 border border-accent-emerald/20 text-accent-emerald text-[11px] font-bold hover:bg-accent-emerald/20 transition-colors cursor-pointer"
                  >
                    Simulate Approve
                  </button>
                </div>
              )}

              {currentUser.kycStatus === 'APPROVED' && (
                <div className="border-t border-card-border pt-6 mt-6 flex items-center gap-2 text-accent-emerald text-xs font-mono">
                  <UserCheck className="w-5 h-5 shrink-0" />
                  KYC Verified: Virtual ledger operations unlocked.
                </div>
              )}
            </div>
          ) : (
            <div className="glass-panel p-8 rounded-2xl h-full flex flex-col items-center justify-center text-slate-500 gap-2">
              <Clock className="w-8 h-8 text-slate-600" />
              <p className="text-sm font-mono">Please register or switch to an active user profile.</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
