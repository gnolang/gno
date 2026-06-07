package valsigner

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	rssrv "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/server"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

type Phase string

const (
	PhaseProposal  Phase = "proposal"
	PhasePrevote   Phase = "prevote"
	PhasePrecommit Phase = "precommit"
)

type Action string

const (
	ActionDrop  Action = "drop"
	ActionDelay Action = "delay"
)

var (
	ErrUnknownSignBytes = errors.New("valsigner: unknown sign bytes")
	ErrRuleDropped      = errors.New("valsigner: signature dropped by control rule")
)

type SignedTarget struct {
	Phase  Phase
	Height int64
	Round  int
}

type Rule struct {
	Action Action        `json:"action"`
	Height *int64        `json:"height,omitempty"`
	Round  *int          `json:"round,omitempty"`
	Delay  time.Duration `json:"-"`
}

type RuleView struct {
	Action Action `json:"action"`
	Height *int64 `json:"height,omitempty"`
	Round  *int   `json:"round,omitempty"`
	Delay  string `json:"delay,omitempty"`
}

type ruleRequest struct {
	Action Action `json:"action"`
	Height *int64 `json:"height,omitempty"`
	Round  *int   `json:"round,omitempty"`
	Delay  string `json:"delay,omitempty"`
}

type PhaseStats struct {
	Matched   int64      `json:"matched"`
	Dropped   int64      `json:"dropped"`
	Delayed   int64      `json:"delayed"`
	LastMatch *TargetHit `json:"last_match,omitempty"`
}

type TargetHit struct {
	Height int64     `json:"height"`
	Round  int       `json:"round"`
	At     time.Time `json:"at"`
}

type StateView struct {
	Address crypto.Address       `json:"address"`
	PubKey  string               `json:"pub_key"`
	Rules   map[Phase]*RuleView  `json:"rules"`
	Stats   map[Phase]PhaseStats `json:"stats"`
}

type Controller struct {
	mu    sync.RWMutex
	rules map[Phase]Rule
	stats map[Phase]PhaseStats
}

func NewController() *Controller {
	return &Controller{
		rules: make(map[Phase]Rule),
		stats: map[Phase]PhaseStats{
			PhaseProposal:  {},
			PhasePrevote:   {},
			PhasePrecommit: {},
		},
	}
}

func ParsePhase(raw string) (Phase, error) {
	switch Phase(strings.ToLower(strings.TrimSpace(raw))) {
	case PhaseProposal:
		return PhaseProposal, nil
	case PhasePrevote:
		return PhasePrevote, nil
	case PhasePrecommit:
		return PhasePrecommit, nil
	default:
		return "", fmt.Errorf("unknown phase %q", raw)
	}
}

func ParseRuleRequest(raw ruleRequest) (Rule, error) {
	rule := Rule{
		Action: raw.Action,
		Height: raw.Height,
		Round:  raw.Round,
	}

	switch raw.Action {
	case ActionDrop:
		if raw.Delay != "" {
			return Rule{}, errors.New("delay must be omitted for drop rules")
		}
	case ActionDelay:
		if raw.Delay == "" {
			return Rule{}, errors.New("delay is required for delay rules")
		}

		delay, err := time.ParseDuration(raw.Delay)
		if err != nil {
			return Rule{}, fmt.Errorf("invalid delay: %w", err)
		}
		if delay <= 0 {
			return Rule{}, errors.New("delay must be > 0")
		}
		rule.Delay = delay
	default:
		return Rule{}, fmt.Errorf("unknown action %q", raw.Action)
	}

	return rule, nil
}

func (r Rule) Matches(target SignedTarget) bool {
	if r.Height != nil && *r.Height != target.Height {
		return false
	}
	if r.Round != nil && *r.Round != target.Round {
		return false
	}
	return true
}

func (r Rule) View() *RuleView {
	view := &RuleView{
		Action: r.Action,
		Height: r.Height,
		Round:  r.Round,
	}
	if r.Delay > 0 {
		view.Delay = r.Delay.String()
	}
	return view
}

func (c *Controller) SetRule(phase Phase, rule Rule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rules[phase] = rule
}

func (c *Controller) ClearRule(phase Phase) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.rules, phase)
}

func (c *Controller) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rules = make(map[Phase]Rule)
}

func (c *Controller) Snapshot() (map[Phase]*RuleView, map[Phase]PhaseStats) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rules := make(map[Phase]*RuleView, len(c.rules))
	for phase, rule := range c.rules {
		rules[phase] = rule.View()
	}

	stats := make(map[Phase]PhaseStats, len(c.stats))
	for phase, stat := range c.stats {
		stats[phase] = stat
	}

	return rules, stats
}

func (c *Controller) Evaluate(target SignedTarget) (Rule, bool) {
	c.mu.RLock()
	rule, ok := c.rules[target.Phase]
	c.mu.RUnlock()
	if !ok || !rule.Matches(target) {
		return Rule{}, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	stat := c.stats[target.Phase]
	stat.Matched++
	now := time.Now().UTC()
	stat.LastMatch = &TargetHit{
		Height: target.Height,
		Round:  target.Round,
		At:     now,
	}
	switch rule.Action {
	case ActionDrop:
		stat.Dropped++
	case ActionDelay:
		stat.Delayed++
	}
	c.stats[target.Phase] = stat

	return rule, true
}

type ControllableSigner struct {
	signer     types.Signer
	controller *Controller
	logger     *slog.Logger
}

func NewControllableSigner(signer types.Signer, controller *Controller, logger *slog.Logger) *ControllableSigner {
	if logger == nil {
		logger = slog.Default()
	}

	return &ControllableSigner{
		signer:     signer,
		controller: controller,
		logger:     logger,
	}
}

func (c *ControllableSigner) PubKey() crypto.PubKey {
	return c.signer.PubKey()
}

func (c *ControllableSigner) Sign(signBytes []byte) ([]byte, error) {
	target, err := ClassifySignBytes(signBytes)
	if err == nil {
		if rule, ok := c.controller.Evaluate(target); ok {
			switch rule.Action {
			case ActionDrop:
				c.logger.Info("dropping signature",
					"phase", target.Phase,
					"height", target.Height,
					"round", target.Round,
				)
				return nil, ErrRuleDropped
			case ActionDelay:
				c.logger.Info("delaying signature",
					"phase", target.Phase,
					"height", target.Height,
					"round", target.Round,
					"delay", rule.Delay,
				)
				time.Sleep(rule.Delay)
			}
		}
	} else if !errors.Is(err, ErrUnknownSignBytes) {
		c.logger.Warn("unable to inspect sign bytes", "err", err)
	}

	return c.signer.Sign(signBytes)
}

func (c *ControllableSigner) Close() error {
	return c.signer.Close()
}

func ClassifySignBytes(signBytes []byte) (SignedTarget, error) {
	var proposal types.CanonicalProposal
	if err := amino.UnmarshalSized(signBytes, &proposal); err == nil && proposal.Type == types.ProposalType {
		return SignedTarget{
			Phase:  PhaseProposal,
			Height: proposal.Height,
			Round:  int(proposal.Round),
		}, nil
	}

	var vote types.CanonicalVote
	if err := amino.UnmarshalSized(signBytes, &vote); err == nil {
		switch vote.Type {
		case types.PrevoteType:
			return SignedTarget{
				Phase:  PhasePrevote,
				Height: vote.Height,
				Round:  int(vote.Round),
			}, nil
		case types.PrecommitType:
			return SignedTarget{
				Phase:  PhasePrecommit,
				Height: vote.Height,
				Round:  int(vote.Round),
			}, nil
		}
	}

	return SignedTarget{}, ErrUnknownSignBytes
}

type Server struct {
	logger       *slog.Logger
	controlAddr  string
	remoteAddr   string
	signerServer *rssrv.RemoteSignerServer
	httpServer   *http.Server
	controller   *Controller
	signer       *ControllableSigner
}

func NewServer(keyFile, controlAddr, remoteAddr string, logger *slog.Logger) (*Server, error) {
	if logger == nil {
		logger = slog.Default()
	}
	if keyFile == "" {
		return nil, errors.New("key file is required")
	}
	if controlAddr == "" {
		return nil, errors.New("control address is required")
	}
	if remoteAddr == "" {
		return nil, errors.New("remote signer address is required")
	}

	localSigner, err := local.LoadOrMakeLocalSigner(keyFile)
	if err != nil {
		return nil, fmt.Errorf("load signer key: %w", err)
	}

	controller := NewController()
	signer := NewControllableSigner(localSigner, controller, logger.With("component", "controllable-signer"))

	signerServer, err := rssrv.NewRemoteSignerServer(
		signer,
		remoteAddr,
		logger.With("component", "remote-signer"),
	)
	if err != nil {
		return nil, fmt.Errorf("create remote signer server: %w", err)
	}

	server := &Server{
		logger:       logger,
		controlAddr:  controlAddr,
		remoteAddr:   remoteAddr,
		signerServer: signerServer,
		controller:   controller,
		signer:       signer,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", server.handleHealthz)
	mux.HandleFunc("/state", server.handleState)
	mux.HandleFunc("/rules/", server.handleRule)
	mux.HandleFunc("/reset", server.handleReset)
	server.httpServer = &http.Server{
		Addr:              controlAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return server, nil
}

func (s *Server) Start() error {
	if err := s.signerServer.Start(); err != nil {
		return err
	}

	go func() {
		s.logger.Info("control API listening", "addr", s.controlAddr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("control API server exited", "err", err)
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	httpErr := s.httpServer.Close()
	signerErr := s.signerServer.Stop()
	closeErr := s.signer.Close()

	switch {
	case httpErr != nil:
		return httpErr
	case signerErr != nil:
		return signerErr
	default:
		return closeErr
	}
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	rules, stats := s.controller.Snapshot()
	writeJSON(w, http.StatusOK, StateView{
		Address: s.signer.PubKey().Address(),
		PubKey:  crypto.PubKeyToBech32(s.signer.PubKey()),
		Rules:   rules,
		Stats:   stats,
	})
}

func (s *Server) handleRule(w http.ResponseWriter, r *http.Request) {
	phaseName := strings.TrimPrefix(r.URL.Path, "/rules/")
	phase, err := ParsePhase(phaseName)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch r.Method {
	case http.MethodPut:
		var req ruleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid json body: %v", err))
			return
		}

		rule, err := ParseRuleRequest(req)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		s.controller.SetRule(phase, rule)
		writeJSON(w, http.StatusOK, map[string]any{
			"phase": phase,
			"rule":  rule.View(),
		})
	case http.MethodDelete:
		s.controller.ClearRule(phase)
		writeJSON(w, http.StatusOK, map[string]any{
			"phase":   phase,
			"cleared": true,
		})
	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	s.controller.Reset()
	writeJSON(w, http.StatusOK, map[string]bool{"reset": true})
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]string{"error": message})
}
