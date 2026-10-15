package market

import (
	"context"
	"encoding/json"
	"time"
)

// A token's minting policy decides whether more of it can ever be made. On
// Cardano the policy ID is the hash of that script, so one script_info request
// covers 25 tokens. The page shows it as Security:
//   - "locked": a native script that can no longer sign a mint (its time lock
//     has passed), so the supply is final;
//   - "open": the creator's key can still mint more;
//   - "script": a smart contract (Plutus) decides, which cannot be read simply;
//   - "unknown": Koios has no script for the policy.

// shelleySlot0 turns a mainnet slot into unix time: slot = unix - shelleySlot0.
const shelleySlot0 = 1591566291

// nativeScript is a native script in the JSON form Koios returns.
type nativeScript struct {
	Type     string         `json:"type"` // sig, all, any, atLeast, before, after
	Slot     int64          `json:"slot"`
	Required int            `json:"required"`
	Scripts  []nativeScript `json:"scripts"`
}

// canMint reports whether the script could still allow a mint at slot now or
// later. A key (sig) can always sign; "before" closes at its slot; "after"
// opens at its slot, so it stays possible.
func (n nativeScript) canMint(now int64) bool {
	switch n.Type {
	case "sig", "after":
		return true
	case "before":
		return now < n.Slot
	case "all", "any", "atLeast":
		ok := 0
		for _, c := range n.Scripts {
			if c.canMint(now) {
				ok++
			}
		}
		need := map[string]int{"all": len(n.Scripts), "any": 1, "atLeast": n.Required}[n.Type]
		return ok >= need
	}
	return true // a type we do not know: do not call it locked
}

// mintPolicy is what was read about a token's policy.
type mintPolicy struct {
	kind   string // "native", "plutus" or "unknown"
	script nativeScript
}

// state is the policy as the page shows it at time now.
func (m *mintPolicy) state(now time.Time) string {
	switch {
	case m == nil:
		return "" // not read yet
	case m.kind == "plutus":
		return "script"
	case m.kind != "native":
		return "unknown"
	case m.script.canMint(now.Unix() - shelleySlot0):
		return "open"
	}
	return "locked"
}

// readPolicies reads the minting script of every given policy, 25 per request.
func (t *Tracker) readPolicies(ctx context.Context, policies []string) (map[string]*mintPolicy, error) {
	out := map[string]*mintPolicy{}
	for start := 0; start < len(policies); start += 25 {
		batch := policies[start:min(start+25, len(policies))]
		infos, err := t.client.ScriptInfos(ctx, batch)
		if err != nil {
			return nil, err
		}
		for _, p := range batch {
			out[p] = &mintPolicy{kind: "unknown"}
		}
		for _, s := range infos {
			m := out[s.ScriptHash]
			if m == nil {
				continue
			}
			switch s.Type {
			case "timelock", "multisig":
				if json.Unmarshal(s.Value, &m.script) == nil && m.script.Type != "" {
					m.kind = "native"
				}
			case "plutusV1", "plutusV2", "plutusV3":
				m.kind = "plutus"
			}
		}
	}
	return out, nil
}
