-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
    VALUES(
        $1,
        NOW(),
        NOW(),
        $2,
        $3
    )
    RETURNING id, created_at,updated_at, body, user_id;

-- name: GetAllChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;



-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1
LIMIT 1;



-- name: GetChirpByUserID :one
SELECT * FROM chirps 
WHERE user_id = $1 AND id = $2
LIMIT 1;


-- name: DeleteChirpByID :exec
DELETE FROM chirps
WHERE id = $1;