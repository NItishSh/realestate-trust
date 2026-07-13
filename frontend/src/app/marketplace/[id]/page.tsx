"use client";

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useStore } from '../../../lib/store';

interface Property {
  id: string;
  address: string;
  description: string;
  value: number;
  thumbnail: string;
  ownerId: string;
}

export default function PropertyDetailsPage() {
  const { id } = useParams();
  const [property, setProperty] = useState<Property | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const [documents, setDocuments] = useState<string[]>([]);
  const [unlocking, setUnlocking] = useState(false);
  const [unlockMessage, setUnlockMessage] = useState('');

  const { currentUser, createPool } = useStore();
  const router = useRouter();

  // Tokenization states
  const [isTokenizing, setIsTokenizing] = useState(false);
  const [totalTokens, setTotalTokens] = useState('1000');
  const [tokenPrice, setTokenPrice] = useState('10000');

  useEffect(() => {
    fetch(`http://localhost:8085/api/v1/properties/${id}`)
      .then((res) => {
        if (!res.ok) throw new Error('Failed to fetch property details');
        return res.json();
      })
      .then((data) => {
        setProperty(data);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message);
        setLoading(false);
      });
  }, [id]);

  const handleUnlock = async () => {
    setUnlocking(true);
    setUnlockMessage('');
    try {
      const res = await fetch(`http://localhost:8085/api/v1/properties/${id}/documents/unlock`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ buyerId: 'usr-aryan.dev@realestate.in' }) // Hardcoded for demo
      });

      const data = await res.json();
      if (!res.ok) throw new Error(data || 'Failed to unlock documents');

      setUnlockMessage(data.message);
      setDocuments(data.documents || []);
    } catch (err: any) {
      setUnlockMessage(err.message);
    } finally {
      setUnlocking(false);
    }
  };

  const handleTokenize = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsTokenizing(true);
    try {
      await createPool({
        propertyId: id as string,
        totalTokens: parseInt(totalTokens),
        tokenPrice: parseFloat(tokenPrice)
      });
      router.push('/portfolio');
    } catch (err: any) {
      alert("Failed to tokenize property: " + err.message);
    } finally {
      setIsTokenizing(false);
    }
  };

  if (loading) return <div className="p-8 text-center">Loading property…</div>;
  if (error || !property) return <div className="p-8 text-center text-red-500">Error: {error || 'Property not found'}</div>;

  return (
    <div className="max-w-4xl mx-auto p-8">
      <div className="bg-white rounded-lg shadow-lg overflow-hidden">
        <img
          src={property.thumbnail}
          alt={property.address}
          className="w-full h-96 object-cover"
        />
        <div className="p-8">
          <h1 className="text-3xl font-bold mb-4">{property.address}</h1>
          <p className="text-xl text-gray-700 mb-6">{property.description}</p>

          <div className="flex justify-between items-center mb-8 pb-8 border-b">
            <div>
              <p className="text-sm text-gray-500 uppercase tracking-wide">Property Value</p>
              <p className="text-3xl font-bold text-green-600">₹{property.value.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500 uppercase tracking-wide">Owner ID</p>
              <p className="font-mono">{property.ownerId}</p>
            </div>
          </div>

          <div className="bg-gray-50 p-6 rounded-lg border border-gray-200">
            <h2 className="text-2xl font-semibold mb-4">Due Diligence Data Room</h2>

            {documents.length > 0 ? (
              <div>
                <div className="bg-green-100 text-green-800 p-3 rounded mb-4">
                  {unlockMessage}
                </div>
                <ul className="space-y-2">
                  {documents.map((doc, idx) => (
                    <li key={idx} className="flex items-center text-blue-600">
                      <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
                      <a href="#" className="hover:underline">{doc}</a>
                    </li>
                  ))}
                </ul>
              </div>
            ) : (
              <div className="text-center py-8">
                <svg className="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path></svg>
                <p className="text-gray-600 mb-6">Legal documents are locked. Deposit earnest money (₹50,000) into escrow to unlock the data room and begin due diligence.</p>
                <button
                  onClick={handleUnlock}
                  disabled={unlocking}
                  className="bg-blue-600 text-white px-6 py-3 rounded-lg font-semibold hover:bg-blue-700 transition-colors disabled:bg-blue-400 disabled:cursor-not-allowed"
                >
                  {unlocking ? 'Processing Escrow…' : 'Deposit ₹50,000 & Unlock Docs'}
                </button>
                {unlockMessage && (
                  <p className="mt-4 text-red-500 text-sm">{unlockMessage}</p>
                )}
              </div>
            )}
          </div>

          {currentUser?.id === property.ownerId && (
            <div className="bg-blue-50 p-6 rounded-lg border border-blue-200 mt-8">
              <h2 className="text-2xl font-semibold mb-2 text-blue-900">Tokenize Property</h2>
              <p className="text-blue-700 mb-6 text-sm">Convert this property into fractional shares and list it on the fractional marketplace.</p>

              <form onSubmit={handleTokenize} className="flex flex-col sm:flex-row gap-4 items-end">
                <div>
                  <label className="text-xs text-blue-800 font-semibold block mb-1">TOTAL SHARES</label>
                  <input
                    type="number"
                    value={totalTokens}
                    onChange={e => setTotalTokens(e.target.value)}
                    className="w-full border border-blue-300 rounded-lg px-4 py-2 outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus:border-blue-500"
                    required
                  />
                </div>
                <div>
                  <label className="text-xs text-blue-800 font-semibold block mb-1">PRICE PER SHARE (₹)</label>
                  <input
                    type="number"
                    value={tokenPrice}
                    onChange={e => setTokenPrice(e.target.value)}
                    className="w-full border border-blue-300 rounded-lg px-4 py-2 outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus:border-blue-500"
                    required
                  />
                </div>
                <button
                  type="submit"
                  disabled={isTokenizing}
                  className="bg-blue-600 text-white px-6 py-2.5 rounded-lg font-semibold hover:bg-blue-700 transition-colors disabled:bg-blue-400"
                >
                  {isTokenizing ? 'Processing…' : 'Create Fractional Pool'}
                </button>
              </form>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
