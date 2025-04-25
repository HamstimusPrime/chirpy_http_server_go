-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2

)
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users
RETURNING *;


-- name: UpdateUserPassword :exec
UPDATE users
SET hashed_password = $1
WHERE id = $2;



-- name: UpdateUserEmail :one
UPDATE users
SET email = $1, updated_at = $2
WHERE ID = $3
RETURNING email, updated_at, created_at;;