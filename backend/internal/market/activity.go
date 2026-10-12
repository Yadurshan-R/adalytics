package market

import (
	"fmt"
	"time"
)

// maxActivity is how many recent activity messages the loading page can show.
const maxActivity = 12

// Activity is one thing the backend did during the first load, in plain words,
// such as "Listed 2,000 Minswap transactions". The loading page shows the most
// recent ones, so people can see the backend is working and what it is on.
type Activity struct {
	Time time.Time `json:"time"`
	Text string    `json:"text"`
}

// Note writes a message to the log and, while the first load is running, also
// keeps it for the loading page. It is exported so main can report Koios
// retries ("Koios is busy, trying again in 15s") the same way.
//
// Do not call Note while holding t.mu: it takes the lock itself.
func (t *Tracker) Note(format string, args ...any) {
	text := fmt.Sprintf(format, args...)
	t.log.Print(text)

	t.mu.Lock()
	defer t.mu.Unlock()
	if t.loaded {
		return
	}
	t.activity = append(t.activity, Activity{Time: t.now(), Text: text})
	if len(t.activity) > maxActivity {
		t.activity = t.activity[len(t.activity)-maxActivity:]
	}
}

// recentActivity returns the kept messages, newest first, as a copy so the
// caller can use it after the lock is released. Call with t.mu held.
func (t *Tracker) recentActivity() []Activity {
	out := make([]Activity, len(t.activity))
	for i, a := range t.activity {
		out[len(t.activity)-1-i] = a
	}
	return out
}

// timeLeft guesses the seconds left in the first load from how long it has
// taken so far and how far it has got. It returns 0 (unknown) until at least
// 5% is done, because earlier guesses jump around too much.
func timeLeft(elapsed time.Duration, percent int) int {
	if percent < 5 || percent >= 100 {
		return 0
	}
	left := elapsed * time.Duration(100-percent) / time.Duration(percent)
	return int(left / time.Second)
}
