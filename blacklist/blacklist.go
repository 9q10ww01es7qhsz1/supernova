package blacklist

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Blacklist struct {
	httpClient      *http.Client
	subscriptionsMu sync.RWMutex
	subscriptions   map[string]string
	stateMu         sync.RWMutex
	state           map[string]struct{}
}

func New(httpClient *http.Client) *Blacklist {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Blacklist{
		httpClient:    httpClient,
		subscriptions: map[string]string{},
		state:         map[string]struct{}{},
	}
}

func (b *Blacklist) Subscribe(blacklistURL string) {
	b.subscriptionsMu.Lock()
	b.subscriptions[blacklistURL] = ""
	b.subscriptionsMu.Unlock()
}

func (b *Blacklist) Watch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var hosts []string

		for blacklistURL, sum := range b.copySubscriptions() {
			newHosts, err := b.updateList(ctx, blacklistURL, sum)
			if err != nil {
				log.Println(fmt.Errorf("failed to update subscription (%s): %w", blacklistURL, err))
				continue
			}

			hosts = append(hosts, newHosts...)
		}

		if len(hosts) > 0 {
			b.updateState(hosts)
		}

		time.Sleep(time.Minute)
	}
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

func (b *Blacklist) fetch(ctx context.Context, blacklistURL string) ([]string, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", blacklistURL, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, resp.Body)
		return nil, "", fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}

	var (
		hasher  = sha256.New()
		scanner = bufio.NewScanner(io.TeeReader(resp.Body, hasher))
	)

	var hosts []string

	for scanner.Scan() {
		host := scanner.Text()

		if host == "" || strings.HasPrefix(host, "#") {
			continue
		}

		if strings.HasPrefix(host, "0.0.0.0") || strings.HasPrefix(host, "127.0.0.1") {
			parts := strings.Split(host, " ")
			host = parts[len(parts)-1]
		}

		hosts = append(hosts, host)
	}

	if err = scanner.Err(); err != nil {
		io.Copy(ioutil.Discard, resp.Body)
		return nil, "", fmt.Errorf("failed to scan response body: %w", err)
	}

	return hosts, hex.EncodeToString(hasher.Sum(nil)), nil
}

func (b *Blacklist) updateState(hosts []string) {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()

	for _, host := range hosts {
		b.state[host] = struct{}{}
	}
}

func (b *Blacklist) Contains(host string) bool {
	b.stateMu.RLock()
	_, blacklisted := b.state[host]
	b.stateMu.RUnlock()
	return blacklisted
}
