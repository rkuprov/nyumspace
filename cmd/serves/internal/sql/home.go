package sql

const (
	AddHomeSQL = `
		INSERT INTO homes (id, owner_id, name, street_address_1, street_address_2, city, state, zip_code, country, description, tags, image)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id;
	`
	GetHomeSQL = `
		SELECT id, owner_id, name, street_address_1, street_address_2, city, state, zip_code, country, description, tags, image, created_at, updated_at
		FROM homes WHERE id = $1;
	`
	GetAllHomesSQL = `
		SELECT id, owner_id, name, street_address_1, street_address_2, city, state, zip_code, country, description, tags, image, created_at, updated_at
		FROM homes;
	`
	UpdateHomeSQL = `
		UPDATE homes
		SET name = COALESCE($2, ''),
	 		description = COALESCE($3, ''),
	 		street_address_1 = COALESCE($4, ''),
	 		street_address_2 = COALESCE($5, ''),
	 		city = COALESCE($6, ''),
	 		state = COALESCE($7, ''),
	 		zip_code = COALESCE($8, ''),
	 		country = COALESCE($9, ''),
	 		image = COALESCE($10, ''),
	 		tags = $11,
	 		updated_at = $12
		WHERE id = $1
		RETURNING id;
	`
	DeleteHomeSQL = `
		DELETE FROM homes WHERE id = $1
		RETURNING id;
	`
	GetAllHomesForUserSQL = `
		SELECT id, owner_id, name, street_address_1, street_address_2, city, state, zip_code, country, description, tags, image, created_at, updated_at
		FROM homes WHERE owner_id = $1;
	`
)
