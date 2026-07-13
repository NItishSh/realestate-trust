'use client';

import React, { useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useStore } from '../lib/store';
import {
  LayoutDashboard,
  ArrowLeftRight,
  FileCheck,
  PieChart,
  Database,
  User as UserIcon,
  ShieldCheck,
  RefreshCw,
  Search,
  LogOut
} from 'lucide-react';
import './globals.css';

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const { currentUser, users, fetchInitialData, logout } = useStore();
  const isAuthRoute = pathname === '/login' || pathname === '/signup';

  useEffect(() => {
    fetchInitialData();
  }, [fetchInitialData]);

  const navItems = [
    { name: 'Dashboard', path: '/', icon: LayoutDashboard },
    { name: 'Marketplace', path: '/marketplace', icon: Search },
    { name: 'Escrow Accounts', path: '/transactions', icon: ArrowLeftRight },
    { name: 'KYC Onboarding', path: '/kyc', icon: FileCheck },
    { name: 'Fractional Pools', path: '/portfolio', icon: PieChart },
    { name: 'Ledger Logs', path: '/ledger', icon: Database },
  ];

  return (
    <html lang="en">
      <body className="flex h-screen bg-bg-primary text-foreground font-sans">
        {!isAuthRoute && (
          <aside className="w-64 bg-bg-secondary border-r border-card-border flex flex-col justify-between shrink-0">
          <div>
            {/* Header */}
            <div className="p-6 border-b border-card-border flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-accent-cyan flex items-center justify-center shadow-lg shadow-accent-cyan/20">
                <ShieldCheck className="w-5 h-5 text-bg-primary" />
              </div>
              <div>
                <h1 className="font-bold text-sm tracking-wide uppercase">Trust RealEstate</h1>
                <span className="text-xs text-accent-cyan font-mono">Portal</span>
              </div>
            </div>

            {/* Navigation links */}
            <nav className="p-4 flex flex-col gap-1">
              {navItems.map((item) => {
                const Icon = item.icon;
                const isActive = pathname === item.path;
                return (
                  <Link
                    key={item.name}
                    href={item.path}
                    className={`flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 ${
                      isActive
                        ? 'bg-accent-cyan/10 text-accent-cyan font-semibold border-l-2 border-accent-cyan'
                        : 'text-slate-400 hover:bg-slate-800/40 hover:text-foreground'
                    }`}
                  >
                    <Icon className="w-5 h-5" />
                    <span className="text-sm">{item.name}</span>
                  </Link>
                );
              })}
            </nav>
          </div>

          {/* Footer User Profile Card */}
          <div className="p-4 border-t border-card-border">
            {currentUser && (
              <div className="glass-panel p-3 rounded-xl flex items-center justify-between gap-3">
                <div className="flex items-center gap-2 overflow-hidden">
                  <div className="w-8 h-8 rounded-full bg-slate-800 flex items-center justify-center shrink-0">
                    <UserIcon className="w-4 h-4 text-slate-400" />
                  </div>
                  <div className="overflow-hidden">
                    <p className="text-xs font-semibold truncate leading-none mb-1">{currentUser.fullName}</p>
                    <span className="text-[10px] text-accent-cyan font-mono">{currentUser.role}</span>
                  </div>
                </div>

                {/* Logout Button */}
                <button
                  onClick={() => logout()}
                  className="bg-transparent border-0 text-slate-400 cursor-pointer hover:text-red-400 transition-colors p-1"
                  title="Logout"
                >
                  <LogOut className="w-4 h-4" />
                </button>
              </div>
            )}
            <div className="mt-3 flex items-center justify-between text-[10px] text-slate-500 font-mono">
              <span>CD Synchronization</span>
              <span className="flex items-center gap-1 text-accent-emerald">
                <RefreshCw className="w-2.5 h-2.5 animate-spin" />
                Active
              </span>
            </div>
          </div>
        </aside>
        )}

        {/* Content Container */}
        <main className={`flex-1 overflow-y-auto relative ${isAuthRoute ? 'flex items-center justify-center p-0 bg-bg-secondary' : 'p-8'}`}>
          <div className={isAuthRoute ? 'w-full max-w-md' : 'max-w-6xl mx-auto'}>
            {children}
          </div>
        </main>
      </body>
    </html>
  );
}
