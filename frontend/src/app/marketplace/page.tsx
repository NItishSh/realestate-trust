"use client";

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api } from '../../lib/api';

interface Property {
  id: string;
  address: string;
  description: string;
  value: number;
  thumbnail: string;
  ownerId: string;
}

export default function MarketplacePage() {
  const [properties, setProperties] = useState<Property[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    api.getProperties()
      .then((data) => {
        setProperties(data || []);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message);
        setLoading(false);
      });
  }, []);

  if (loading) return <div className="p-8 text-center">Loading marketplace…</div>;
  if (error) return <div className="p-8 text-center text-red-500">Error: {error}</div>;

  return (
    <div className="max-w-6xl mx-auto p-8">
      <h1 className="text-3xl font-bold mb-8">Property Marketplace</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {properties.map((property) => (
          <div key={property.id} className="border rounded-lg overflow-hidden shadow-sm hover:shadow-md transition-shadow">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={property.thumbnail}
              alt={property.address}
              className="w-full h-48 object-cover"
            />
            <div className="p-4">
              <h2 className="text-xl font-semibold mb-2">{property.address}</h2>
              <p className="text-gray-600 mb-4 line-clamp-2">{property.description}</p>
              <div className="flex justify-between items-center">
                <span className="text-lg font-bold">₹{property.value.toLocaleString()}</span>
                <Link
                  href={`/marketplace/${property.id}`}
                  className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition-colors"
                >
                  View Details
                </Link>
              </div>
            </div>
          </div>
        ))}
      </div>

      {properties.length === 0 && (
        <div className="text-center text-gray-500 mt-12">
          No properties currently available on the marketplace.
        </div>
      )}
    </div>
  );
}
