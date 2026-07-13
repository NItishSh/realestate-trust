'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useStore } from '../../lib/store';
import { ShieldCheck, Mail, Lock, ArrowRight, Loader2, User, Briefcase } from 'lucide-react';

export default function SignupPage() {
  const [formData, setFormData] = useState({
    fullName: '',
    email: '',
    password: '',
    role: 'BUYER' as 'BUYER' | 'SELLER' | 'BROKER' | 'OFFICER' | 'ADMIN'
  });
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const { registerUser, login, logAction } = useStore();

  const handleSignup = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      await registerUser(formData);
      // Automatically login after signup
      await login(formData.email, formData.password);

      // Log to ledger after we have a valid token
      await logAction(`User Registered: ${formData.fullName} (${formData.role})`);

      window.location.href = '/';
    } catch (err) {
      setError('Registration failed. Please try again.');
      setIsLoading(false);
    }
  };

  return (
    <div className="w-full">
      <div className="text-center mb-10">
        <div className="w-16 h-16 rounded-2xl bg-accent-cyan flex items-center justify-center shadow-xl shadow-accent-cyan/20 mx-auto mb-6">
          <ShieldCheck className="w-10 h-10 text-bg-primary" />
        </div>
        <h1 className="text-3xl font-bold tracking-tight mb-2">Create an Account</h1>
        <p className="text-slate-400">Join Trust RealEstate Portal</p>
      </div>

      <div className="glass-panel p-8 rounded-3xl relative overflow-hidden">
        {/* Decorative background glow */}
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full h-1/2 bg-accent-cyan/10 blur-3xl rounded-full -z-10 pointer-events-none" />

        <form onSubmit={handleSignup} className="space-y-5">
          {error && (
            <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <label className="text-sm font-semibold text-slate-300">Full Name</label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                <User className="h-5 w-5 text-slate-500" />
              </div>
              <input
                type="text"
                required
                className="w-full bg-slate-900/50 border border-card-border rounded-xl py-3 pl-12 pr-4 outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus-visible:border-accent-cyan focus:ring-1 focus:ring-accent-cyan transition-colors"
                placeholder="John Doe"
                value={formData.fullName}
                onChange={(e) => setFormData({ ...formData, fullName: e.target.value })}
              />
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-semibold text-slate-300">Email Address</label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                <Mail className="h-5 w-5 text-slate-500" />
              </div>
              <input
                type="email"
                required
                className="w-full bg-slate-900/50 border border-card-border rounded-xl py-3 pl-12 pr-4 outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus-visible:border-accent-cyan focus:ring-1 focus:ring-accent-cyan transition-colors"
                placeholder="investor@gmail.com"
                value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              />
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-semibold text-slate-300">Password</label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                <Lock className="h-5 w-5 text-slate-500" />
              </div>
              <input
                type="password"
                required
                className="w-full bg-slate-900/50 border border-card-border rounded-xl py-3 pl-12 pr-4 outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus-visible:border-accent-cyan focus:ring-1 focus:ring-accent-cyan transition-colors"
                placeholder="••••••••"
                value={formData.password}
                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              />
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-semibold text-slate-300">Role</label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                <Briefcase className="h-5 w-5 text-slate-500" />
              </div>
              <select
                className="w-full bg-slate-900/50 border border-card-border rounded-xl py-3 pl-12 pr-4 outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus-visible:border-accent-cyan focus:ring-1 focus:ring-accent-cyan transition-colors appearance-none"
                value={formData.role}
                onChange={(e) => setFormData({ ...formData, role: e.target.value as any })}
              >
                <option value="BUYER">Buyer (Investor)</option>
                <option value="SELLER">Seller (Property Owner)</option>
                <option value="BROKER">Broker</option>
                <option value="OFFICER">Compliance Officer</option>
                <option value="ADMIN">System Admin</option>
              </select>
            </div>
          </div>

          <button
            type="submit"
            disabled={isLoading || !formData.email || !formData.fullName || !formData.password}
            className="w-full bg-accent-cyan hover:bg-accent-cyan/90 text-bg-primary font-bold py-3 px-4 rounded-xl flex items-center justify-center gap-2 transition-colors transition-transform transition-opacity disabled:opacity-50 disabled:cursor-not-allowed mt-2"
          >
            {isLoading ? (
              <Loader2 className="w-5 h-5 animate-spin" />
            ) : (
              <>
                Create Account
                <ArrowRight className="w-5 h-5" />
              </>
            )}
          </button>
        </form>

        <div className="mt-8 text-center text-sm text-slate-400">
          Already have an account?{' '}
          <Link href="/login" className="text-accent-cyan hover:underline font-medium">
            Sign In
          </Link>
        </div>
      </div>
    </div>
  );
}
