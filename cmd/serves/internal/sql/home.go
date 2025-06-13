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
	//UpdateHomeSQL = `
	//	UPDATE homes
	//	SET name = COALESCE($2, name),
	//  		owner_id = COALESCE($3, owner_id),
	//  		street_address_1 = COALESCE($4, street_address_1),
	//  		street_address_2 = COALESCE($5, street_address_2),
	//  		city = COALESCE($6, city),
	//  		state = COALESCE($7, state),
	//  		zip_code = COALESCE($8, zip_code),
	//  		country = COALESCE($9, country),
	//  		description = COALESCE($10, description),
	//  		tags = COALESCE($11, tags),
	//  		image = COALESCE($12, image),
	//  		updated_at = $13
	//	WHERE id = $1
	//	RETURNING id;
	//`
	DeleteHomeSQL = `
		DELETE FROM homes WHERE id = $1
		RETURNING id;
	`
	GetAllHomesForUserSQL = `
		SELECT id, owner_id, name, street_address_1, street_address_2, city, state, zip_code, country, description, tags, image, created_at, updated_at
		FROM homes WHERE owner_id = $1;
	`
)
