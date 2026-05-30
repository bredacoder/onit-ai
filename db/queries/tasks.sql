-- name: ListTasksByUser :many
SELECT
    id,
    user_id,
    service_type,
    location,
    time_window,
    budget_cap,
    task_constraints,
    state,
    attributes,
    created_at
FROM tasks
WHERE user_id = $1
ORDER BY created_at;
