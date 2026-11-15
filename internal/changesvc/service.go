package changesvc

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"netorchestrator/internal/preflight"
)

// Status of a change request.
type Status string

const (
	PendingApproval Status = "pending_approval"
	Approved        Status = "approved"
	Rejected        Status = "rejected"
	Cancelled       Status = "cancelled"
	Applied         Status = "applied"
)

// Event type names written to the audit log.
const (
	EvSubmitted = "change.submitted"
	EvApproved  = "change.approved"
	EvRejected  = "change.rejected"
	EvCancelled = "change.cancelled"
	EvApplied   = "change.applied"
)

// Domain errors (mapped to HTTP statuses by the transport layer).
var (
	ErrNotFound          = errors.New("change not found")
	ErrInvalidState      = errors.New("change is not in a state that allows this action")
	ErrSelfApproval      = errors.New("requester cannot review their own change")
	ErrDuplicateApproval = errors.New("already approved by this user")
	ErrForbidden         = errors.New("only the requester may do this")
	ErrInvalid           = errors.New("invalid request")
)

// Analyzer assesses a planned change against a topology. The default
// implementation is in-process; HTTPAnalyzer calls topology-svc.
type Analyzer interface {
	Assess(ctx context.Context, nodes []preflight.Node, links []preflight.Link, c preflight.Change) (*preflight.Report, error)
}

// LocalAnalyzer runs preflight in-process.
type LocalAnalyzer struct{}

func (LocalAnalyzer) Assess(_ context.Context, n []preflight.Node, l []preflight.Link, c preflight.Change) (*preflight.Report, error) {
	return preflight.Assess(n, l, c)
}

// Change is the current state of a change request, derived purely by folding
// its events.
type Change struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Requester    string            `json:"requester"`
	Change       preflight.Change  `json:"change"`
	TopologyHash string            `json:"topology_hash"`
	Risk         preflight.Risk    `json:"risk"`
	Report       *preflight.Report `json:"report"`
	Required     int               `json:"approvals_required"`
	Approvers    []string          `json:"approvers"`
	Status       Status            `json:"status"`
	Reason       string            `json:"reason,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// RequiredApprovals is the approval policy: riskier changes need more
// independent people. Every change needs at least one reviewer who is not
// the requester (four-eyes).
func RequiredApprovals(r preflight.Risk) int {
	if r == preflight.RiskCritical {
		return 2
	}
	return 1
}

type submitted struct {
	Title        string            `json:"title"`
	Change       preflight.Change  `json:"change"`
	TopologyHash string            `json:"topology_hash"`
	Report       *preflight.Report `json:"report"`
	Required     int               `json:"required"`
}

// Service holds the in-memory projection of the log and enforces the rules.
// Decisions are made against the projection and then appended, all under one
// lock, so concurrent approvals cannot both observe "not yet approved".
type Service struct {
	log      *Log
	analyzer Analyzer

	mu      sync.Mutex
	changes map[string]*Change
	order   []string
}

// NewService replays the log to rebuild state (event sourcing).
func NewService(log *Log, a Analyzer) (*Service, error) {
	if a == nil {
		a = LocalAnalyzer{}
	}
	s := &Service{log: log, analyzer: a, changes: map[string]*Change{}}
	if err := log.Verify(); err != nil {
		return nil, err
	}
	for _, e := range log.Events("") {
		if err := s.fold(e); err != nil {
			return nil, fmt.Errorf("replay seq %d: %w", e.Seq, err)
		}
	}
	return s, nil
}

// fold applies one event to the projection. It is the ONLY place state
// changes, used identically at replay and at runtime.
func (s *Service) fold(e Event) error {
	switch e.Type {
	case EvSubmitted:
		var d submitted
		if err := json.Unmarshal(e.Data, &d); err != nil {
			return err
		}
		s.changes[e.ChangeID] = &Change{
			ID: e.ChangeID, Title: d.Title, Requester: e.Actor, Change: d.Change, TopologyHash: d.TopologyHash,
			Risk: d.Report.Risk, Report: d.Report, Required: d.Required, Approvers: []string{},
			Status: PendingApproval, CreatedAt: e.Time, UpdatedAt: e.Time,
		}
		s.order = append(s.order, e.ChangeID)
		return nil
	}
	c, ok := s.changes[e.ChangeID]
	if !ok {
		return fmt.Errorf("%w: event for unknown change %s", ErrNotFound, e.ChangeID)
	}
	c.UpdatedAt = e.Time
	switch e.Type {
	case EvApproved:
		c.Approvers = append(c.Approvers, e.Actor)
		if len(c.Approvers) >= c.Required {
			c.Status = Approved
		}
	case EvRejected:
		var d struct{ Reason string }
		_ = json.Unmarshal(e.Data, &d)
		c.Status, c.Reason = Rejected, d.Reason
	case EvCancelled:
		c.Status = Cancelled
	case EvApplied:
		c.Status = Applied
	default:
		return fmt.Errorf("unknown event type %q", e.Type)
	}
	return nil
}

func (s *Service) commit(actor, typ, id string, data any) error {
	e, err := s.log.Append(actor, typ, id, data)
	if err != nil {
		return err
	}
	return s.fold(e)
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "chg_" + hex.EncodeToString(b)
}

// SubmitRequest is the input to Submit.
type SubmitRequest struct {
	Title  string           `json:"title"`
	Change preflight.Change `json:"change"`
	Nodes  []preflight.Node `json:"nodes"`
	Links  []preflight.Link `json:"links"`
}

func topologyHash(n []preflight.Node, l []preflight.Link) string {
	sort.Slice(n, func(i, j int) bool { return n[i].ID < n[j].ID })
	sort.Slice(l, func(i, j int) bool { return l[i].ID < l[j].ID })
	b, _ := json.Marshal(struct {
		N []preflight.Node
		L []preflight.Link
	}{n, l})
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// Submit runs preflight and opens a change request for approval. The report
// and a hash of the exact topology it was computed from are written into the
// log, so reviewers (and auditors) can see what the approval was based on.
func (s *Service) Submit(ctx context.Context, actor string, r SubmitRequest) (*Change, error) {
	if actor == "" {
		return nil, fmt.Errorf("%w: missing actor", ErrInvalid)
	}
	if strings.TrimSpace(r.Title) == "" || len(r.Title) > 200 {
		return nil, fmt.Errorf("%w: title is required (max 200 chars)", ErrInvalid)
	}
	rep, err := s.analyzer.Assess(ctx, r.Nodes, r.Links, r.Change)
	if err != nil {
		return nil, err
	}
	id := newID()
	d := submitted{Title: r.Title, Change: r.Change, TopologyHash: topologyHash(r.Nodes, r.Links), Report: rep, Required: RequiredApprovals(rep.Risk)}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.commit(actor, EvSubmitted, id, d); err != nil {
		return nil, err
	}
	return s.copyOf(id), nil
}

func (s *Service) copyOf(id string) *Change {
	c := *s.changes[id]
	c.Approvers = append([]string{}, c.Approvers...)
	return &c
}

func (s *Service) get(id string) (*Change, error) {
	c, ok := s.changes[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}

// Approve records one reviewer's approval.
func (s *Service) Approve(actor, id string) (*Change, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, err := s.get(id)
	if err != nil {
		return nil, err
	}
	switch {
	case c.Status != PendingApproval:
		return nil, ErrInvalidState
	case actor == c.Requester:
		return nil, ErrSelfApproval
	}
	for _, a := range c.Approvers {
		if a == actor {
			return nil, ErrDuplicateApproval
		}
	}
	if err := s.commit(actor, EvApproved, id, struct{}{}); err != nil {
		return nil, err
	}
	return s.copyOf(id), nil
}

// Reject closes a pending change with a reason.
func (s *Service) Reject(actor, id, reason string) (*Change, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, err := s.get(id)
	if err != nil {
		return nil, err
	}
	if c.Status != PendingApproval {
		return nil, ErrInvalidState
	}
	if actor == c.Requester {
		return nil, ErrSelfApproval // requesters cancel, they do not "reject"
	}
	if err := s.commit(actor, EvRejected, id, map[string]string{"Reason": reason}); err != nil {
		return nil, err
	}
	return s.copyOf(id), nil
}

// Cancel withdraws a change; only the requester may, and only before apply.
func (s *Service) Cancel(actor, id string) (*Change, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, err := s.get(id)
	if err != nil {
		return nil, err
	}
	if actor != c.Requester {
		return nil, ErrForbidden
	}
	if c.Status != PendingApproval && c.Status != Approved {
		return nil, ErrInvalidState
	}
	if err := s.commit(actor, EvCancelled, id, struct{}{}); err != nil {
		return nil, err
	}
	return s.copyOf(id), nil
}

// Apply marks an approved change as executed. Actually pushing configuration
// to devices is the job of an executor that calls this once it has done so.
func (s *Service) Apply(actor, id string) (*Change, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, err := s.get(id)
	if err != nil {
		return nil, err
	}
	if c.Status != Approved {
		return nil, ErrInvalidState
	}
	if err := s.commit(actor, EvApplied, id, struct{}{}); err != nil {
		return nil, err
	}
	return s.copyOf(id), nil
}

// Get returns one change.
func (s *Service) Get(id string) (*Change, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.get(id); err != nil {
		return nil, err
	}
	return s.copyOf(id), nil
}

// List returns changes newest first, optionally filtered by status.
func (s *Service) List(status Status) []*Change {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []*Change{}
	for i := len(s.order) - 1; i >= 0; i-- {
		if c := s.changes[s.order[i]]; status == "" || c.Status == status {
			out = append(out, s.copyOf(c.ID))
		}
	}
	return out
}

// Log exposes the audit log for verification endpoints.
func (s *Service) Log() *Log { return s.log }
