package sql

var (
	CreateSession = `
	INSERT INTO user_sessions (session_token, user_id, expires_at)
VALUES ($1, $2, $3)
RETURNING session_token;`

	DeleteSession = `
	DELETE FROM user_sessions WHERE session_token = $1;
`
	GetSession = `
    	SELECT user_id, expires_at FROM user_sessions WHERE session_token = $1;
`
)
