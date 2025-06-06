package sql

var (
	RegisterUser = `
	INSERT INTO users (name, email, password_hash)
	VALUES ($1, $2, $3)
	RETURNING id;
`
	GetUser = `
	SELECT id, name, email FROM users WHERE id = $1;
`
	UpdateUser = `
	UPDATE users
	SET name = $2, email = $3
	WHERE id = $1
	RETURNING id;
`
	DeleteUser = `
	DELETE FROM users WHERE id = $1
	RETURNING id;
`
)

var (
	CreateSession = `
	INSERT INTO user_sessions (session_token, user_id, expires_at)
	VALUES ($1, $2, $3)
	RETURNING session_token;
`
	GetUserByEmail = `
	SELECT id, password_hash FROM users WHERE email = $1;
`
	DeleteSession = `
	DELETE FROM user_sessions WHERE session_token = $1;
`
)
