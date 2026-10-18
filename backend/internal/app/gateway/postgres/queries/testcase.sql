-- name: CreateTestCase :exec
INSERT INTO test_cases(id,demand_id,cycle,title,steps,expected,actual,result,author,created_at,author_id) VALUES(sqlc.arg(id),sqlc.arg(demand_id),sqlc.arg(cycle),sqlc.arg(title),sqlc.arg(steps),sqlc.arg(expected),sqlc.arg(actual),sqlc.arg(result),sqlc.arg(author),sqlc.arg(created_at),NULLIF(sqlc.arg(author_id)::text,'')::uuid);

-- name: GetQACycleStats :one
SELECT count(*) FILTER (WHERE result='passed')::int AS passed,count(*) FILTER (WHERE result='failed')::int AS failed FROM test_cases WHERE demand_id=sqlc.arg(demand_id) AND cycle=sqlc.arg(cycle);
