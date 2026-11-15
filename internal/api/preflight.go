package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"netorchestrator/internal/models"
	"netorchestrator/internal/preflight"
)

// TopologySource is what the preflight endpoints need from the data layer.
// *services.NetworkService satisfies it; tests use an in-memory fake.
type TopologySource interface {
	GetNetwork(ctx context.Context, id uuid.UUID) (*models.Network, error)
	ListNodes(ctx context.Context, networkID uuid.UUID) ([]models.Node, error)
	ListLinks(ctx context.Context, networkID uuid.UUID) ([]models.Link, error)
}

// PreflightHandlers serves change-impact analysis for a network's topology.
type PreflightHandlers struct {
	src    TopologySource
	logger *zap.Logger
}

// NewPreflightHandlers wires the handlers to a topology source.
func NewPreflightHandlers(src TopologySource, logger *zap.Logger) *PreflightHandlers {
	return &PreflightHandlers{src: src, logger: logger}
}

// PreflightRequest is the body of POST /networks/:id/preflight.
type PreflightRequest struct {
	Kind preflight.ChangeKind `json:"kind" binding:"required"` // remove_node | remove_link
	ID   string               `json:"id" binding:"required"`   // node or link UUID
}

// Preflight reports what would lose connectivity if a node or link is removed.
//
// @Summary  Preflight a topology change
// @Tags     networks
// @Accept   json
// @Produce  json
// @Param    id path string true "Network ID"
// @Param    change body PreflightRequest true "Planned change"
// @Success  200 {object} map[string]interface{}
// @Router   /networks/{id}/preflight [post]
func (h *PreflightHandlers) Preflight(c *gin.Context) {
	networkID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req PreflightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body must be {\"kind\": \"remove_node|remove_link\", \"id\": \"<uuid>\"}"})
		return
	}
	nodes, links, ok := h.load(c, networkID)
	if !ok {
		return
	}
	report, err := preflight.Assess(nodes, links, preflight.Change{Kind: req.Kind, ID: req.ID})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"network_id": networkID, "report": report})
}

// Resilience lists single points of failure in the network's topology.
//
// @Summary  Topology resilience report
// @Tags     networks
// @Produce  json
// @Param    id path string true "Network ID"
// @Success  200 {object} map[string]interface{}
// @Router   /networks/{id}/resilience [get]
func (h *PreflightHandlers) Resilience(c *gin.Context) {
	networkID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	nodes, links, ok := h.load(c, networkID)
	if !ok {
		return
	}
	res, err := preflight.Analyze(nodes, links)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"network_id": networkID, "resilience": res})
}

func (h *PreflightHandlers) load(c *gin.Context, id uuid.UUID) ([]preflight.Node, []preflight.Link, bool) {
	ctx := c.Request.Context()
	if _, err := h.src.GetNetwork(ctx, id); err != nil {
		if err.Error() == "network not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Network not found"})
		} else {
			h.logger.Error("preflight: load network", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load network"})
		}
		return nil, nil, false
	}
	ns, err := h.src.ListNodes(ctx, id)
	if err != nil {
		h.logger.Error("preflight: list nodes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load nodes"})
		return nil, nil, false
	}
	ls, err := h.src.ListLinks(ctx, id)
	if err != nil {
		h.logger.Error("preflight: list links", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load links"})
		return nil, nil, false
	}
	nodes, links := toPreflight(ns, ls)
	return nodes, links, true
}

func toPreflight(ns []models.Node, ls []models.Link) ([]preflight.Node, []preflight.Link) {
	nodes := make([]preflight.Node, 0, len(ns))
	for _, n := range ns {
		nodes = append(nodes, preflight.Node{ID: n.ID.String(), Name: n.Name})
	}
	links := make([]preflight.Link, 0, len(ls))
	for _, l := range ls {
		links = append(links, preflight.Link{ID: l.ID.String(), Source: l.SourceNodeID.String(), Target: l.TargetNodeID.String()})
	}
	return nodes, links
}

func (h *PreflightHandlers) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, preflight.ErrUnknownTarget):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, preflight.ErrInvalidChange):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, preflight.ErrInvalidTopology):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		h.logger.Error("preflight failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "preflight failed"})
	}
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid " + name})
		return uuid.Nil, false
	}
	return id, true
}
