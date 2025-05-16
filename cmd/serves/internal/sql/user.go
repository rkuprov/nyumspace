package sql

var (
	RegisterUser = `
	INSERT INTO users (name, email, password_hash)
	VALUES ($1, $2, $3)
	RETURNING id;
`
)
