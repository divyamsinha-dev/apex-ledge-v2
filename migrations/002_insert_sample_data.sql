-- Insert sample accounts for testing
-- Note: In production, accounts would be created through a proper API
INSERT INTO accounts (id, balance_cents, currency) VALUES
    ('account-001', 100000, 'USD'),
    ('account-002', 50000, 'USD'),
    ('account-003', 75000, 'USD')
ON CONFLICT (id) DO NOTHING;
