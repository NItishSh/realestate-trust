'use client';

import React, { useState } from 'react';
import { MessageSquare, Send, Star, X, Paperclip } from 'lucide-react';
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
        className="fixed bottom-6 right-6 z-50 flex items-center gap-2 px-4 py-3 rounded-full bg-[var(--color-primary)] text-[var(--color-on-primary)] shadow-lg hover:shadow-xl hover:scale-105 active:scale-95 transition-all duration-300 border-0 cursor-pointer"
        aria-label="Provide Feedback"
      >
        <MessageSquare className="w-5 h-5" />
        <span className="text-sm font-semibold">Feedback</span>
      </button>

      {/* Backdrop & Modal */}
      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm transition-opacity duration-300">
          <div
            className="w-full max-w-md bg-[var(--color-surface)] rounded-3xl shadow-[0_20px_25px_-5px_rgba(0,0,0,0.1),0_10px_10px_-5px_rgba(0,0,0,0.04)] border border-[var(--color-outline-variant)]/30 overflow-hidden flex flex-col relative animate-in fade-in zoom-in-95 duration-200"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Header */}
            <div className="px-6 py-5 border-b border-[var(--color-outline-variant)]/50 flex justify-between items-center bg-[var(--color-surface)]">
              <div>
                <h2 className="font-bold text-xl text-[var(--color-on-surface)]">Share Your Feedback</h2>
                <p className="text-sm text-[var(--color-on-surface-variant)] mt-1">Help us improve TrustEstate</p>
              </div>
              <button
                onClick={() => setIsOpen(false)}
                className="p-2 rounded-full hover:bg-[var(--color-surface-variant)] text-[var(--color-on-surface-variant)] transition-colors border-0 bg-transparent cursor-pointer focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)]/50"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {submitted ? (
              <div className="flex flex-col items-center justify-center py-10 px-6 text-center">
                <div className="w-16 h-16 rounded-full bg-[var(--color-success-container)] text-[var(--color-on-success-container)] flex items-center justify-center mb-4">
                  <Star className="w-8 h-8 fill-current" />
                </div>
                <h3 className="text-xl font-bold mb-2 text-[var(--color-on-surface)]">Thank you!</h3>
                <p className="text-[var(--color-on-surface-variant)] text-sm">Your feedback helps us build a better platform.</p>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="flex flex-col gap-6">
                <div className="p-6 pb-2 flex flex-col gap-6">
                  {error && (
                    <div className="p-3 text-sm bg-[var(--color-error-container)] border border-[var(--color-error)] text-[var(--color-on-error-container)] rounded-lg">
                      {error}
                    </div>
                  )}

                  {/* Rating Section */}
                  <div className="flex flex-col items-center gap-2">
                    <label className="font-semibold text-sm text-[var(--color-on-surface)] self-start">How was your experience?</label>
                    <div className="flex gap-2 justify-center py-2">
                      {[1, 2, 3, 4, 5].map((star) => (
                        <button
                          key={star}
                          type="button"
                          onClick={() => setRating(star)}
                          className="bg-transparent border-0 p-1 cursor-pointer transition-transform hover:scale-110 focus:outline-none"
                        >
                          <Star
                            className={`w-9 h-9 transition-colors ${
                              star <= rating
                                ? 'text-amber-500 fill-amber-500'
                                : 'text-[var(--color-outline-variant)]'
                            }`}
                          />
                        </button>
                      ))}
                    </div>
                  </div>

                  {/* Category Selector (Segmented Control Style) */}
                  <div className="flex flex-col gap-2">
                    <label className="font-semibold text-sm text-[var(--color-on-surface)]">Feedback Type</label>
                    <div className="flex p-1 bg-[var(--color-surface-variant)]/50 rounded-lg">
                      {[
                        { id: 'general', label: 'General' },
                        { id: 'bug', label: 'Bug' },
                        { id: 'feature_request', label: 'Feature' },
                      ].map((cat) => (
                        <button
                          key={cat.id}
                          type="button"
                          onClick={() => setCategory(cat.id)}
                          className={`flex-1 py-2 px-3 text-sm font-medium rounded-md border cursor-pointer transition-all ${
                            category === cat.id
                              ? 'bg-[var(--color-surface)] shadow-sm text-[var(--color-primary)] border-[var(--color-outline-variant)]/20'
                              : 'border-transparent text-[var(--color-on-surface-variant)] hover:text-[var(--color-on-surface)] hover:bg-[var(--color-surface)]/50 bg-transparent'
                          }`}
                        >
                          {cat.label}
                        </button>
                      ))}
                    </div>
                  </div>

                  {/* Text Area */}
                  <div className="flex flex-col gap-2">
                    <label htmlFor="feedback-details" className="font-semibold text-sm text-[var(--color-on-surface)]">
                      Feedback Details
                    </label>
                    <textarea
                      id="feedback-details"
                      rows={4}
                      value={message}
                      onChange={(e) => setMessage(e.target.value)}
                      placeholder="Tell us more about your experience..."
                      className="w-full rounded-xl border border-[var(--color-outline-variant)] bg-[var(--color-surface)] px-4 py-3 text-sm text-[var(--color-on-surface)] placeholder:text-[var(--color-on-surface-variant)] focus:border-[var(--color-primary)] focus:ring-1 focus:ring-[var(--color-primary)] resize-none transition-shadow outline-none"
                    />
                  </div>

                  {/* Optional Attachment (Bonus UI detail) */}
                  <div className="flex items-center gap-3">
                    <button type="button" className="flex items-center gap-2 text-sm font-medium text-[var(--color-primary)] hover:opacity-80 transition-colors border-0 bg-transparent cursor-pointer">
                      <Paperclip className="w-4 h-4" />
                      Add screenshot (optional)
                    </button>
                  </div>
                </div>

                {/* Footer / Action */}
                <div className="p-6 pt-2 bg-[var(--color-surface)]">
                  <button
                    type="submit"
                    disabled={submitting}
                    className="w-full bg-[var(--color-primary)] hover:opacity-90 text-[var(--color-on-primary)] font-semibold text-base py-3 px-6 rounded-xl shadow-sm transition-opacity focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[var(--color-primary)] flex justify-center items-center gap-2 border-0 cursor-pointer disabled:opacity-50"
                  >
                    {submitting ? 'Submitting...' : 'Submit Feedback'}
                    {!submitting && <Send className="w-4 h-4" />}
                  </button>
                  <p className="text-xs text-center text-[var(--color-on-surface-variant)] mt-4">
                    Protected by TrustEstate Security Guarantee
                  </p>
                </div>
              </form>
            )}
          </div>
        </div>
      )}
    </>
  );
}
