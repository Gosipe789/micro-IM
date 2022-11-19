-- name: CreateTransfer :exec
INSERT INTO transfer
(amount,
 from_address,
 to_address,
 block,
 transaction_id,
 time,
 updated_at,
 created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetTransfer :one
SELECT *
FROM transfer
WHERE transaction_id = ? LIMIT 1;


-- name: DeleteTransfer :exec
DELETE
FROM transfer
WHERE transaction_id = ?;

-- name: Transfer :one
SELECT * FROM transfer WHERE from_address = ? AND to_address = ? LIMIT 1;

-- name: IsExistTransfer :exec
SELECT EXISTS(SELECT 1 FROM transfer WHERE to_address = ? LIMIT 1);
