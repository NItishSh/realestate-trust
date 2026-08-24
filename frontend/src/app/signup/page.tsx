'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ShieldCheck, Mail, Lock, User, ArrowRight, Loader2, Building, UserCheck } from 'lucide-react';

export default function SignupPage() {
  const router = useRouter();
  const [formData, setFormData] = useState({
    fullName: '',
    email: '',
    password: '',
    role: 'buyer'
  });
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSignup = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });

      if (!response.ok) {
        throw new Error('Registration failed');
      }

      router.push('/login?registered=true');
    } catch (err) {
      setError('Registration failed. Please try again.');
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex bg-gray-50">
      {/* Left side - Image/Branding */}
      <div className="hidden lg:flex w-1/2 bg-[#001a41] relative flex-col justify-between p-12 overflow-hidden">
        <div className="relative z-10">
          <div className="w-12 h-12 rounded-lg bg-[#0d47a1] flex items-center justify-center mb-6 shadow-md">
            <ShieldCheck className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-4xl font-bold text-white mb-4">TrustEstate</h1>
          <p className="text-xl text-blue-200 font-light">Secure, transparent, and compliant real estate transactions.</p>
        </div>

        {/* Decorative elements */}
        <div className="absolute top-0 right-0 w-[800px] h-[800px] bg-[#0d47a1] rounded-full blur-3xl opacity-20 -translate-y-1/2 translate-x-1/3"></div>
        <div className="absolute bottom-0 left-0 w-[600px] h-[600px] bg-blue-400 rounded-full blur-3xl opacity-10 translate-y-1/3 -translate-x-1/4"></div>

        <div className="relative z-10 text-blue-300/80 text-sm">
          &copy; {new Date().getFullYear()} TrustEstate. All rights reserved. FDIC Insured equivalent escrow protection.
        </div>
      </div>

      {/* Right side - Form */}
      <div className="w-full lg:w-1/2 flex items-center justify-center p-8">
        <div className="max-w-md w-full">
          <div className="text-center mb-8 lg:hidden">
            <div className="w-16 h-16 rounded-xl bg-[#0d47a1] flex items-center justify-center mx-auto mb-6 shadow-md">
              <ShieldCheck className="w-10 h-10 text-white" />
            </div>
            <h1 className="text-3xl font-bold tracking-tight text-gray-900 mb-2">TrustEstate</h1>
            <p className="text-gray-500">Create your account</p>
          </div>

          <div className="lg:text-left mb-8 hidden lg:block">
            <h2 className="text-3xl font-bold tracking-tight text-gray-900 mb-2">Create an account</h2>
            <p className="text-gray-500">Join the secure real estate network</p>
          </div>

          <div className="bg-white p-8 rounded-2xl shadow-sm border border-gray-200">
            <form onSubmit={handleSignup} className="space-y-5">
              {error && (
                <div className="p-4 rounded-lg bg-red-50 border border-red-200 text-red-600 text-sm">
                  {error}
                </div>
              )}

              {/* Role Selection */}
              <div className="space-y-3 mb-6">
                <label className="text-sm font-semibold text-gray-700">I am a...</label>
                <div className="grid grid-cols-3 gap-3">
                  {[
                    { id: 'buyer', label: 'Buyer', icon: User },
                    { id: 'seller', label: 'Seller', icon: Building },
                    { id: 'admin', label: 'Manager', icon: UserCheck }
                  ].map((role) => (
                    <div
                      key={role.id}
                      onClick={() => setFormData({...formData, role: role.id})}
                      className={`cursor-pointer rounded-xl border p-3 flex flex-col items-center justify-center gap-2 transition-all ${
                        formData.role === role.id
                          ? 'border-[#0d47a1] bg-[#d8e2ff] text-[#001a41]'
                          : 'border-gray-200 hover:border-gray-300 text-gray-600'
                      }`}
                    >
                      <role.icon className="w-5 h-5" />
                      <span className="text-xs font-semibold">{role.label}</span>
                    </div>
                  ))}
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-semibold text-gray-700">Full Name</label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <User className="h-5 w-5 text-gray-400" />
                  </div>
                  <input
                    type="text"
                    required
                    className="w-full bg-white border border-gray-300 rounded-lg py-2.5 pl-12 pr-4 outline-none focus:ring-2 focus:ring-[#0d47a1] focus:border-[#0d47a1] transition-all text-gray-900"
                    placeholder="John Doe"
                    value={formData.fullName}
                    onChange={(e) => setFormData({...formData, fullName: e.target.value})}
                  />
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-semibold text-gray-700">Email Address</label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <Mail className="h-5 w-5 text-gray-400" />
                  </div>
                  <input
                    type="email"
                    required
                    className="w-full bg-white border border-gray-300 rounded-lg py-2.5 pl-12 pr-4 outline-none focus:ring-2 focus:ring-[#0d47a1] focus:border-[#0d47a1] transition-all text-gray-900"
                    placeholder="john@example.com"
                    value={formData.email}
                    onChange={(e) => setFormData({...formData, email: e.target.value})}
                  />
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-semibold text-gray-700">Password</label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <Lock className="h-5 w-5 text-gray-400" />
                  </div>
                  <input
                    type="password"
                    required
                    className="w-full bg-white border border-gray-300 rounded-lg py-2.5 pl-12 pr-4 outline-none focus:ring-2 focus:ring-[#0d47a1] focus:border-[#0d47a1] transition-all text-gray-900"
                    placeholder="••••••••"
                    value={formData.password}
                    onChange={(e) => setFormData({...formData, password: e.target.value})}
                  />
                </div>
              </div>

              <button
                type="submit"
                disabled={isLoading || !formData.email || !formData.password || !formData.fullName}
                className="w-full bg-[#0d47a1] hover:bg-[#1565c0] text-white font-semibold py-3 px-4 rounded-lg flex items-center justify-center gap-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed mt-4"
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

            <div className="mt-6 text-center text-sm text-gray-500">
              Already have an account?{' '}
              <Link href="/login" className="text-[#0d47a1] hover:underline font-semibold">
                Sign in
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
