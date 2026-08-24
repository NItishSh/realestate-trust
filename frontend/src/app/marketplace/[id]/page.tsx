"use client";

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useStore } from '../../../lib/store';
import { Property, api } from '../../../lib/api';
import { ShieldCheck, ShieldAlert, CheckCircle2, Loader2, X, Home, Maximize, Bed, Bath, Calendar, MapPin, Edit3 } from 'lucide-react';


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

  // Title Insurance states
  const [showInsuranceModal, setShowInsuranceModal] = useState(false);
  const [modalStep, setModalStep] = useState(0); // 0: Idle, 1: Scanning, 2: Liens, 3: Underwriting, 4: Committing, 5: Success
  const [policyNum, setPolicyNum] = useState('');
  const [insurer] = useState('SafeTitle National Insurance Corp');

  // Edit specs states
  const [showEditModal, setShowEditModal] = useState(false);
  const [isEditingSpecs, setIsEditingSpecs] = useState(false);
  const [editSqFt, setEditSqFt] = useState('1800');
  const [editBedrooms, setEditBedrooms] = useState('3');
  const [editBathrooms, setEditBathrooms] = useState('2');
  const [editYearBuilt, setEditYearBuilt] = useState('2018');
  const [editPropertyType, setEditPropertyType] = useState('Residential');

  const handleEditSpecs = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsEditingSpecs(true);
    try {
      const updatedProp = await api.updatePropertyDetails(id as string, {
        sqft: parseInt(editSqFt),
        bedrooms: parseInt(editBedrooms),
        bathrooms: parseInt(editBathrooms),
        yearBuilt: parseInt(editYearBuilt),
        propertyType: editPropertyType
      });
      setProperty(updatedProp);
      setShowEditModal(false);
    } catch (err: any) {
      alert("Failed to update specifications: " + err.message);
    } finally {
      setIsEditingSpecs(false);
    }
  };


  useEffect(() => {
    api.getProperty(id as string)
      .then((data) => {
        setProperty(data);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message);
        setLoading(false);
      });
  }, [id]);

  const startInsuranceWorkflow = () => {
    setModalStep(1);
    const policy = 'TI-ST-' + Math.floor(100000 + Math.random() * 900000);
    setPolicyNum(policy);

    // Step 1: Scanning (1.5s)
    setTimeout(() => {
      setModalStep(2);
      // Step 2: Liens (1.5s)
      setTimeout(() => {
        setModalStep(3);
        // Step 3: Underwriting (1.5s)
        setTimeout(() => {
          setModalStep(4);
          // Step 4: Committing (1.5s)
          setTimeout(async () => {
            try {
              const updatedProp = await api.verifyTitleInsurance(id as string, {
                company: insurer,
                policy: policy,
              });
              setProperty(updatedProp);
              setModalStep(5);
            } catch (err) {
              console.error(err);
              alert("Failed to commit title insurance to ledger.");
              setModalStep(0);
              setShowInsuranceModal(false);
            }
          }, 1500);
        }, 1500);
      }, 1500);
    }, 1500);
  };

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

  if (loading) return <div className="p-8 text-center text-slate-600">Loading property details…</div>;
  if (error || !property) return <div className="p-8 text-center text-red-500">Error: {error || 'Property not found'}</div>;


  return (
    <div className="max-w-4xl mx-auto p-8">
      <div className="bg-white rounded-2xl shadow-2xl overflow-hidden border border-slate-200">
        <img
          src={property.thumbnail}
          alt={property.address}
          className="w-full h-96 object-cover"
        />
        <div className="p-8">
          <h1 className="text-3xl font-extrabold mb-4 text-slate-900">{property.address}</h1>
          <p className="text-lg text-slate-700 mb-6">{property.description}</p>

          <div className="flex justify-between items-center mb-8 pb-8 border-b border-slate-200">
            <div>
              <p className="text-xs text-slate-600 uppercase tracking-wider font-mono">Property Value</p>
              <p className="text-3xl font-bold text-success">₹{property.value.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-xs text-slate-600 uppercase tracking-wider font-mono">Owner ID</p>
              <p className="font-mono text-sm text-slate-700">{property.ownerId}</p>
            </div>
          </div>

          <div className="bg-white shadow-sm p-6 rounded-2xl border border-slate-200">
            <h2 className="text-xl font-bold mb-4 text-slate-800">Due Diligence Data Room</h2>

            {documents.length > 0 ? (
              <div>
                <div className="bg-green-50 border border-accent-emerald/20 text-success p-3.5 rounded-xl mb-4 text-sm font-semibold">
                  {unlockMessage}
                </div>
                <ul className="space-y-2">
                  {documents.map((doc, idx) => (
                    <li key={idx} className="flex items-center text-primary">
                      <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
                      <a href="#" className="hover:underline text-sm font-medium">{doc}</a>
                    </li>
                  ))}
                </ul>
              </div>
            ) : (
              <div className="text-center py-8">
                <svg className="w-16 h-16 mx-auto text-slate-600 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path></svg>
                <p className="text-slate-600 mb-6 text-sm max-w-md mx-auto">Legal documents are locked. Deposit earnest money (₹50,000) into escrow to unlock the data room and begin due diligence.</p>
                <button
                  onClick={handleUnlock}
                  disabled={unlocking}
                  className="btn-primary text-white px-6 py-3 rounded-xl font-bold text-sm hover:opacity-90 active:scale-95 transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed border-0"
                >
                  {unlocking ? 'Processing Escrow…' : 'Deposit ₹50,000 & Unlock Docs'}
                </button>
                {unlockMessage && (
                  <p className="mt-4 text-red-400 text-sm">{unlockMessage}</p>
                )}
              </div>
            )}
          </div>

          {/* Specifications & Location Map */}
          <div className="mt-8 grid grid-cols-1 md:grid-cols-2 gap-8">
            {/* Specs Card */}
            <div className="bg-white shadow-sm p-6 rounded-2xl border border-slate-200 flex flex-col justify-between">
              <div>
                <div className="flex items-center justify-between mb-6">
                  <h3 className="text-lg font-bold text-slate-800 flex items-center gap-2">
                    <Home className="w-5 h-5 text-primary" />
                    Specifications
                  </h3>
                  {currentUser?.id === property.ownerId && (
                    <button
                      onClick={() => {
                        setEditSqFt(property.sqft?.toString() || '1800');
                        setEditBedrooms(property.bedrooms?.toString() || '3');
                        setEditBathrooms(property.bathrooms?.toString() || '2');
                        setEditYearBuilt(property.yearBuilt?.toString() || '2018');
                        setEditPropertyType(property.propertyType || 'Residential');
                        setShowEditModal(true);
                      }}
                      className="p-2 rounded-lg bg-slate-100 text-slate-700 hover:text-slate-900 hover:bg-slate-700/60 border-0 cursor-pointer transition-all flex items-center gap-1.5 text-xs font-semibold"
                    >
                      <Edit3 className="w-3.5 h-3.5" />
                      Edit Specs
                    </button>
                  )}
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="p-3 bg-slate-50 border border-slate-200 rounded-xl flex items-center gap-3">
                    <Maximize className="w-5 h-5 text-primary" />
                    <div>
                      <span className="text-[10px] text-slate-500 block uppercase font-mono">Area</span>
                      <span className="text-sm font-bold text-slate-800">{property.sqft || 1800} sq ft</span>
                    </div>
                  </div>
                  <div className="p-3 bg-slate-50 border border-slate-200 rounded-xl flex items-center gap-3">
                    <Bed className="w-5 h-5 text-primary" />
                    <div>
                      <span className="text-[10px] text-slate-500 block uppercase font-mono">Bedrooms</span>
                      <span className="text-sm font-bold text-slate-800">{property.bedrooms || 3} Beds</span>
                    </div>
                  </div>
                  <div className="p-3 bg-slate-50 border border-slate-200 rounded-xl flex items-center gap-3">
                    <Bath className="w-5 h-5 text-primary" />
                    <div>
                      <span className="text-[10px] text-slate-500 block uppercase font-mono">Bathrooms</span>
                      <span className="text-sm font-bold text-slate-800">{property.bathrooms || 2} Baths</span>
                    </div>
                  </div>
                  <div className="p-3 bg-slate-50 border border-slate-200 rounded-xl flex items-center gap-3">
                    <Calendar className="w-5 h-5 text-primary" />
                    <div>
                      <span className="text-[10px] text-slate-500 block uppercase font-mono">Year Built</span>
                      <span className="text-sm font-bold text-slate-800">{property.yearBuilt || 2018}</span>
                    </div>
                  </div>
                </div>
              </div>

              <div className="mt-4 p-3 bg-slate-50 border border-slate-200 rounded-xl flex items-center gap-3">
                <Home className="w-5 h-5 text-primary" />
                <div>
                  <span className="text-[10px] text-slate-500 block uppercase font-mono">Property Type</span>
                  <span className="text-sm font-bold text-slate-800">{property.propertyType || 'Residential'}</span>
                </div>
              </div>
            </div>

            {/* Map Card */}
            <div className="bg-white shadow-sm p-6 rounded-2xl border border-slate-200 flex flex-col">
              <h3 className="text-lg font-bold text-slate-800 flex items-center gap-2 mb-4">
                <MapPin className="w-5 h-5 text-primary" />
                Location
              </h3>
              <div className="w-full h-48 rounded-xl overflow-hidden border border-slate-200">
                <iframe
                  width="100%"
                  height="100%"
                  frameBorder="0"
                  style={{ border: 0 }}
                  src={`https://maps.google.com/maps?q=${encodeURIComponent(property.address)}&t=&z=13&ie=UTF8&iwloc=&output=embed`}
                  allowFullScreen
                />
              </div>
              <p className="text-xs text-slate-600 mt-3 flex items-center gap-1.5">
                <MapPin className="w-3.5 h-3.5 text-primary shrink-0" />
                {property.address}
              </p>
            </div>
          </div>


          {/* Title Insurance Integration */}
          <div className="mt-8 bg-white shadow-sm p-6 rounded-2xl border border-slate-200">
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-4">
              <div>
                <h2 className="text-xl font-bold flex items-center gap-2 text-slate-800">
                  <ShieldCheck className="w-6 h-6 text-primary" />
                  Title Insurance Registry
                </h2>
                <p className="text-xs text-slate-500 mt-1">Decentralized title underwriting & verification</p>
              </div>
              {property.titleInsuranceStatus === 'INSURED' ? (
                <span className="flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-green-50 text-success border border-accent-emerald/20">
                  <CheckCircle2 className="w-4 h-4 fill-current" />
                  Insured & Verified
                </span>
              ) : (
                <span className="flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200">
                  <ShieldAlert className="w-4 h-4" />
                  Uninsured Title
                </span>
              )}
            </div>

            {property.titleInsuranceStatus === 'INSURED' ? (
              <div className="flex flex-col md:flex-row gap-6 items-center justify-between p-4 rounded-xl bg-slate-50 border border-slate-200">
                <div className="space-y-1 text-slate-700">
                  <p className="text-sm font-semibold">Underwritten by {property.titleInsuranceCompany}</p>
                  <p className="text-xs font-mono text-slate-600">Policy #: {property.titleInsurancePolicy}</p>
                  <p className="text-[10px] text-slate-500 font-mono">Verified On-Chain: {property.titleInsuranceVerifiedAt ? new Date(property.titleInsuranceVerifiedAt).toLocaleString() : 'N/A'}</p>
                </div>
                <div className="text-right">
                  <p className="text-xs text-slate-500">Coverage Limit</p>
                  <p className="text-lg font-bold text-success">₹{property.value.toLocaleString('en-IN')}</p>
                </div>
              </div>
            ) : (
              <div className="text-center py-6">
                <p className="text-slate-600 text-sm mb-4 max-w-md mx-auto">No active title insurance policy has been registered for this property on the decentralized ledger. Protect your purchase against defects in ownership history, undisclosed liens, or encumbrances.</p>
                <button
                  onClick={() => { setShowInsuranceModal(true); startInsuranceWorkflow(); }}
                  className="px-5 py-2.5 rounded-xl btn-primary text-white font-bold text-sm border-0 cursor-pointer transition-all hover:opacity-90 active:scale-95 flex items-center justify-center gap-2 mx-auto"
                >
                  <ShieldCheck className="w-4 h-4" />
                  Run Title Search & Buy Insurance
                </button>
              </div>
            )}
          </div>

          {currentUser?.id === property.ownerId && (
            <div className="bg-white shadow-sm p-6 rounded-2xl border border-slate-200 mt-8">
              <h2 className="text-2xl font-bold mb-2 text-slate-800">Tokenize Property</h2>
              <p className="text-slate-600 mb-6 text-sm">Convert this property into fractional shares and list it on the fractional marketplace.</p>

              <form onSubmit={handleTokenize} className="flex flex-col sm:flex-row gap-4 items-end">
                <div>
                  <label className="text-xs text-slate-600 font-bold block mb-1">TOTAL SHARES</label>
                  <input
                    type="number"
                    value={totalTokens}
                    onChange={e => setTotalTokens(e.target.value)}
                    className="w-full bg-white border border-slate-200 text-slate-900 rounded-xl px-4 py-2 text-sm outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus:border-accent-cyan transition-all"
                    required
                  />
                </div>
                <div>
                  <label className="text-xs text-slate-600 font-bold block mb-1">PRICE PER SHARE (₹)</label>
                  <input
                    type="number"
                    value={tokenPrice}
                    onChange={e => setTokenPrice(e.target.value)}
                    className="w-full bg-white border border-slate-200 text-slate-900 rounded-xl px-4 py-2 text-sm outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus:border-accent-cyan transition-all"
                    required
                  />
                </div>
                <button
                  type="submit"
                  disabled={isTokenizing}
                  className="btn-primary text-white px-6 py-2.5 rounded-xl font-bold text-sm hover:opacity-90 active:scale-95 transition-all cursor-pointer border-0"
                >
                  {isTokenizing ? 'Processing…' : 'Create Fractional Pool'}
                </button>
              </form>
            </div>
          )}
        </div>
      </div>

      {/* Title Insurance Modal */}
      {showInsuranceModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-white/60 backdrop-blur-sm transition-opacity duration-300">
          <div
            className="bg-white w-full max-w-md p-6 rounded-2xl border border-slate-200 shadow-2xl relative"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Close Button only if success or idle */}
            {(modalStep === 5 || modalStep === 0) && (
              <button
                onClick={() => { setShowInsuranceModal(false); setModalStep(0); }}
                className="absolute top-4 right-4 p-1.5 rounded-lg text-slate-600 hover:text-slate-900 hover:bg-slate-100/60 transition-colors border-0 bg-transparent cursor-pointer"
              >
                <X className="w-5 h-5" />
              </button>
            )}

            {modalStep === 5 ? (
              <div className="flex flex-col items-center justify-center py-6 text-center">
                <div className="w-16 h-16 rounded-full bg-green-500/20 text-success flex items-center justify-center mb-4">
                  <CheckCircle2 className="w-8 h-8 fill-current" />
                </div>
                <h3 className="text-xl font-bold text-slate-800 mb-2">Title Insured Successfully</h3>
                <p className="text-slate-600 text-xs mb-6">Verification recorded in ledger audit trail.</p>

                {/* Certificate */}
                <div className="w-full bg-slate-50 border border-slate-200 border-dashed rounded-xl p-5 text-left relative overflow-hidden">
                  <div className="absolute top-0 right-0 transform translate-x-4 -translate-y-4 w-24 h-24 bg-green-50 rounded-full flex items-center justify-center">
                    <ShieldCheck className="w-10 h-10 text-success/30" />
                  </div>
                  <h4 className="text-[10px] text-slate-500 uppercase tracking-widest font-bold mb-3">SafeTitle Certificate</h4>
                  <div className="space-y-2 text-slate-700 text-xs">
                    <div>
                      <span className="text-[10px] text-slate-500 block">POLICY NUMBER</span>
                      <span className="font-mono font-bold text-slate-800">{policyNum}</span>
                    </div>
                    <div>
                      <span className="text-[10px] text-slate-500 block">COVERAGE AMOUNT</span>
                      <span className="font-bold text-success">₹{property.value.toLocaleString('en-IN')}</span>
                    </div>
                    <div>
                      <span className="text-[10px] text-slate-500 block">UNDERWRITER</span>
                      <span className="font-bold text-slate-800">{insurer}</span>
                    </div>
                  </div>
                </div>

                <button
                  onClick={() => { setShowInsuranceModal(false); setModalStep(0); }}
                  className="mt-6 w-full py-2.5 rounded-xl btn-primary text-white font-bold text-sm border-0 cursor-pointer hover:opacity-90 active:scale-95"
                >
                  Done
                </button>
              </div>
            ) : (
              <div className="flex flex-col gap-5">
                <div>
                  <h3 className="text-lg font-bold text-slate-800">Underwriting Title Insurance</h3>
                  <p className="text-xs text-slate-600 mt-1">Simulating title verification and smart contract log commit.</p>
                </div>

                {/* Steps */}
                <div className="flex flex-col gap-4">
                  {[
                    { step: 1, label: 'Scanning Chain-of-Custody Ledger', desc: 'Analyzing historical deed registrations...' },
                    { step: 2, label: 'Lien & Encumbrance Verification', desc: 'Checking registry database for outstanding debts...' },
                    { step: 3, label: 'Policy Underwriting Approval', desc: 'Issuing coverage limit of ₹' + property.value.toLocaleString('en-IN') },
                    { step: 4, label: 'Writing Block to Audit Ledger', desc: 'Finalizing smart contract log and publishing...' }
                  ].map((s) => {
                    const isDone = modalStep > s.step;
                    const isActive = modalStep === s.step;
                    const isUpcoming = modalStep < s.step;

                    return (
                      <div key={s.step} className="flex gap-3">
                        <div className="flex flex-col items-center">
                          <div className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold transition-all duration-300 ${
                            isDone ? 'bg-green-500/20 text-success' :
                            isActive ? 'bg-accent-cyan text-white shadow-md shadow-primary/20' :
                            'bg-slate-100 text-slate-500'
                          }`}>
                            {isDone ? <CheckCircle2 className="w-3.5 h-3.5 fill-current text-success" /> : s.step}
                          </div>
                          {s.step < 4 && (
                            <div className={`w-0.5 h-8 my-1 transition-colors duration-500 ${
                              isDone ? 'bg-green-500' : 'bg-slate-100'
                            }`} />
                          )}
                        </div>
                        <div className="flex-1 py-0.5">
                          <p className={`text-xs font-bold transition-colors ${
                            isUpcoming ? 'text-slate-500' : 'text-slate-800'
                          }`}>
                            {s.label}
                          </p>
                          {(isActive || isDone) && (
                            <p className="text-[10px] text-slate-600 mt-0.5 flex items-center gap-1.5 animate-pulse">
                              {isActive && <Loader2 className="w-2.5 h-2.5 animate-spin" />}
                              {s.desc}
                            </p>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Edit Specifications Modal */}
      {showEditModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-white/60 backdrop-blur-sm transition-opacity duration-300">
          <div
            className="bg-white w-full max-w-md p-6 rounded-2xl border border-slate-200 shadow-2xl relative"
            onClick={(e) => e.stopPropagation()}
          >
            <button
              onClick={() => setShowEditModal(false)}
              className="absolute top-4 right-4 p-1.5 rounded-lg text-slate-600 hover:text-slate-900 hover:bg-slate-100/60 transition-colors border-0 bg-transparent cursor-pointer"
            >
              <X className="w-5 h-5" />
            </button>

            <form onSubmit={handleEditSpecs} className="flex flex-col gap-4">
              <div>
                <h3 className="text-lg font-bold text-slate-800">Edit Property Specifications</h3>
                <p className="text-xs text-slate-600 mt-1">Update specifications and registry details for this property.</p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-xs text-slate-600 font-bold block mb-1">AREA (SQ FT)</label>
                  <input
                    type="number"
                    value={editSqFt}
                    onChange={(e) => setEditSqFt(e.target.value)}
                    className="w-full bg-white border border-slate-200 text-slate-900 rounded-xl px-4 py-2 text-sm outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus:border-accent-cyan transition-all"
                    required
                  />
                </div>
                <div>
                  <label className="text-xs text-slate-600 font-bold block mb-1">PROPERTY TYPE</label>
                  <select
                    value={editPropertyType}
                    onChange={(e) => setEditPropertyType(e.target.value)}
                    className="w-full bg-white border border-slate-200 text-slate-900 rounded-xl px-4 py-2 text-sm outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus:border-accent-cyan transition-all"
                    required
                  >
                    <option value="Residential">Residential</option>
                    <option value="Commercial">Commercial</option>
                    <option value="Industrial">Industrial</option>
                    <option value="Land">Land</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-3">
                <div>
                  <label className="text-[10px] text-slate-600 font-bold block mb-1">BEDROOMS</label>
                  <input
                    type="number"
                    value={editBedrooms}
                    onChange={(e) => setEditBedrooms(e.target.value)}
                    className="w-full bg-white border border-slate-200 text-slate-900 rounded-xl px-4 py-2 text-sm outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus:border-accent-cyan transition-all"
                    required
                  />
                </div>
                <div>
                  <label className="text-[10px] text-slate-600 font-bold block mb-1">BATHROOMS</label>
                  <input
                    type="number"
                    value={editBathrooms}
                    onChange={(e) => setEditBathrooms(e.target.value)}
                    className="w-full bg-white border border-slate-200 text-slate-900 rounded-xl px-4 py-2 text-sm outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus:border-accent-cyan transition-all"
                    required
                  />
                </div>
                <div>
                  <label className="text-[10px] text-slate-600 font-bold block mb-1">YEAR BUILT</label>
                  <input
                    type="number"
                    value={editYearBuilt}
                    onChange={(e) => setEditYearBuilt(e.target.value)}
                    className="w-full bg-white border border-slate-200 text-slate-900 rounded-xl px-4 py-2 text-sm outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus:border-accent-cyan transition-all"
                    required
                  />
                </div>
              </div>

              <button
                type="submit"
                disabled={isEditingSpecs}
                className="mt-2 w-full py-2.5 rounded-xl btn-primary text-white font-bold text-sm border-0 cursor-pointer hover:opacity-90 active:scale-95 disabled:opacity-50"
              >
                {isEditingSpecs ? 'Saving Changes...' : 'Save Specifications'}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
