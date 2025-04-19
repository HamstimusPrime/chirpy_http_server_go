-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1

)
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users
RETURNING *;


-- name: GetAllUsers :many

SELECT * FROM users
ORDER BY created_at ASC;
