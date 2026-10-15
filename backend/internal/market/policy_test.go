package market

import (
	"encoding/json"
	"testing"
	"time"
)

// TestPolicyState checks how a minting policy is shown: locked once no key can
// sign a mint any more, open while one can, and smart contracts apart.
func TestPolicyState(t *testing.T) {
	now := time.Unix(shelleySlot0+1000, 0) // slot 1000
	cases := []struct{ name, script, want string }{
		{"key only", `{"type":"sig","keyHash":"aa"}`, "open"},
		{"key until a past slot", `{"type":"all","scripts":[{"type":"sig","keyHash":"aa"},{"type":"before","slot":900}]}`, "locked"},
		{"key until a future slot", `{"type":"all","scripts":[{"type":"sig","keyHash":"aa"},{"type":"before","slot":2000}]}`, "open"},
		{"key after a slot", `{"type":"all","scripts":[{"type":"sig","keyHash":"aa"},{"type":"after","slot":5000}]}`, "open"},
		{"any: one branch still open", `{"type":"any","scripts":[{"type":"before","slot":900},{"type":"sig","keyHash":"aa"}]}`, "open"},
		{"any: every branch closed", `{"type":"any","scripts":[{"type":"before","slot":900},{"type":"before","slot":10}]}`, "locked"},
		{"2 of 3, only 1 open", `{"type":"atLeast","required":2,"scripts":[{"type":"sig","keyHash":"aa"},{"type":"before","slot":900},{"type":"before","slot":10}]}`, "locked"},
		{"2 of 3, 2 open", `{"type":"atLeast","required":2,"scripts":[{"type":"sig","keyHash":"aa"},{"type":"sig","keyHash":"bb"},{"type":"before","slot":10}]}`, "open"},
		{"all of nothing: anyone can mint", `{"type":"all","scripts":[]}`, "open"},
	}
	for _, c := range cases {
		m := &mintPolicy{kind: "native"}
		if err := json.Unmarshal([]byte(c.script), &m.script); err != nil {
			t.Fatal(err)
		}
		if got := m.state(now); got != c.want {
			t.Errorf("%s: %s, want %s", c.name, got, c.want)
		}
	}
	var none *mintPolicy
	if none.state(now) != "" || (&mintPolicy{kind: "plutus"}).state(now) != "script" || (&mintPolicy{kind: "unknown"}).state(now) != "unknown" {
		t.Error("policy not read, Plutus or unknown shown wrong")
	}
}
