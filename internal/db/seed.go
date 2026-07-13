package db

import (
	"fmt"
	"os"

	"github.com/realestate-trust/monorepo/internal/core"
)

// ShouldSeed returns true when the environment is non-production.
// Controlled by APP_ENV; defaults to "development" when unset.
func ShouldSeed() bool {
	env := os.Getenv("APP_ENV")
	return env != "production"
}

// SeedProperties populates the property registry with demo properties.
func SeedProperties(repo PropertyRepository) {
	properties := []struct {
		Address     string
		Description string
		Value       float64
		Thumbnail   string
		OwnerID     string
	}{
		{"123 Ocean View Dr, Malibu", "Luxury beachfront villa", 4500000, "https://images.unsplash.com/photo-1512917774080-9991f1c4c750?w=800", "usr-priya.sharma@realestate.in"},
		{"456 Silicon Ave, San Jose", "Modern tech hub office space", 7800000, "https://images.unsplash.com/photo-1486406146926-c627a92ad1ab?w=800", "usr-priya.sharma@realestate.in"},
		{"789 Alpine Way, Aspen", "Ski-in/ski-out cabin", 12500000, "https://images.unsplash.com/photo-1518780664697-55e3ad937233?w=800", "usr-priya.sharma@realestate.in"},
	}

	for _, p := range properties {
		prop, _ := repo.CreateProperty(p.Address, p.Description, p.Value, p.Thumbnail, p.OwnerID)
		fmt.Printf("  [seed] Property: %s — ₹%.0f (ID: %s)\n", prop.Address, prop.Value, prop.ID)
	}
}

// SeedUsers populates the identity store with demo accounts and KYC submissions.
func SeedUsers(repo *InMemoryUserRepository) {
	users := []struct {
		Email    string
		FullName string
		Role     core.UserRole
	}{
		{"aryan.dev@realestate.in", "Aryan Dev", core.Buyer},
		{"priya.sharma@realestate.in", "Priya Sharma", core.Seller},
		{"ravi.kumar@realestate.in", "Ravi Kumar", core.Broker},
	}

	for _, u := range users {
		user, _ := repo.CreateUser(u.Email, "dummyhash", u.FullName, u.Role)
		// Submit KYC for seed users
		repo.SubmitKYC(user.ID, "PASSPORT", "SEED-"+user.ID)
		fmt.Printf("  [seed] User: %s (%s) — KYC SUBMITTED\n", user.FullName, user.ID)
	}
}

// SeedTransactions populates the transaction store with demo escrow deals.
func SeedTransactions(repo *InMemoryTransactionRepository) {
	txs := []struct {
		PropertyID  string
		BuyerID     string
		SellerID    string
		TotalAmount float64
		Advance     bool
	}{
		{"prop-101", "usr-aryan.dev@realestate.in", "usr-priya.sharma@realestate.in", 4500000, true},
		{"prop-202", "usr-aryan.dev@realestate.in", "usr-priya.sharma@realestate.in", 7800000, false},
		{"prop-303", "usr-ravi.kumar@realestate.in", "usr-priya.sharma@realestate.in", 12500000, false},
	}

	for _, t := range txs {
		tx, _ := repo.CreateTransaction(t.PropertyID, t.BuyerID, t.SellerID, t.TotalAmount)
		if t.Advance {
			// Advance from DRAFT → ESCROW
			repo.UpdateTransactionStatus(tx.ID, core.Escrow)
		}
		fmt.Printf("  [seed] Transaction: %s — ₹%.0f (%s)\n", tx.ID, tx.TotalAmount, tx.Status)
	}
}

// SeedLoans populates the financing store with demo mortgage applications.
func SeedLoans(repo *InMemoryFinancingRepository) {
	loans := []struct {
		TxID      string
		UserID    string
		ReqAmount float64
		Approved  bool
		AppAmount float64
	}{
		{"tx-prop-101", "usr-aryan.dev@realestate.in", 3500000, true, 3200000},
		{"tx-prop-202", "usr-aryan.dev@realestate.in", 5000000, false, 0},
	}

	for _, l := range loans {
		loan, _ := repo.CreateLoan(l.TxID, l.UserID, l.ReqAmount)
		if l.Approved {
			repo.ApproveLoan(loan.ID, l.AppAmount)
		}
		fmt.Printf("  [seed] Loan: %s — ₹%.0f (status: %s)\n", loan.ID, loan.RequestedAmount, loan.Status)
	}
}

// SeedPools populates the tokenization store with demo fractional property pools.
func SeedPools(repo *InMemoryTokenizationRepository) {
	pools := []struct {
		PropertyID  string
		TotalTokens int64
		TokenPrice  float64
		SoldTokens  int64
	}{
		{"prop-101", 1000, 450, 320},
		{"prop-303", 5000, 170, 480},
	}

	for _, p := range pools {
		pool, _ := repo.CreatePool(p.PropertyID, p.TotalTokens, p.TokenPrice)
		if p.SoldTokens > 0 {
			repo.BuyTokens(pool.ID, "usr-aryan.dev@realestate.in", p.SoldTokens)
		}
		fmt.Printf("  [seed] Pool: %s — ₹%.0f/share (%d/%d sold)\n", pool.PropertyID, pool.TokenPrice, p.SoldTokens, p.TotalTokens)
	}
}

// SeedLedger populates the ledger store with a genesis block and sample audit entries.
func SeedLedger(repo *InMemoryLedgerRepository) {
	entries := []string{
		"Genesis Block — RealEstate Trust Ledger Initialized",
		"User Registered: aryan.dev@realestate.in (BUYER)",
		"User Registered: priya.sharma@realestate.in (SELLER)",
		"Transaction Created: tx-prop-101 — ₹4,50,000 INR",
		"Escrow Funded: tx-prop-101 — Status advanced to ESCROW",
		"Loan Approved: loan-tx-prop-101 — ₹32,00,000 INR",
		"Pool Created: pool-prop-101 — 1000 shares @ ₹450/share",
		"Shares Purchased: 320 shares of pool-prop-101 by aryan.dev@realestate.in",
	}

	for _, payload := range entries {
		repo.WriteLog(payload)
	}
	fmt.Printf("  [seed] Ledger: %d audit entries sealed\n", len(entries))
}
