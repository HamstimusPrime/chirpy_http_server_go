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
