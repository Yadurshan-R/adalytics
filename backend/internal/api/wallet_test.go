// Tests for the wallet address check on /api/wallets/{wallet}.
package api

import "testing"

// TestValidWallet checks only real-looking stake and payment addresses pass.
func TestValidWallet(t *testing.T) {
	good := []string{
		"stake1u82zmwat9k3qyv4lxr2dnf0p7q4c5y6h8j9k0l2m3n4p5q6r7s8t9",
		"addr1qy68dr9w0e5rqz3n8a1k2j3h4g5f6d7s8a9q0w1e2r3t4y5u6i7o8p9a0s1d2f3g4h",
	}
	bad := []string{"", "stake1short", "STAKE1U82ZMWAT9K3QYV4LXR2DNF0P7Q4C5Y6H8J9K0L2M3N4", "stake1u82zmwat9k3qyv4lxr2dn/../../etc/passwd0000000000", "hello1u82zmwat9k3qyv4lxr2dnf0p7q4c5y6h8j9k0l2m3n4p5q6"}
	for _, w := range good {
		if !validWallet(w) {
			t.Errorf("%q should pass", w)
		}
	}
	for _, w := range bad {
		if validWallet(w) {
			t.Errorf("%q should fail", w)
		}
	}
}
