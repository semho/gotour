CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    nickname VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE chats (
    id UUID PRIMARY KEY,
    history_size INTEGER NOT NULL,
    ttl TIMESTAMP WITH TIME ZONE,
    read_only BOOLEAN NOT NULL DEFAULT FALSE,
    private BOOLEAN NOT NULL DEFAULT FALSE,
    owner_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE messages (
    id UUID PRIMARY KEY,
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    nickname VARCHAR(255) NOT NULL,
    text TEXT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE chat_access (
    chat_id UUID REFERENCES chats(id) ON DELETE CASCADE,
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    granted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (chat_id, session_id)
);

CREATE TABLE anon_nicknames (
    chat_id UUID REFERENCES chats(id) ON DELETE CASCADE,
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    nickname VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (chat_id, session_id)
);

CREATE TABLE anon_counts (
    chat_id UUID PRIMARY KEY REFERENCES chats(id) ON DELETE CASCADE,
    count INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_sessions_nickname ON sessions(nickname);
CREATE INDEX idx_chats_owner_id ON chats(owner_id);
CREATE INDEX idx_chats_ttl ON chats(ttl) WHERE ttl IS NOT NULL;
CREATE INDEX idx_messages_chat_id ON messages(chat_id);
CREATE INDEX idx_messages_session_id ON messages(session_id);
CREATE INDEX idx_messages_timestamp ON messages(timestamp);
CREATE INDEX idx_chat_access_session_id ON chat_access(session_id);
CREATE INDEX idx_anon_nicknames_session_id ON anon_nicknames(session_id);