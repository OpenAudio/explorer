-- name: GetChainStats :one
select * from chain_stats where id = 1;

-- name: GetTransactionWindowCounts :one
select * from get_transaction_window_counts();

-- ========================================
-- blocks queries
-- ========================================

-- name: GetBlock :one
select * from blocks where height = $1;

-- name: GetBlockByHash :one
select * from blocks where hash = $1;

-- name: GetLatestBlock :one
select * from blocks order by height desc limit 1;

-- name: ListBlocks :many
select * from blocks
order by height desc
limit $1 offset $2;

-- name: CountBlocks :one
select count(*) from blocks;

-- name: GetRecentBlocks :many
select * from blocks
order by height desc
limit $1;

-- ========================================
-- transactions queries
-- ========================================

-- name: GetTransaction :one
select * from transactions where tx_hash = $1;

-- name: ListTransactions :many
select * from transactions
order by created_at desc
limit $1 offset $2;

-- name: CountTransactions :one
select count(*) from transactions;

-- name: GetRecentTransactions :many
select * from transactions
order by created_at desc
limit $1;

-- name: GetTransactionsByBlock :many
select * from transactions
where block_height = $1
order by tx_index asc;

-- name: CountTransactionsByBlock :one
select count(*) from transactions
where block_height = $1;

-- name: GetTransactionsByType :many
select * from transactions
where tx_type = $1
order by created_at desc
limit $2 offset $3;

-- name: GetTransactionsBySender :many
select * from transactions
where sender = $1
order by created_at desc
limit $2 offset $3;

-- name: GetTransactionsByProposer :many
select * from transactions
where proposer = $1
order by created_at desc
limit $2 offset $3;

-- name: GetTransactionTypeStats :many
select * from transaction_type_stats
order by total_count desc;

-- name: GetTransactionTypeStat :one
select * from transaction_type_stats
where tx_type = $1;

-- ========================================
-- validators queries
-- ========================================

-- name: GetValidator :one
select * from validators where address = $1;

-- name: GetValidatorByCometAddress :one
select * from validators where comet_address = $1;

-- name: ListValidators :many
select * from validators
order by voting_power desc
limit $1 offset $2;

-- name: ListValidatorsByStatus :many
select * from validators
where status = $1
order by voting_power desc
limit $2 offset $3;

-- name: ListActiveValidators :many
select * from validators
where status = 'active'
order by voting_power desc
limit $1 offset $2;

-- name: CountValidators :one
select count(*) from validators;

-- name: CountValidatorsByStatus :one
select count(*) from validators where status = $1;

-- name: FilterValidatorsByEndpoint :many
select * from validators
where endpoint like '%' || $1 || '%'
  and ($2::text is null or status = $2)
order by voting_power desc
limit $3 offset $4;

-- name: GetValidatorStats :one
select * from validator_stats where validator_address = $1;

-- ========================================
-- validator events queries
-- ========================================

-- name: GetValidatorEvents :many
select * from validator_events
where validator_address = $1
order by created_at desc
limit $2;

-- name: GetValidatorEventsByType :many
select * from validator_events
where validator_address = $1 and event_type = $2
order by created_at desc
limit $3;

-- ========================================
-- sla rollups queries
-- ========================================

-- name: GetSLARollup :one
select * from sla_rollups where id = $1;

-- name: GetLatestSLARollup :one
select * from sla_rollups
order by id desc
limit 1;

-- name: ListSLARollups :many
select * from sla_rollups
order by id desc
limit $1 offset $2;

-- name: CountSLARollups :one
select count(*) from sla_rollups;

-- name: GetRecentSLARollups :many
select * from sla_rollups
order by id desc
limit $1;

-- ========================================
-- sla node reports queries
-- ========================================

-- name: GetSLANodeReport :one
select * from sla_node_reports
where sla_rollup_id = $1 and validator_address = $2;

-- name: ListSLANodeReportsByRollup :many
select * from sla_node_reports
where sla_rollup_id = $1
order by num_blocks_proposed desc;

-- name: ListSLANodeReportsByValidator :many
select * from sla_node_reports
where validator_address = $1
order by sla_rollup_id desc
limit $2;

-- name: GetRecentSLANodeReportsForValidator :many
select * from sla_node_reports
where validator_address = $1
order by sla_rollup_id desc
limit $2;

-- name: GetValidatorUptimeMap :many
-- Get last N SLA rollups for multiple validators
select
    validator_address,
    sla_rollup_id,
    num_blocks_proposed,
    challenges_received,
    challenges_failed,
    created_at
from sla_node_reports
where validator_address = any($1::text[])
  and sla_rollup_id in (
    select id from sla_rollups order by id desc limit $2
  )
order by validator_address, sla_rollup_id desc;

-- ========================================
-- accounts queries
-- ========================================

-- name: GetAccount :one
select * from accounts where address = $1;

-- name: GetAccountByUserID :one
select * from accounts where user_id = $1;

-- name: ListAccounts :many
select * from accounts
limit $1 offset $2;

-- ========================================
-- plays queries
-- ========================================

-- name: GetPlay :one
select * from plays where id = $1;

-- name: ListPlays :many
select * from plays
order by played_at desc
limit $1 offset $2;

-- name: GetPlaysByUser :many
select * from plays
where user_id = $1
order by played_at desc
limit $2 offset $3;

-- name: GetPlaysByTrack :many
select * from plays
where track_id = $1
order by played_at desc
limit $2 offset $3;

-- name: GetPlaysByCountry :many
select * from plays
where country = $1
order by played_at desc
limit $2 offset $3;

-- name: GetRecentPlays :many
select * from plays
order by played_at desc
limit $1;

-- ========================================
-- manage entities queries
-- ========================================

-- name: GetManageEntity :one
select * from manage_entities where id = $1;

-- name: ListManageEntities :many
select * from manage_entities
order by created_at desc
limit $1 offset $2;

-- name: GetManageEntitiesByAddress :many
select * from manage_entities
where address = $1
order by created_at desc
limit $2 offset $3;

-- name: GetManageEntitiesByType :many
select * from manage_entities
where entity_type = $1
order by created_at desc
limit $2 offset $3;

-- name: GetManageEntitiesByEntity :many
select * from manage_entities
where entity_type = $1 and entity_id = $2
order by created_at desc
limit $3;

-- name: GetManageEntitiesByAddressAndType :many
select * from manage_entities
where address = $1 and entity_type = $2
order by created_at desc
limit $3 offset $4;

-- ========================================
-- dashboard queries
-- ========================================

-- name: GetDashboardStats :one
-- Consolidated query for dashboard statistics
select
    cs.latest_block_height,
    cs.latest_block_hash,
    cs.latest_block_time,
    cs.total_transactions,
    cs.total_validators,
    cs.active_validators,
    cs.avg_block_time_seconds,
    cs.transactions_24h,
    cs.transactions_7d,
    cs.transactions_30d,
    (select count(*) from validators where status = 'active') as current_active_validators
from chain_stats cs
where cs.id = 1;

-- name: GetTransactionBreakdown :many
-- Get transaction type breakdown for dashboard
select
    tx_type,
    total_count,
    latest_transaction
from transaction_type_stats
order by total_count desc;

-- name: GetSLAPerformanceData :many
-- Get SLA performance data for charting (last N rollups)
select
    sr.id as rollup_id,
    sr.block_start,
    sr.block_end,
    sr.validator_count,
    sr.block_quota,
    sr.created_at,
    count(distinct snr.validator_address) filter (
        where snr.num_blocks_proposed >= (sr.block_quota * 0.8)
          and (snr.challenges_received = 0 or
               (1.0 - (snr.challenges_failed::float / snr.challenges_received)) >= 0.8)
    ) as healthy_validators
from sla_rollups sr
left join sla_node_reports snr on snr.sla_rollup_id = sr.id
group by sr.id, sr.block_start, sr.block_end, sr.validator_count, sr.block_quota, sr.created_at
order by sr.id desc
limit $1;

-- ========================================
-- validators uptime page queries
-- ========================================

-- name: GetValidatorsWithRecentUptime :many
-- Get all validators with their recent SLA rollup performance
select
    v.address,
    v.comet_address,
    v.endpoint,
    v.node_type,
    v.spid,
    v.voting_power,
    v.status,
    v.registered_at,
    v.created_at,
    v.updated_at
from validators v
where ($1::text is null or v.status = $1)
order by v.voting_power desc
limit $2 offset $3;
