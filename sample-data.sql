-- Sample data for NetOrchestrator
-- Add more sample networks, nodes, and other entities

-- Insert additional users (password is hashed version of user123)
INSERT INTO users (username, email, password_hash, role) 
VALUES 
  ('user', 'user@netorchestrator.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'user'),
  ('demo', 'demo@netorchestrator.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'user')
ON CONFLICT (username) DO NOTHING;

-- Get the admin user ID for inserting sample networks
DO $$
DECLARE
    admin_user_id UUID;
    user_user_id UUID;
    network1_id UUID;
    network2_id UUID;
    network3_id UUID;
    node1_id UUID;
    node2_id UUID;
    node3_id UUID;
BEGIN
    -- Get user IDs
    SELECT id INTO admin_user_id FROM users WHERE username = 'admin';
    SELECT id INTO user_user_id FROM users WHERE username = 'user';
    
    -- Insert sample networks
    INSERT INTO networks (id, name, description, status, network_type, subnet, gateway, user_id, metadata)
    VALUES 
      (gen_random_uuid(), 'Production Network', 'Main production environment with high availability setup', 'active', 'enterprise', '10.0.0.0/16', '10.0.0.1', admin_user_id, '{"environment": "production", "region": "us-east-1"}'),
      (gen_random_uuid(), 'Development Network', 'Development and testing environment', 'active', 'development', '172.16.0.0/24', '172.16.0.1', admin_user_id, '{"environment": "development", "purpose": "testing"}'),
      (gen_random_uuid(), 'User Demo Network', 'User created demo network for testing features', 'provisioning', 'demo', '192.168.100.0/24', '192.168.100.1', user_user_id, '{"demo": true, "created_by": "user"}'
    ) RETURNING id INTO network1_id;
    
    -- Get network IDs for adding nodes and links
    SELECT id INTO network1_id FROM networks WHERE name = 'Production Network';
    SELECT id INTO network2_id FROM networks WHERE name = 'Development Network';
    SELECT id INTO network3_id FROM networks WHERE name = 'User Demo Network';
    
    -- Insert sample nodes for Production Network
    INSERT INTO nodes (id, network_id, name, node_type, ip_address, status, cpu_cores, memory_mb, metadata)
    VALUES 
      (gen_random_uuid(), network1_id, 'web-server-01', 'container', '10.0.1.10', 'active', 4, 8192, '{"role": "web-server", "tier": "frontend"}'),
      (gen_random_uuid(), network1_id, 'db-server-01', 'container', '10.0.2.10', 'active', 8, 16384, '{"role": "database", "tier": "backend"}'),
      (gen_random_uuid(), network1_id, 'load-balancer', 'container', '10.0.1.1', 'active', 2, 4096, '{"role": "load-balancer", "tier": "edge"}'
    ) RETURNING id INTO node1_id;
    
    -- Insert sample nodes for Development Network
    INSERT INTO nodes (network_id, name, node_type, ip_address, status, cpu_cores, memory_mb, metadata)
    VALUES 
      (network2_id, 'dev-web-01', 'container', '172.16.0.10', 'active', 2, 4096, '{"role": "web-server", "environment": "development"}'),
      (network2_id, 'dev-db-01', 'container', '172.16.0.11', 'active', 4, 8192, '{"role": "database", "environment": "development"}');
    
    -- Insert sample nodes for User Demo Network  
    INSERT INTO nodes (network_id, name, node_type, ip_address, status, cpu_cores, memory_mb, metadata)
    VALUES 
      (network3_id, 'demo-app', 'container', '192.168.100.10', 'provisioning', 1, 2048, '{"role": "demo-application", "demo": true}');
    
    -- Insert some sample metrics
    INSERT INTO metrics (entity_type, entity_id, metric_name, metric_value, unit, metadata)
    VALUES 
      ('network', network1_id, 'total_bandwidth_mbps', 1000.0, 'mbps', '{"measurement_type": "peak"}'),
      ('network', network1_id, 'active_connections', 150, 'count', '{"measurement_type": "current"}'),
      ('network', network2_id, 'total_bandwidth_mbps', 500.0, 'mbps', '{"measurement_type": "peak"}'),
      ('network', network2_id, 'active_connections', 25, 'count', '{"measurement_type": "current"}');
    
    -- Insert sample events
    INSERT INTO events (event_type, entity_type, entity_id, user_id, message, severity, metadata)
    VALUES 
      ('network_created', 'network', network1_id, admin_user_id, 'Production Network created successfully', 'info', '{"action": "create"}'),
      ('network_created', 'network', network2_id, admin_user_id, 'Development Network created successfully', 'info', '{"action": "create"}'),
      ('node_provisioned', 'node', node1_id, admin_user_id, 'Web server node provisioned in production network', 'info', '{"action": "provision"}');

END $$;

-- Update the existing Default Network with better sample data
UPDATE networks 
SET description = 'Default starter network with basic configuration', 
    status = 'active',
    network_type = 'basic',
    metadata = '{"default": true, "starter_template": true}'
WHERE name = 'Default Network';
