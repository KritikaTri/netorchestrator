package database

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"netorchestrator/internal/models"
)

// SeedDatabase seeds the database with sample data
func SeedDatabase(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Starting database seeding...")

	// Step 1: Create a sample user (admin)
	user := &models.User{
		Username:  "admin",
		Email:     "admin@netorchestrator.com",
		Password:  "hashed_password_here", // In production, use bcrypt
		Role:      models.UserRoleAdmin,
		Status:    models.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Check if user exists, if not create
	var existingUser models.User
	if err := db.Where("username = ?", user.Username).First(&existingUser).Error; err == gorm.ErrRecordNotFound {
		if err := db.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		logger.Info("Created admin user", zap.String("username", user.Username))
	} else {
		user = &existingUser
		logger.Info("Admin user already exists", zap.String("username", user.Username))
	}

	// Step 2: Create Star Network
	starNetwork := &models.Network{
		Name:        "Star Network Topology",
		Description: "Sample star topology network with central hub",
		Status:      models.NetworkStatusActive,
		UserID:      user.ID,
		Config: models.NetworkConfig{
			Topology: "star",
			Subnet:   "10.1.0.0/24",
			Gateway:  "10.1.0.1",
			DNS:      []string{"8.8.8.8", "8.8.4.4"},
			MTU:      1500,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(starNetwork).Error; err != nil {
		return fmt.Errorf("failed to create star network: %w", err)
	}
	logger.Info("Created star network", zap.String("network_id", starNetwork.ID.String()))

	// Step 3: Create Mesh Network
	meshNetwork := &models.Network{
		Name:        "Mesh Network Topology",
		Description: "Sample mesh topology network with full connectivity",
		Status:      models.NetworkStatusActive,
		UserID:      user.ID,
		Config: models.NetworkConfig{
			Topology: "mesh",
			Subnet:   "10.2.0.0/24",
			Gateway:  "10.2.0.1",
			DNS:      []string{"8.8.8.8", "8.8.4.4"},
			MTU:      1500,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(meshNetwork).Error; err != nil {
		return fmt.Errorf("failed to create mesh network: %w", err)
	}
	logger.Info("Created mesh network", zap.String("network_id", meshNetwork.ID.String()))

	// Step 4: Create Nodes for Star Network (central hub + 5 nodes)
	starNodes := []*models.Node{
		{
			NetworkID:  starNetwork.ID,
			Name:       "Star-Hub",
			Type:       models.NodeTypeRouter,
			IPAddress:  "10.1.0.1",
			MACAddress: "00:11:22:33:44:01",
			Status:     models.NodeStatusActive,
			Position:   models.Position{X: 0, Y: 0},
			Config: models.NodeConfig{
				CPU:    4,
				Memory: 8192,
				Storage: 50,
				OS:     "linux",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			NetworkID:  starNetwork.ID,
			Name:       "Star-Node-1",
			Type:       models.NodeTypeHost,
			IPAddress:  "10.1.0.10",
			MACAddress: "00:11:22:33:44:02",
			Status:     models.NodeStatusActive,
			Position:   models.Position{X: 100, Y: 0},
			Config: models.NodeConfig{
				CPU:    2,
				Memory: 4096,
				Storage: 25,
				OS:     "linux",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			NetworkID:  starNetwork.ID,
			Name:       "Star-Node-2",
			Type:       models.NodeTypeHost,
			IPAddress:  "10.1.0.11",
			MACAddress: "00:11:22:33:44:03",
			Status:     models.NodeStatusActive,
			Position:   models.Position{X: -100, Y: 0},
			Config: models.NodeConfig{
				CPU:    2,
				Memory: 4096,
				Storage: 25,
				OS:     "linux",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			NetworkID:  starNetwork.ID,
			Name:       "Star-Node-3",
			Type:       models.NodeTypeHost,
			IPAddress:  "10.1.0.12",
			MACAddress: "00:11:22:33:44:04",
			Status:     models.NodeStatusActive,
			Position:   models.Position{X: 0, Y: 100},
			Config: models.NodeConfig{
				CPU:    2,
				Memory: 4096,
				Storage: 25,
				OS:     "linux",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			NetworkID:  starNetwork.ID,
			Name:       "Star-Node-4",
			Type:       models.NodeTypeHost,
			IPAddress:  "10.1.0.13",
			MACAddress: "00:11:22:33:44:05",
			Status:     models.NodeStatusActive,
			Position:   models.Position{X: 0, Y: -100},
			Config: models.NodeConfig{
				CPU:    2,
				Memory: 4096,
				Storage: 25,
				OS:     "linux",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, node := range starNodes {
		if err := db.Create(node).Error; err != nil {
			return fmt.Errorf("failed to create star node %s: %w", node.Name, err)
		}
		logger.Info("Created star node", zap.String("node", node.Name))
	}

	// Step 5: Create Nodes for Mesh Network (5 nodes)
	meshNodes := []*models.Node{
		{
			NetworkID:  meshNetwork.ID,
			Name:       "Mesh-Node-1",
			Type:       models.NodeTypeRouter,
			IPAddress:  "10.2.0.10",
			MACAddress: "00:11:22:33:44:10",
			Status:     models.NodeStatusActive,
			Position:   models.Position{X: 0, Y: 0},
			Config: models.NodeConfig{
				CPU:    4,
				Memory: 8192,
				Storage: 50,
				OS:     "linux",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			NetworkID:  meshNetwork.ID,
			Name:       "Mesh-Node-2",
			Type:       models.NodeTypeRouter,
			IPAddress:  "10.2.0.11",
			MACAddress: "00:11:22:33:44:11",
			Status:     models.NodeStatusActive,
			Position:   models.Position{X: 150, Y: 0},
			Config: models.NodeConfig{
				CPU:    4,
				Memory: 8192,
				Storage: 50,
				OS:     "linux",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			NetworkID:  meshNetwork.ID,
			Name:       "Mesh-Node-3",
			Type:       models.NodeTypeHost,
			IPAddress:  "10.2.0.12",
			MACAddress: "00:11:22:33:44:12",
			Status:     models.NodeStatusActive,
			Position:   models.Position{X: 150, Y: 150},
			Config: models.NodeConfig{
				CPU:    2,
				Memory: 4096,
				Storage: 25,
				OS:     "linux",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			NetworkID:  meshNetwork.ID,
			Name:       "Mesh-Node-4",
			Type:       models.NodeTypeHost,
			IPAddress:  "10.2.0.13",
			MACAddress: "00:11:22:33:44:13",
			Status:     models.NodeStatusActive,
			Position:   models.Position{X: 0, Y: 150},
			Config: models.NodeConfig{
				CPU:    2,
				Memory: 4096,
				Storage: 25,
				OS:     "linux",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			NetworkID:  meshNetwork.ID,
			Name:       "Mesh-Node-5",
			Type:       models.NodeTypeSwitch,
			IPAddress:  "10.2.0.14",
			MACAddress: "00:11:22:33:44:14",
			Status:     models.NodeStatusActive,
			Position:   models.Position{X: -150, Y: 0},
			Config: models.NodeConfig{
				CPU:    1,
				Memory: 2048,
				Storage: 10,
				OS:     "linux",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, node := range meshNodes {
		if err := db.Create(node).Error; err != nil {
			return fmt.Errorf("failed to create mesh node %s: %w", node.Name, err)
		}
		logger.Info("Created mesh node", zap.String("node", node.Name))
	}

	// Step 6: Create Links for Star Network (hub to each node)
	starLinks := []*models.Link{
		{
			NetworkID:   starNetwork.ID,
			SourceNodeID: starNodes[0].ID, // Hub
			TargetNodeID: starNodes[1].ID, // Node 1
			Status:      models.LinkStatusActive,
			Config: models.LinkConfig{
				Bandwidth:  1000,
				Latency:    1,
				Jitter:     0,
				PacketLoss: 0.0,
				Protocol:   "ethernet",
				Encryption: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			NetworkID:   starNetwork.ID,
			SourceNodeID: starNodes[0].ID, // Hub
			TargetNodeID: starNodes[2].ID, // Node 2
			Status:      models.LinkStatusActive,
			Config: models.LinkConfig{
				Bandwidth:  1000,
				Latency:    1,
				Jitter:     0,
				PacketLoss: 0.0,
				Protocol:   "ethernet",
				Encryption: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			NetworkID:   starNetwork.ID,
			SourceNodeID: starNodes[0].ID, // Hub
			TargetNodeID: starNodes[3].ID, // Node 3
			Status:      models.LinkStatusActive,
			Config: models.LinkConfig{
				Bandwidth:  1000,
				Latency:    1,
				Jitter:     0,
				PacketLoss: 0.0,
				Protocol:   "ethernet",
				Encryption: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			NetworkID:   starNetwork.ID,
			SourceNodeID: starNodes[0].ID, // Hub
			TargetNodeID: starNodes[4].ID, // Node 4
			Status:      models.LinkStatusActive,
			Config: models.LinkConfig{
				Bandwidth:  1000,
				Latency:    1,
				Jitter:     0,
				PacketLoss: 0.0,
				Protocol:   "ethernet",
				Encryption: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, link := range starLinks {
		if err := db.Create(link).Error; err != nil {
			return fmt.Errorf("failed to create star link: %w", err)
		}
		logger.Info("Created star link",
			zap.String("source", link.SourceNodeID.String()[:8]),
			zap.String("target", link.TargetNodeID.String()[:8]),
		)
	}

	// Step 7: Create Links for Mesh Network (full mesh connectivity)
	meshLinks := []*models.Link{
		// Node 1 to Node 2
		{
			NetworkID:   meshNetwork.ID,
			SourceNodeID: meshNodes[0].ID,
			TargetNodeID: meshNodes[1].ID,
			Status:      models.LinkStatusActive,
			Config: models.LinkConfig{
				Bandwidth:  1000,
				Latency:    1,
				Jitter:     0,
				PacketLoss: 0.0,
				Protocol:   "ethernet",
				Encryption: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		// Node 1 to Node 3
		{
			NetworkID:   meshNetwork.ID,
			SourceNodeID: meshNodes[0].ID,
			TargetNodeID: meshNodes[2].ID,
			Status:      models.LinkStatusActive,
			Config: models.LinkConfig{
				Bandwidth:  1000,
				Latency:    1,
				Jitter:     0,
				PacketLoss: 0.0,
				Protocol:   "ethernet",
				Encryption: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		// Node 1 to Node 4
		{
			NetworkID:   meshNetwork.ID,
			SourceNodeID: meshNodes[0].ID,
			TargetNodeID: meshNodes[3].ID,
			Status:      models.LinkStatusActive,
			Config: models.LinkConfig{
				Bandwidth:  1000,
				Latency:    1,
				Jitter:     0,
				PacketLoss: 0.0,
				Protocol:   "ethernet",
				Encryption: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		// Node 1 to Node 5
		{
			NetworkID:   meshNetwork.ID,
			SourceNodeID: meshNodes[0].ID,
			TargetNodeID: meshNodes[4].ID,
			Status:      models.LinkStatusActive,
			Config: models.LinkConfig{
				Bandwidth:  1000,
				Latency:    1,
				Jitter:     0,
				PacketLoss: 0.0,
				Protocol:   "ethernet",
				Encryption: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		// Node 2 to Node 3
		{
			NetworkID:   meshNetwork.ID,
			SourceNodeID: meshNodes[1].ID,
			TargetNodeID: meshNodes[2].ID,
			Status:      models.LinkStatusActive,
			Config: models.LinkConfig{
				Bandwidth:  1000,
				Latency:    1,
				Jitter:     0,
				PacketLoss: 0.0,
				Protocol:   "ethernet",
				Encryption: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		// Node 2 to Node 4
		{
			NetworkID:   meshNetwork.ID,
			SourceNodeID: meshNodes[1].ID,
			TargetNodeID: meshNodes[3].ID,
			Status:      models.LinkStatusActive,
			Config: models.LinkConfig{
				Bandwidth:  1000,
				Latency:    1,
				Jitter:     0,
				PacketLoss: 0.0,
				Protocol:   "ethernet",
				Encryption: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		// Node 3 to Node 4
		{
			NetworkID:   meshNetwork.ID,
			SourceNodeID: meshNodes[2].ID,
			TargetNodeID: meshNodes[3].ID,
			Status:      models.LinkStatusActive,
			Config: models.LinkConfig{
				Bandwidth:  1000,
				Latency:    1,
				Jitter:     0,
				PacketLoss: 0.0,
				Protocol:   "ethernet",
				Encryption: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, link := range meshLinks {
		if err := db.Create(link).Error; err != nil {
			return fmt.Errorf("failed to create mesh link: %w", err)
		}
		logger.Info("Created mesh link",
			zap.String("source", link.SourceNodeID.String()[:8]),
			zap.String("target", link.TargetNodeID.String()[:8]),
		)
	}

	logger.Info("Database seeding completed successfully",
		zap.Int("networks", 2),
		zap.Int("nodes", len(starNodes)+len(meshNodes)),
		zap.Int("links", len(starLinks)+len(meshLinks)),
	)

	return nil
}

// ResetDatabase resets the database by TRUNCATING all core tables with CASCADE
func ResetDatabase(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Resetting database (TRUNCATING all core tables with CASCADE)...")

	// TRUNCATE core tables in order, using CASCADE to handle foreign keys
	// Order: links -> nodes -> policies -> networks (children first, then parents)
	tables := []string{"links", "nodes", "policies", "networks"}
	
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE;", table)).Error; err != nil {
			logger.Error(fmt.Sprintf("Failed to truncate %s", table), zap.Error(err))
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
		logger.Info(fmt.Sprintf("Truncated table: %s", table))
	}

	logger.Info("Database reset completed (all core tables truncated)")
	return nil
}

// GetRowCounts returns the row counts for networks, nodes, and links
func GetRowCounts(db *gorm.DB, logger *zap.Logger) (networks, nodes, links int64, err error) {
	if err := db.Model(&models.Network{}).Count(&networks).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("failed to count networks: %w", err)
	}
	
	if err := db.Model(&models.Node{}).Count(&nodes).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("failed to count nodes: %w", err)
	}
	
	if err := db.Model(&models.Link{}).Count(&links).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("failed to count links: %w", err)
	}
	
	return networks, nodes, links, nil
}

