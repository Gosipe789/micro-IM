-- name: GetConfig :one
SELECT *
FROM config LIMIT 1;

-- name: UpdateConfig :exec
UPDATE config
SET start_block      = ?,
    url              = ?,
    status           = ?,
    first_amount     = ?,
    second_amount    = ?,
    time_limit       = ?,
    amount_condition = ?;
