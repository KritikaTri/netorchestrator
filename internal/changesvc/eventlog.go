package changesvc

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// Event is one immutable fact in the audit log. Hash covers every other
// field plus PrevHash, so altering, deleting or reordering any earlier event
// invalidates every later hash.
type Event struct {
	Seq      uint64          `json:"seq"`
	Time     time.Time       `json:"time"`
	Actor    string          `json:"actor"`
	Type     string          `json:"type"`
	ChangeID string          `json:"change_id"`
	Data     json.RawMessage `json:"data"`
	PrevHash string          `json:"prev_hash"`
	Hash     string          `json:"hash"`
}

// ErrTampered is returned by Verify when the chain does not check out.
var ErrTampered = errors.New("changesvc: audit log integrity violation")

// computeHash length-prefixes each field so no two distinct events can
// serialise to the same byte string (e.g. actor "a|b" vs actor "a", type "b").
func computeHash(e Event) string {
	h := sha256.New()
	for _, f := range []string{
		strconv.FormatUint(e.Seq, 10), e.Time.UTC().Format(time.RFC3339Nano), e.Actor, e.Type,
		e.ChangeID, string(e.Data), e.PrevHash,
	} {
		h.Write([]byte(strconv.Itoa(len(f))))
		h.Write([]byte{':'})
		h.Write([]byte(f))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Log is an append-only, hash-chained event log, optionally persisted as
// JSON lines. It is safe for concurrent use.
type Log struct {
	mu     sync.Mutex
	events []Event
	f      *os.File
	sync   bool
	now    func() time.Time
}

// NewMemLog returns a non-persistent log (tests, demos).
func NewMemLog() *Log { return &Log{now: time.Now} }

// OpenLog opens (or creates) a file-backed log, loading and verifying any
// existing events. If syncWrites is true every append is fsync'd.
func OpenLog(path string, syncWrites bool) (*Log, error) {
	l := &Log{now: time.Now, sync: syncWrites}
	if f, err := os.Open(path); err == nil {
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1<<20), 16<<20)
		for sc.Scan() {
			var e Event
			if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
				f.Close()
				return nil, fmt.Errorf("%w: unreadable line %d", ErrTampered, len(l.events)+1)
			}
			l.events = append(l.events, e)
		}
		f.Close()
		if err := sc.Err(); err != nil {
			return nil, err
		}
		if err := l.Verify(); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	l.f = f
	return l, nil
}

// Append adds an event, chaining it to the previous one.
func (l *Log) Append(actor, typ, changeID string, data any) (Event, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return Event{}, err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	e := Event{Seq: uint64(len(l.events)) + 1, Time: l.now().UTC(), Actor: actor, Type: typ, ChangeID: changeID, Data: raw}
	if n := len(l.events); n > 0 {
		e.PrevHash = l.events[n-1].Hash
	}
	e.Hash = computeHash(e)
	if l.f != nil {
		line, _ := json.Marshal(e)
		if _, err := l.f.Write(append(line, '\n')); err != nil {
			return Event{}, err // not appended in memory either: memory never runs ahead of disk
		}
		if l.sync {
			if err := l.f.Sync(); err != nil {
				return Event{}, err
			}
		}
	}
	l.events = append(l.events, e)
	return e, nil
}

// Events returns a copy of all events (optionally for one change).
func (l *Log) Events(changeID string) []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []Event
	for _, e := range l.events {
		if changeID == "" || e.ChangeID == changeID {
			out = append(out, e)
		}
	}
	return out
}

// Head returns the latest sequence number and hash. Publishing the head
// somewhere the service cannot rewrite (a ticket, a git tag, a WORM bucket)
// also detects truncation, which a pure hash chain cannot.
func (l *Log) Head() (uint64, string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if n := len(l.events); n > 0 {
		return l.events[n-1].Seq, l.events[n-1].Hash
	}
	return 0, ""
}

// Verify recomputes the whole chain.
func (l *Log) Verify() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	prev := ""
	for i, e := range l.events {
		switch {
		case e.Seq != uint64(i)+1:
			return fmt.Errorf("%w: sequence gap at position %d (seq %d)", ErrTampered, i+1, e.Seq)
		case e.PrevHash != prev:
			return fmt.Errorf("%w: broken link at seq %d", ErrTampered, e.Seq)
		case computeHash(e) != e.Hash:
			return fmt.Errorf("%w: content altered at seq %d", ErrTampered, e.Seq)
		}
		prev = e.Hash
	}
	return nil
}

// Close flushes and closes the backing file.
func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f == nil {
		return nil
	}
	err := l.f.Close()
	l.f = nil
	return err
}
