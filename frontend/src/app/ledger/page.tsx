'use client';

import React, { useState } from 'react';
import { useStore } from '../../lib/store';
import {
  Database,
  Search,
  Plus,
  Cpu,
  HelpCircle,
  Hash,
  Link as LinkIcon
} from 'lucide-react';
import { api } from '../../lib/api';

export default function LedgerPage() {
  const { ledger } = useStore();
  const [searchQuery, setSearchQuery] = useState('');
  const [logPayload, setLogPayload] = useState('');
  const [writeLoading, setWriteLoading] = useState(false);

  const handleWriteLog = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!logPayload) return;
    setWriteLoading(true);
    try {
      const newBlock = await api.writeLedgerLog(logPayload);
      useStore.setState((state) => ({ ledger: [...state.ledger, newBlock] }));
      setLogPayload('');
    } catch (e) {
      console.error(e);
    } finally {
      setWriteLoading(false);
    }
  };

  const filteredBlocks = ledger.filter(block =>
    block.payload.toLowerCase().includes(searchQuery.toLowerCase()) ||
    block.hash.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div>
      {/* Header */}
      <div className="mb-8 flex justify-between items-center gap-4 flex-wrap">
        <div>
          <h2 className="text-3xl font-extrabold tracking-tight mb-2">Cryptographic Ledger</h2>
          <p className="text-slate-400 text-sm">Verify the immutable transaction hashes sealed in the blockchain-style log chain.</p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left Side: Audit Log Writer */}
        <div className="lg:col-span-1 flex flex-col gap-6">
          <div className="glass-panel p-6 rounded-2xl">
            <h3 className="text-md font-bold mb-4 flex items-center gap-2">
              <Plus className="w-5 h-5 text-accent-cyan" />
              Write Audit Log
            </h3>
            <form onSubmit={handleWriteLog} className="flex flex-col gap-4">
              <div>
                <label className="text-[10px] text-slate-500 font-mono block mb-1">LOG PAYLOAD / MESSAGE</label>
                <textarea
                  value={logPayload}
                  onChange={e => setLogPayload(e.target.value)}
                  className="w-full bg-slate-900 border border-card-border rounded-xl px-4 py-2 text-xs outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan focus-visible:border-accent-cyan min-h-[100px]"
                  placeholder="e.g. Escrow verification released for transaction tx-1"
                  required
                />
              </div>

              <button
                type="submit"
                disabled={writeLoading}
                className="w-full bg-gradient-to-r from-accent-cyan to-accent-blue text-bg-primary font-bold text-xs py-3 rounded-xl hover:opacity-90 active:scale-95 transition-colors transition-transform transition-opacity flex items-center justify-center gap-2 cursor-pointer"
              >
                {writeLoading ? 'Writing Block…' : 'Publish Audit Block'}
              </button>
            </form>
          </div>

          <div className="glass-panel p-6 rounded-2xl text-xs text-slate-400 leading-normal font-mono flex flex-col gap-3">
            <div className="flex items-center gap-2 text-accent-cyan font-bold">
              <Cpu className="w-5 h-5 shrink-0" />
              BLOCKCHAIN MECHANICS
            </div>
            <p>
              Each log block carries a unique sequential Index, UTC Timestamp, and String Payload.
            </p>
            <p>
              The block Hash is calculated by hashing the index, timestamp, payload, and the previous block's SHA256 Hash.
            </p>
            <p className="border-t border-card-border pt-2 text-[10px] text-slate-500">
              Any attempt to modify a block's payload invalidates its hash, breaking the entire downstream log validation chain.
            </p>
          </div>
        </div>

        {/* Right Side: Ledger Blocks Timeline Search */}
        <div className="lg:col-span-2 glass-panel p-6 rounded-2xl">
          {/* Search bar */}
          <div className="flex items-center gap-3 bg-slate-900 border border-card-border rounded-xl px-4 py-3 mb-6">
            <Search className="w-5 h-5 text-slate-500 shrink-0" />
            <input
              type="text"
              placeholder="Search logs by payload message or hash…"
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
              className="bg-transparent border-0 text-xs w-full text-foreground outline-none focus-visible:ring-1 focus-visible:ring-accent-cyan placeholder-slate-500"
            />
          </div>

          {/* Block Timeline */}
          <div className="flex flex-col gap-6 max-h-[500px] overflow-y-auto pr-2">
            {filteredBlocks.length > 0 ? (
              filteredBlocks.slice().reverse().map((block, idx) => {
                const isGenesis = block.index === 0;
                return (
                  <div key={block.index} className="flex gap-4 relative">
                    {/* Visual Line */}
                    <div className="flex flex-col items-center shrink-0">
                      <div className="w-8 h-8 rounded-lg bg-slate-800 border border-card-border flex items-center justify-center font-mono text-xs font-bold text-slate-400">
                        #{block.index}
                      </div>
                      {idx !== filteredBlocks.length - 1 && (
                        <div className="w-[1px] h-full bg-slate-800 my-2" />
                      )}
                    </div>

                    {/* Block Info */}
                    <div className="flex-1 p-4 rounded-xl bg-slate-900/40 border border-card-border hover:border-accent-cyan/20 transition-colors transition-transform transition-opacity duration-300">
                      <div className="flex justify-between items-start gap-4 mb-2 flex-wrap">
                        <p className="text-xs font-bold text-slate-200 font-sans">{block.payload}</p>
                        <span className="text-[10px] text-slate-500 font-mono">
                          {new Date(block.timestamp).toLocaleString()}
                        </span>
                      </div>

                      <div className="flex flex-col gap-1 border-t border-card-border/50 pt-2 font-mono text-[9px] text-slate-500">
                        <div className="flex items-center gap-1.5">
                          <Hash className="w-3 h-3 text-accent-cyan shrink-0" />
                          <span className="text-accent-cyan">BLOCK HASH:</span>
                          <span className="text-slate-400 truncate">{block.hash}</span>
                        </div>
                        {!isGenesis && (
                          <div className="flex items-center gap-1.5">
                            <LinkIcon className="w-3 h-3 text-slate-600 shrink-0" />
                            <span>PREV HASH:</span>
                            <span className="truncate">{block.previousHash}</span>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })
            ) : (
              <div className="flex flex-col items-center justify-center p-12 text-slate-500 gap-2 font-mono text-xs">
                <HelpCircle className="w-8 h-8 text-slate-600" />
                No matching cryptographic blocks found.
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
