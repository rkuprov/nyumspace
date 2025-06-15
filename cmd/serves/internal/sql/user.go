package sql

var (
	RegisterUser = `
	INSERT INTO users (id, name, email, password_hash)
	VALUES ($1, $2, $3, $4)
	RETURNING id;
`
	GetUser = `
	SELECT id, name, email FROM users WHERE id = $1;
`
	GetAllUsers = `
	SELECT id, name, email FROM users;
`
	UpdateUser = `
	UPDATE users
	SET name = $2, 
	email = $3,
	password_hash = coalesce($4, password_hash),
	WHERE id = $1
	RETURNING id;
`
	DeleteUser = `
	DELETE FROM users WHERE id = $1
`
)

var (
	GetUserByEmail = `
	SELECT id, password_hash FROM users WHERE email = $1;
`
)
