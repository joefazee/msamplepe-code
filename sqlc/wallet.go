package db

import (
	"fmt"
)

// GetBalanceAsFloat64 converts the balance to a float64
func (w *Wallet) GetBalanceAsFloat64() float64 {
	f, _ := w.Balance.Float64()
	return f
}

// Balance2String converts the balance to a string
func (w *Wallet) Balance2String() string {
	return fmt.Sprintf("%.2f", w.GetBalanceAsFloat64())
}
