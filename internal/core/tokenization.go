package core

import "errors"

type FractionalPool struct {
	ID             string  `json:"id"`
	PropertyID     string  `json:"propertyId"`
	TotalTokens    int64   `json:"totalTokens"`
	TokensSold     int64   `json:"tokensSold"`
	TokenPrice     float64 `json:"tokenPrice"`
}

type FractionalHolding struct {
	ID          string `json:"id"`
	PoolID      string `json:"poolId"`
	InvestorID  string `json:"investorId"`
	TokenCount  int64  `json:"tokenCount"`
}

// ValidateTokenPurchase checks if a purchase requested exceeds available pool bounds.
func (p *FractionalPool) ValidateTokenPurchase(requestedTokens int64) error {
	if requestedTokens <= 0 {
		return errors.New("requested token count must be positive")
	}
	if p.TokensSold+requestedTokens > p.TotalTokens {
		return errors.New("insufficient tokens available in the fractional pool")
	}
	return nil
}
