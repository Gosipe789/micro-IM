-- name: CreateTransferAlternative :exec
INSERT INTO transfer_alternative
(amount,
 from_address,
 to_address,
 block,
 transaction_id,
 time,
 updated_at,
 created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetTransferAlternative :one
SELECT *
FROM transfer_alternative
WHERE transaction_id = ? LIMIT 1;


-- name: DeleteTransferAlternative :exec
DELETE
FROM transfer_alternative
WHERE transaction_id = ?;

-- name: DeleteTransferAlternativeByTime :exec
DELETE
FROM transfer_alternative
WHERE time < ?;

-- name: TransferAlternative :one
SELECT * FROM transfer_alternative WHERE from_address = ? AND to_address = ? LIMIT 1;
