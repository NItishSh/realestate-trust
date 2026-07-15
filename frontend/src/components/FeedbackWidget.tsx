'use client';

import React, { useState } from 'react';
import { MessageSquare, Send, Star, X } from 'lucide-react';
import { api } from '../lib/api';

export default function FeedbackWidget() {
  const [isOpen, setIsOpen] = useState(false);
  const [message, setMessage] = useState('');
  const [category, setCategory] = useState('general');
  const [rating, setRating] = useState(5);
  const [submitting, setSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!message.trim()) {
      setError('Please provide feedback message.');
      return;
    }

    setSubmitting(true);
    setError('');

    try {
      await api.submitFeedback({ message, category, rating });
      setSubmitted(true);
      setMessage('');
      setCategory('general');
      setRating(5);
      setTimeout(() => {
        setSubmitted(false);
        setIsOpen(false);
      }, 2000);
    } catch (err) {
      console.error(err);
      setError('Failed to submit feedback. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      {/* Floating Button */}
      <button
        onClick={() => setIsOpen(true)}
        className="fixed bottom-6 right-6 z-50 flex items-center gap-2 px-4 py-3 rounded-full bg-accent-cyan hover:bg-cyan-500 text-bg-primary font-semibold shadow-lg shadow-accent-cyan/30 hover:shadow-accent-cyan/50 hover:scale-105 active:scale-95 transition-all duration-300 border-0 cursor-pointer"
        aria-label="Provide Feedback"
      >
        <MessageSquare className="w-5 h-5" />
        <span className="text-sm font-semibold">Feedback</span>
      </button>

      {/* Backdrop & Modal */}
      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/60 backdrop-blur-sm transition-opacity duration-300">
          <div
            className="glass-panel w-full max-w-md p-6 rounded-2xl border border-card-border bg-bg-secondary/90 shadow-2xl relative animate-in fade-in zoom-in-95 duration-200"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Close Button */}
            <button
              onClick={() => setIsOpen(false)}
              className="absolute top-4 right-4 p-1.5 rounded-lg text-slate-400 hover:text-foreground hover:bg-slate-800/60 transition-colors border-0 bg-transparent cursor-pointer"
            >
              <X className="w-5 h-5" />
            </button>

            {submitted ? (
              <div className="flex flex-col items-center justify-center py-8 text-center">
                <div className="w-16 h-16 rounded-full bg-accent-emerald/20 text-accent-emerald flex items-center justify-center mb-4">
                  <Star className="w-8 h-8 fill-current" />
                </div>
                <h3 className="text-xl font-bold mb-2">Thank you!</h3>
                <p className="text-slate-400 text-sm">Your feedback helps us build a better platform.</p>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="flex flex-col gap-4">
                <div>
                  <h3 className="text-lg font-bold">Help us improve</h3>
                  <p className="text-xs text-slate-400 mt-1">We value your input on new features, bugs, or improvements.</p>
                </div>

                {error && (
                  <div className="p-3 text-xs bg-red-500/10 border border-red-500/20 text-red-400 rounded-lg">
                    {error}
                  </div>
                )}

                {/* Rating */}
                <div className="flex flex-col gap-2">
                  <label className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Rating</label>
                  <div className="flex gap-2">
                    {[1, 2, 3, 4, 5].map((star) => (
                      <button
                        key={star}
                        type="button"
                        onClick={() => setRating(star)}
                        className="bg-transparent border-0 p-1 cursor-pointer transition-transform hover:scale-110"
                      >
                        <Star
                          className={`w-7 h-7 transition-colors ${
                            star <= rating
                              ? 'text-yellow-400 fill-yellow-400'
                              : 'text-slate-600 hover:text-slate-400'
                          }`}
                        />
                      </button>
                    ))}
                  </div>
                </div>

                {/* Category */}
                <div className="flex flex-col gap-2">
                  <label className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Category</label>
                  <div className="grid grid-cols-3 gap-2">
                    {[
                      { id: 'general', label: 'General' },
                      { id: 'bug', label: 'Bug' },
                      { id: 'feature_request', label: 'Feature Request' },
                    ].map((cat) => (
                      <button
                        key={cat.id}
                        type="button"
                        onClick={() => setCategory(cat.id)}
                        className={`py-2 px-3 text-xs rounded-xl border font-medium cursor-pointer transition-colors ${
                          category === cat.id
                            ? 'bg-accent-cyan/15 border-accent-cyan text-accent-cyan'
                            : 'bg-slate-800/40 border-card-border text-slate-400 hover:bg-slate-800/80 hover:text-foreground'
                        }`}
                      >
                        {cat.label}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Message */}
                <div className="flex flex-col gap-2">
                  <label htmlFor="feedback-message" className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
                    Feedback Details
                  </label>
                  <textarea
                    id="feedback-message"
                    rows={4}
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    placeholder="Describe your feedback or request..."
                    className="w-full bg-slate-900/60 border border-card-border focus:border-accent-cyan focus:ring-1 focus:ring-accent-cyan rounded-xl p-3 text-sm text-foreground placeholder:text-slate-500 focus:outline-none resize-none transition-all"
                  />
                </div>

                {/* Submit */}
                <button
                  type="submit"
                  disabled={submitting}
                  className="flex items-center justify-center gap-2 mt-2 w-full py-3 rounded-xl bg-accent-cyan hover:bg-cyan-500 text-bg-primary font-bold shadow-lg shadow-accent-cyan/10 hover:shadow-accent-cyan/20 cursor-pointer disabled:opacity-50 transition-all border-0"
                >
                  {submitting ? 'Submitting...' : 'Submit Feedback'}
                  {!submitting && <Send className="w-4 h-4" />}
                </button>
              </form>
            )}
          </div>
        </div>
      )}
    </>
  );
}
