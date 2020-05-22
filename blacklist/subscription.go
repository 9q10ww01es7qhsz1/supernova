package blacklist

import (
	"context"
	"fmt"
	"log"
)

func (b *Blacklist) Subscribe(blacklistURL string) {
	b.subscriptionsMu.Lock()
	b.subscriptions[blacklistURL] = ""
	b.subscriptionsMu.Unlock()
}

func (b *Blacklist) copySubscriptions() map[string]string {
	b.subscriptionsMu.RLock()
	defer b.subscriptionsMu.RUnlock()

	copy := map[string]string{}

	for blacklistURL, sum := range b.subscriptions {
		copy[blacklistURL] = sum
	}

	return copy
}

func (b *Blacklist) updateList(ctx context.Context, blacklistURL, sum string) ([]string, error) {
	hosts, newSum, err := b.fetch(ctx, blacklistURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	if newSum == sum {
		// log.Println("same hash", blacklistURL)
		return nil, nil
	}

	b.subscriptionsMu.Lock()
	b.subscriptions[blacklistURL] = newSum
	b.subscriptionsMu.Unlock()

	log.Println("blacklist subscription updated", blacklistURL, newSum)

	return hosts, nil
}
