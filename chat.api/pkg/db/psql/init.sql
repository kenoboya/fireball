DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS chats;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS media;
DROP TABLE IF EXISTS pinned_chats;
DROP TABLE IF EXISTS pinned_messages;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS messages_files;
DROP TABLE IF EXISTS messages_media;
DROP TABLE IF EXISTS messages_locations;
DROP TABLE IF EXISTS message_audit_log;
DROP TABLE IF EXISTS chats_participants;
DROP TABLE IF EXISTS chat_roles;
DROP TABLE IF EXISTS chat_history;
DROP TABLE IF EXISTS chat_blocked_users;
DROP TABLE IF EXISTS chat_messages;

DROP TYPE IF EXISTS message_status;
DROP TYPE IF EXISTS message_type;
DROP TYPE IF EXISTS message_action;
DROP TYPE IF EXISTS chat_role;
DROP TYPE IF EXISTS chat_action;
DROP TYPE IF EXISTS chat_type;
DROP TYPE IF EXISTS media_type;
DROP TYPE IF EXISTS file_type;

CREATE TYPE message_status AS ENUM ('sent', 'delivered', 'read');
CREATE TYPE message_type AS ENUM ('text', 'media', 'file', 'location', 'mixed');
CREATE TYPE message_action AS ENUM ('edited', 'blurred', 'deleted', 'password', 'replied', 'pinned');
CREATE TYPE chat_role AS ENUM ('user', 'admin');
CREATE TYPE chat_action AS ENUM ('chat was created', 'chat was deleted', 'added to chat by', 'left chat', 'changed the chat name to', 'was kicked by');

CREATE TYPE chat_type AS ENUM (
    'private',   
    'group',     
    'channel'    
);

CREATE TYPE media_type AS ENUM(
    'image/jpeg',
    'image/png',
    'image/gif',
    'video/mp4',
    'video/webm',
    'image/webp'
);

CREATE TYPE file_type AS ENUM(
    'image/jpeg',
    'image/png',
    'image/gif',
    'image/webp',
    'video/mp4',
    'video/webm',
    'audio/mp3',
    'application/pdf',
    'application/zip',
    'text/plain'
);

CREATE TABLE locations (
    location_id BIGINT GENERATED ALWAYS AS IDENTITY,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_locations PRIMARY KEY(location_id)
);

CREATE TABLE media (
    media_id BIGINT GENERATED ALWAYS AS IDENTITY,
    url VARCHAR(255) NOT NULL,
    type media_type NOT NULL,
    size BIGINT,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_media PRIMARY KEY (media_id)
);

CREATE TABLE files (
    file_id BIGINT GENERATED ALWAYS AS IDENTITY,
    url VARCHAR(255) NOT NULL,
    type file_type NOT NULL,
    size BIGINT,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_files PRIMARY KEY (file_id)
);

CREATE TABLE messages (
    message_id BIGINT GENERATED ALWAYS AS IDENTITY,
    sender_id VARCHAR(255) NOT NULL,
    content TEXT,
    status message_status DEFAULT 'sent',
    type message_type DEFAULT 'text',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_messages PRIMARY KEY (message_id)
);

CREATE TABLE message_audit_log (
    message_id BIGINT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    action_type message_action,
    action_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_message_audit_log PRIMARY KEY(message_id, user_id),
    CONSTRAINT fk_message_audit_log_message_id FOREIGN KEY(message_id) REFERENCES messages(message_id) ON DELETE CASCADE
);

CREATE TABLE messages_files (
    message_id BIGINT NOT NULL,
    file_id BIGINT NOT NULL,
    CONSTRAINT pk_messages_files PRIMARY KEY(message_id, file_id),
    CONSTRAINT fk_messages_files_message_id FOREIGN KEY(message_id) REFERENCES messages(message_id) ON DELETE CASCADE,
    CONSTRAINT fk_messages_files_file_id FOREIGN KEY(file_id) REFERENCES files(file_id) ON DELETE CASCADE
);

CREATE TABLE messages_media (
    message_id BIGINT NOT NULL,
    media_id BIGINT NOT NULL,
    CONSTRAINT pk_messages_media PRIMARY KEY(message_id, media_id),
    CONSTRAINT fk_messages_media_message_id FOREIGN KEY(message_id) REFERENCES messages(message_id) ON DELETE CASCADE,
    CONSTRAINT fk_messages_media_image_id FOREIGN KEY(media_id) REFERENCES media(media_id) ON DELETE CASCADE
);

CREATE TABLE messages_locations (
    message_id BIGINT NOT NULL,
    location_id BIGINT NOT NULL,
    CONSTRAINT pk_messages_locations PRIMARY KEY(message_id, location_id),
    CONSTRAINT fk_messages_locations_message_id FOREIGN KEY(message_id) REFERENCES messages(message_id) ON DELETE CASCADE,
    CONSTRAINT fk_messages_locations_location_id FOREIGN KEY(location_id) REFERENCES locations(location_id) ON DELETE CASCADE
);

CREATE TABLE chats (
    chat_id BIGINT GENERATED ALWAYS AS IDENTITY,
    creator_id VARCHAR(255),
    name VARCHAR(50),
    description TEXT,
    type chat_type DEFAULT 'private',
    avatar_url TEXT,
    encrypted BOOLEAN DEFAULT false, -- E2EE
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_chats PRIMARY KEY(chat_id)
);

CREATE TABLE chats_participants (
    chat_id BIGINT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    CONSTRAINT pk_chats_participants PRIMARY KEY(chat_id, user_id),
    CONSTRAINT fk_chats_participants_chat_id FOREIGN KEY(chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE
);

CREATE TABLE pinned_chats (
    chat_id BIGINT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    priority SMALLINT CHECK (priority BETWEEN 1 AND 5),
    CONSTRAINT pk_pinned_chats PRIMARY KEY(chat_id, user_id),
    CONSTRAINT fk_pinned_chats_chat_id FOREIGN KEY(chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE
);

CREATE TABLE pinned_messages (
    chat_id BIGINT NOT NULL,
    message_id BIGINT NOT NULL,
    pinned_by_user_id VARCHAR(255) NOT NULL,
    priority SMALLINT CHECK (priority BETWEEN 1 AND 5),
    CONSTRAINT pk_pinned_messages PRIMARY KEY(chat_id, message_id),
    CONSTRAINT fk_pinned_messages_chat_id FOREIGN KEY(chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE,
    CONSTRAINT fk_pinned_messages_message_id FOREIGN KEY(message_id) REFERENCES messages(message_id) ON DELETE CASCADE
);

CREATE TABLE chat_roles (
    chat_id BIGINT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    granter_id VARCHAR(255) NOT NULL,
    nickname VARCHAR(50) NOT NULL,
    role chat_role DEFAULT 'user',
    CONSTRAINT pk_chat_roles PRIMARY KEY(chat_id, user_id),
    CONSTRAINT fk_chat_roles_chat_id FOREIGN KEY(chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE
);

CREATE TABLE chat_history (
    chat_id BIGINT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    action_type chat_action NOT NULL,
    action_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_chat_history PRIMARY KEY(chat_id, user_id),
    CONSTRAINT fk_chat_history_chat_id FOREIGN KEY(chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE
);

CREATE TABLE chat_blocked_users (
    chat_id BIGINT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    blocked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_chat_blocked_users PRIMARY KEY (chat_id, user_id),
    CONSTRAINT fk_chat_blocked_users_chat_id FOREIGN KEY(chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE
);

CREATE TABLE chat_messages (
    chat_id BIGINT NOT NULL,
    message_id BIGINT NOT NULL,
    CONSTRAINT pk_chat_messages PRIMARY KEY(chat_id, message_id),
    CONSTRAINT fk_chat_messages_chat_id FOREIGN KEY(chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE,
    CONSTRAINT fk_chat_messages_message_id FOREIGN KEY(message_id) REFERENCES messages(message_id) ON DELETE CASCADE
);

CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_chats_participants_user_id ON chats_participants(user_id);
CREATE INDEX idx_pinned_chats_user_id ON pinned_chats(user_id);
CREATE INDEX idx_chat_roles_user_id ON chat_roles(user_id);
CREATE INDEX idx_chat_history_user_id ON chat_history(user_id);
CREATE INDEX idx_message_audit_log_user_id ON message_audit_log(user_id);
CREATE INDEX idx_chat_blocked_users_user_id ON chat_blocked_users(user_id);
