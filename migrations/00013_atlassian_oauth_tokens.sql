-- +goose Up
CREATE TABLE atlassian_oauth_tokens (
  user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  cloud_id TEXT NOT NULL DEFAULT '',
  site_url TEXT NOT NULL DEFAULT '',
  account_email TEXT NOT NULL DEFAULT '',
  access_token_encrypted BLOB NOT NULL,
  refresh_token_encrypted BLOB NOT NULL,
  expires_at TEXT NOT NULL,
  scopes TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS atlassian_oauth_tokens;
