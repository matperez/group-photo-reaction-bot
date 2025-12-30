-- Откат миграции 001: удаление схемы БД

-- Удаляем триггеры
DROP TRIGGER IF EXISTS update_admin_users_timestamp ON admin_users;
DROP TRIGGER IF EXISTS update_chats_timestamp ON chats;
DROP TRIGGER IF EXISTS update_sessions_timestamp ON sessions;

-- Удаляем функцию
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаляем индексы
DROP INDEX IF EXISTS idx_admin_sessions_expires;
DROP INDEX IF EXISTS idx_admin_sessions_user_id;
DROP INDEX IF EXISTS idx_votes_user_id;
DROP INDEX IF EXISTS idx_votes_session_id;
DROP INDEX IF EXISTS idx_sessions_message;
DROP INDEX IF EXISTS idx_sessions_is_active;
DROP INDEX IF EXISTS idx_sessions_chat_id;

-- Удаляем таблицы (в правильном порядке из-за foreign keys)
DROP TABLE IF EXISTS admin_sessions;
DROP TABLE IF EXISTS admin_users;
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS chats;

