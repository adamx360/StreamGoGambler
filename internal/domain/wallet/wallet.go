package wallet

import "sync"

type Wallet struct {
	mu      sync.Mutex
	balance int
}

func New(initial int) *Wallet {
	return &Wallet{balance: initial}
}

func (w *Wallet) GetBalance() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.balance
}

func (w *Wallet) SetBalance(amount int) {
	w.mu.Lock()
	w.balance = amount
	w.mu.Unlock()
}

func (w *Wallet) AddBalance(delta int) {
	w.mu.Lock()
	w.balance += delta
	w.mu.Unlock()
}

func (w *Wallet) Spend(amount int) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.balance < amount {
		return false
	}
	w.balance -= amount
	return true
}

func (w *Wallet) CanAfford(amount int) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.balance >= amount
}
