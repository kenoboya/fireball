CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY NOT NULL,
    username VARCHAR(70) NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(255) NULL
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS verification_log (
    verification_log_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    recipient VARCHAR(75) NOT NULL,
    code VARCHAR(10) NOT NULL,
    method ENUM('EMAIL', 'SMS') NOT NULL
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS chats (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    chat_external_id BIGINT NOT NULL UNIQUE,
    creator_id VARCHAR(255) NOT NULL,
    name VARCHAR(50) NOT NULL,
    encrypted BOOLEAN NOT NULL,
    avatar_url TEXT,
    updated_at TIMESTAMP NOT NULL,
    action ENUM(
        'private chat was created',
        'group chat was created',
        'chat was deleted',
        'added to chat by',
        'left chat',
        'changed the chat name to',
        'was kicked by'
    ) NULL,
    FOREIGN KEY (creator_id) REFERENCES users(user_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    message_external_id BIGINT NOT NULL UNIQUE,
    sender_id VARCHAR(255) NOT NULL,
    type ENUM('text', 'media', 'file', 'location', 'mixed') NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    action ENUM('send', 'edited', 'blurred', 'password', 'replied', 'pinned') DEFAULT 'send',
    FOREIGN KEY (sender_id) REFERENCES users(user_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS chat_messages (
    chat_id BIGINT NOT NULL,
    message_id BIGINT NOT NULL,
    PRIMARY KEY (chat_id, message_id),
    FOREIGN KEY (chat_id) REFERENCES chats(id),
    FOREIGN KEY (message_id) REFERENCES messages(id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS message_notifications (
    message_id BIGINT NOT NULL,
    recipient_id VARCHAR(255) NOT NULL,
    is_read BOOLEAN DEFAULT false,
    PRIMARY KEY (message_id, recipient_id),
    FOREIGN KEY (message_id) REFERENCES messages(id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS chat_notifications (
    chat_id BIGINT NOT NULL,
    recipient_id VARCHAR(255) NOT NULL,
    is_read BOOLEAN DEFAULT false,
    PRIMARY KEY (chat_id, recipient_id),
    FOREIGN KEY (chat_id) REFERENCES chats(id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS user_notifications (
    user_id VARCHAR(255) NOT NULL,
    chat_id BIGINT NOT NULL,
    mute BOOLEAN DEFAULT true,
    term TIMESTAMP NULL,
    PRIMARY KEY (user_id, chat_id)
) ENGINE=InnoDB;