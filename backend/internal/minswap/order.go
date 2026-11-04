package minswap

import "cardano-analytics/internal/koios"

// A Minswap V2 order waits at the order script until a batcher swaps it. Its
// inline datum says who gets the result and which pool the order is for:
//
//	field 0  canceller
//	field 1  refund receiver (an address)
//	field 2  refund receiver's datum
//	field 3  success receiver: the address the swap's output is paid to
//	field 4  success receiver's datum
//	field 5  the LP asset of the pool the order is for
//	field 6  the order step (swap exact in, ...)
//
// An address is Constr 0 [payment credential, stake credential], and a
// payment credential is Constr 0 [key hash] or Constr 1 [script hash].

// Order is what an order's datum says about where its result goes.
type Order struct {
	Receiver string      // payment credential (hex) of the success receiver
	Refund   string      // payment credential (hex) of the refund receiver: paid when the batcher kills the order
	LPAsset  koios.Asset // the pool the order is for
}

// ReadOrder reads a Minswap V2 order's datum. ok is false when the datum is
// missing or not shaped like a Minswap V2 order.
func ReadOrder(d *koios.InlineDatum) (Order, bool) {
	if d == nil || len(d.Value.Fields) < 6 {
		return Order{}, false
	}
	f := d.Value.Fields
	success, refund := paymentCred(f[3]), paymentCred(f[1])
	if success == "" || refund == "" {
		return Order{}, false
	}
	lp, err := asset(f[5])
	if err != nil || lp.PolicyID != LPPolicy {
		return Order{}, false
	}
	return Order{Receiver: success, Refund: refund, LPAsset: lp}, true
}

// paymentCred reads the payment credential (hex) of an address in a datum:
// Constr 0 [Constr 0|1 [hash], stake credential]. It is "" when the field is
// not shaped like an address.
func paymentCred(addr koios.PlutusData) string {
	if len(addr.Fields) < 1 || len(addr.Fields[0].Fields) != 1 || addr.Fields[0].Fields[0].Bytes == nil {
		return ""
	}
	return *addr.Fields[0].Fields[0].Bytes
}
