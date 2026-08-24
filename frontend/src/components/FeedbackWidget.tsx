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
        className="fixed bottom-6 right-6 z-50 flex items-center gap-2 px-4 py-3 rounded-full btn-primary shadow-lg hover:shadow-xl hover:scale-105 active:scale-95 transition-all duration-300 border-0 cursor-pointer"
        aria-label="Provide Feedback"
      >
        <MessageSquare className="w-5 h-5" />
        <span className="text-sm font-semibold">Feedback</span>
      </button>

      {/* Backdrop & Modal */}
      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-[var(--color-on-surface)]/40 backdrop-blur-sm transition-opacity duration-300">
          <div
            className="w-full max-w-md p-6 rounded-xl border border-[var(--color-outline-variant)] bg-[var(--color-surface)] shadow-xl relative animate-in fade-in zoom-in-95 duration-200"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Close Button */}
            <button
              onClick={() => setIsOpen(false)}
              className="absolute top-4 right-4 p-1.5 rounded-lg text-[var(--color-on-surface-variant)] hover:text-[var(--color-on-surface)] hover:bg-[var(--color-surface-container-high)] transition-colors border-0 bg-transparent cursor-pointer"
            >
              <X className="w-5 h-5" />
            </button>

            {submitted ? (
              <div className="flex flex-col items-center justify-center py-8 text-center">
                <div className="w-16 h-16 rounded-full bg-[var(--color-success-container)] text-[var(--color-on-success-container)] flex items-center justify-center mb-4">
                  <Star className="w-8 h-8 fill-current" />
                </div>
                <h3 className="text-xl font-bold mb-2 text-[var(--color-on-surface)]">Thank you!</h3>
                <p className="text-[var(--color-on-surface-variant)] text-sm">Your feedback helps us build a better platform.</p>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="flex flex-col gap-5 mt-2">
                <div>
                  <h3 className="text-xl font-bold text-[var(--color-on-surface)]">Help us improve</h3>
                  <p className="text-sm text-[var(--color-on-surface-variant)] mt-1">We value your input on new features, bugs, or improvements.</p>
                </div>

                {error && (
                  <div className="p-3 text-sm bg-[var(--color-error-container)] border border-[var(--color-error)] text-[var(--color-on-error-container)] rounded-lg">
                    {error}
                  </div>
                )}

                {/* Rating */}
                <div className="flex flex-col gap-2">
                  <label className="text-xs font-semibold text-[var(--color-on-surface-variant)] uppercase tracking-wider">Rating</label>
                  <div className="flex gap-2">
                    {[1, 2, 3, 4, 5].map((star) => (
                      <button
                        key={star}
                        type="button"
                        onClick={() => setRating(star)}
                        className="bg-transparent border-0 p-1 cursor-pointer transition-transform hover:scale-110"
                      >
                        <Star
                          className={`w-8 h-8 transition-colors ${
                            star <= rating
                              ? 'text-[var(--color-warning)] fill-[var(--color-warning)]'
                              : 'text-[var(--color-outline-variant)] hover:text-[var(--color-on-surface-variant)]'
                          }`}
                        />
                      </button>
                    ))}
                  </div>
                </div>

                {/* Category */}
                <div className="flex flex-col gap-2">
                  <label className="text-xs font-semibold text-[var(--color-on-surface-variant)] uppercase tracking-wider">Category</label>
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
                        className={`py-2 px-3 text-sm rounded-lg border font-medium cursor-pointer transition-all ${
                          category === cat.id
                            ? 'bg-[var(--color-primary-container)] border-[var(--color-primary)] text-[var(--color-on-primary-container)]'
                            : 'bg-[var(--color-surface-container-low)] border-[var(--color-outline-variant)] text-[var(--color-on-surface)] hover:bg-[var(--color-surface-container-high)]'
                        }`}
                      >
                        {cat.label}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Message */}
                <div className="flex flex-col gap-2">
                  <label htmlFor="feedback-message" className="text-xs font-semibold text-[var(--color-on-surface-variant)] uppercase tracking-wider">
                    Feedback Details
                  </label>
                  <textarea
                    id="feedback-message"
                    rows={4}
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    placeholder="Describe your feedback or request..."
                    className="w-full bg-[var(--color-surface-container-low)] border border-[var(--color-outline-variant)] focus:border-[var(--color-primary)] focus:ring-1 focus:ring-[var(--color-primary)] rounded-lg p-3 text-sm text-[var(--color-on-surface)] placeholder:text-[var(--color-on-surface-variant)] focus:outline-none resize-none transition-all"
                  />
                </div>

                {/* Submit */}
                <button
                  type="submit"
                  disabled={submitting}
                  className="btn-primary w-full flex items-center justify-center gap-2 py-3 mt-2 font-bold cursor-pointer disabled:opacity-50"
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
