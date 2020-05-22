package main

import (
	"bufio"
	"fmt"
	"os"
)

func readSubscriptions(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	subscriptions := []string{}

	for scanner.Scan() {
		subscriptions = append(subscriptions, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan subscriptions: %w", err)
	}

	return subscriptions, nil
}
