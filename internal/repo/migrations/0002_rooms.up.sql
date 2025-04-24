-- Таблица комнат
CREATE TABLE IF NOT EXISTS rooms (
    id INT PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    creator_id UUID NOT NULL
);

-- Таблица участников комнат
CREATE TABLE IF NOT EXISTS room_members (
    room_id INT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INT NOT NULL,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_admin BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (room_id, user_id)
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_room_members_user_id ON room_members(user_id);
CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages(room_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);