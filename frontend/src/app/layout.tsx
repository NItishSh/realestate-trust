'use client';

import React, { useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useStore } from '../lib/store';
import {
  LayoutDashboard,
  Home,
  Briefcase,
  ShieldCheck,
  RefreshCw,
  LogOut,
  Building
} from 'lucide-react';
import './globals.css';
import FeedbackWidget from '../components/FeedbackWidget';

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const { currentUser, fetchInitialData, logout } = useStore();
  const isAuthRoute = pathname === '/login' || pathname === '/signup';

  useEffect(() => {
    fetchInitialData();
  }, [fetchInitialData]);

  const navItems = [
    { name: 'Buyer Dashboard', path: '/buyer', icon: LayoutDashboard },
    { name: 'Seller Dashboard', path: '/seller', icon: Home },
    { name: 'Escrow Manager', path: '/admin', icon: Briefcase },
    { name: 'Properties', path: '/marketplace', icon: Building },
  ];

  return (
    <html lang="en">
      <body className="flex h-screen bg-[#f8f9fa] text-[#1f1f1f] font-sans">
        {!isAuthRoute && (
          <aside className="w-64 bg-white border-r border-[#e1e2e5] flex flex-col justify-between shrink-0 shadow-sm">
            <div>
              {/* Header */}
              <div className="p-6 border-b border-[#e1e2e5] flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-[#0d47a1] flex items-center justify-center shadow-sm">
                  <ShieldCheck className="w-5 h-5 text-white" />
                </div>
                <div>
                  <h1 className="font-bold text-sm tracking-wider uppercase text-[#0d47a1]">TrustEstate</h1>
                  <span className="text-[10px] text-gray-500 font-bold tracking-widest block uppercase">Secure Portal</span>
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
                      className={`flex items-center gap-3 px-4 py-3 rounded-lg transition-all duration-200 cursor-pointer ${
                        isActive
                          ? 'bg-[#d8e2ff] text-[#001a41] font-semibold'
                          : 'text-[#444746] hover:bg-[#f1f3f4] hover:text-[#1f1f1f]'
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
            <div className="p-4 border-t border-[#e1e2e5]">
              {currentUser && (
                <div className="bg-[#f8f9fa] p-3 rounded-lg flex items-center justify-between gap-3 border border-[#e1e2e5]">
                  <div className="flex items-center gap-2 overflow-hidden">
                    <div className="w-8 h-8 rounded-full bg-[#0d47a1] flex items-center justify-center shrink-0">
                      <span className="text-white text-xs font-bold">
                        {currentUser.fullName.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <div className="overflow-hidden">
                      <p className="text-xs font-semibold truncate leading-none mb-1 text-[#1f1f1f]">{currentUser.fullName}</p>
                      <span className="text-[10px] text-[#444746] font-medium">{currentUser.role}</span>
                    </div>
                  </div>

                  {/* Logout Button */}
                  <button
                    onClick={() => logout()}
                    className="bg-transparent border-0 text-[#444746] cursor-pointer hover:text-[#b3261e] transition-colors p-1"
                    title="Logout"
                  >
                    <LogOut className="w-4 h-4" />
                  </button>
                </div>
              )}
              <div className="mt-3 flex items-center justify-between text-[10px] text-gray-500">
                <span>System Status</span>
                <span className="flex items-center gap-1 text-[#146c2e] font-semibold">
                  <RefreshCw className="w-2.5 h-2.5 animate-spin" />
                  Secure
                </span>
              </div>
            </div>
          </aside>
        )}

        {/* Content Container */}
        <main className={`flex-1 overflow-y-auto relative ${isAuthRoute ? 'flex items-center justify-center p-0 bg-[#f1f3f4]' : 'bg-[#f1f3f4]'}`}>
          <div className={isAuthRoute ? 'w-full h-full' : 'w-full'}>
            {children}
          </div>
        </main>
        {!isAuthRoute && <FeedbackWidget />}
      </body>
    </html>
  );
}
