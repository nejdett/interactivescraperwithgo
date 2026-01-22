-- CTI Dashboard Database Schema Initialization

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'analyst',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Categories table
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    default_criticality INT CHECK (default_criticality BETWEEN 1 AND 10),
    color VARCHAR(7),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Content items table
CREATE TABLE content_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    source_name VARCHAR(255) NOT NULL,
    source_url TEXT NOT NULL,
    content TEXT NOT NULL,
    published_at TIMESTAMP,
    criticality_score INT CHECK (criticality_score BETWEEN 1 AND 10),
    collected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Content-Category mapping (many-to-many)
CREATE TABLE content_categories (
    content_id UUID REFERENCES content_items(id) ON DELETE CASCADE,
    category_id UUID REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (content_id, category_id)
);

-- Indexes for performance
CREATE INDEX idx_content_items_published_at ON content_items(published_at DESC);
CREATE INDEX idx_content_items_criticality ON content_items(criticality_score DESC);
CREATE INDEX idx_content_items_source ON content_items(source_name);

-- Seed data: Default categories
INSERT INTO categories (name, description, default_criticality, color) VALUES
    ('ransomware', 'Ransomware campaigns and related threats', 10, '#FF0000'),
    ('data-leak', 'Data breaches and leaked information', 9, '#FF6B00'),
    ('vulnerability', 'Software vulnerabilities and exploits', 7, '#FFB800'),
    ('phishing', 'Phishing campaigns and social engineering', 6, '#FFF000'),
    ('malware', 'Malware samples and analysis', 8, '#FF4500'),
    ('exploit', 'Exploit kits and techniques', 7, '#FF8C00');

-- Seed data: Default admin user
-- Password: admin123 (hashed with bcrypt cost 12)
INSERT INTO users (username, password_hash, role) VALUES
    ('admin', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIr.8VfGHO', 'admin')
ON CONFLICT (username) DO NOTHING;

-- Add comment for password reference
COMMENT ON TABLE users IS 'Default admin credentials: username=admin, password=admin123';
