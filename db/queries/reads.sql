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

-- name: GetBlocksByPage :many
select * from blocks
order by height desc
limit $1 offset $2;

-- name: GetBlockByHeight :one
select * from blocks where height = $1;

-- name: GetBlockTransactionCount :one
select count(*) from transactions where block_height = $1;

-- ========================================
-- transactions queries
-- ========================================

-- name: GetTransaction :one
select * from transactions where tx_hash = $1;

-- name: GetTransactionByHash :one
select * from transactions where tx_hash = $1;

-- name: ListTransactions :many
select * from transactions
order by created_at desc
limit $1 offset $2;

-- name: GetTransactionsByPage :many
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

-- name: GetTransactionsByAddress :many
-- Get transactions for an address with optional filters
select
    t.id,
    t.tx_hash,
    t.block_height,
    t.tx_index,
    t.tx_type,
    t.created_at,
    case
        when t.sender = lower($1) then 'sender'
        when t.proposer = lower($1) then 'proposer'
        else 'unknown'
    end as relation
from transactions t
where (lower(t.sender) = lower($1) or lower(t.proposer) = lower($1))
  and ($2 = '' or case
        when t.sender = lower($1) then 'sender'
        when t.proposer = lower($1) then 'proposer'
        else 'unknown'
    end = $2)
  and ($3::timestamp is null or t.created_at >= $3)
  and ($4::timestamp is null or t.created_at <= $4)
order by t.created_at desc
limit $5 offset $6;

-- name: GetTransactionCountByAddress :one
select count(*)
from transactions t
where (lower(t.sender) = lower($1) or lower(t.proposer) = lower($1))
  and ($2 = '' or case
        when t.sender = lower($1) then 'sender'
        when t.proposer = lower($1) then 'proposer'
        else 'unknown'
    end = $2)
  and ($3::timestamp is null or t.created_at >= $3)
  and ($4::timestamp is null or t.created_at <= $4);

-- name: GetRelationTypesByAddress :many
select distinct case
    when t.sender = lower($1) then 'sender'
    when t.proposer = lower($1) then 'proposer'
    else 'unknown'
end as relation
from transactions t
where lower(t.sender) = lower($1) or lower(t.proposer) = lower($1);

-- ========================================
-- validators queries
-- ========================================

-- name: GetValidator :one
select * from validators where address = $1;

-- name: GetValidatorByAddress :one
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

-- name: GetActiveValidators :many
select * from validators
where status = 'active'
order by voting_power desc
limit $1 offset $2;

-- name: GetActiveValidatorCount :one
select count(*) from validators where status = 'active';

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

-- name: GetValidatorRegistrationByTxHash :one
select * from validator_events
where tx_hash = $1 and event_type = 'registration'
limit 1;

-- name: GetValidatorDeregistrationByTxHash :one
select * from validator_events
where tx_hash = $1 and event_type = 'deregistration'
limit 1;

-- name: GetValidatorRegistrations :many
select
    ve.id,
    ve.validator_address,
    ve.block_height,
    ve.tx_hash,
    ve.created_at,
    v.address,
    v.endpoint,
    v.comet_address,
    v.node_type,
    v.spid,
    v.voting_power
from validator_events ve
join validators v on v.address = ve.validator_address
where ve.event_type = 'registration'
order by ve.created_at desc
limit $1 offset $2;

-- name: GetValidatorDeregistrations :many
select
    ve.id,
    ve.validator_address as comet_address,
    ve.block_height,
    ve.tx_hash,
    ve.created_at,
    v.address,
    v.endpoint,
    v.node_type,
    v.spid,
    v.voting_power
from validator_events ve
left join validators v on v.address = ve.validator_address
where ve.event_type = 'deregistration'
order by ve.created_at desc
limit $1 offset $2;

-- ========================================
-- sla rollups queries
-- ========================================

-- name: GetSLARollup :one
select * from sla_rollups where id = $1;

-- name: GetSlaRollupById :one
select * from sla_rollups where id = $1;

-- name: GetSlaRollupByTxHash :one
select * from sla_rollups where tx_hash = $1;

-- name: GetLatestSLARollup :one
select * from sla_rollups
order by id desc
limit 1;

-- name: ListSLARollups :many
select * from sla_rollups
order by id desc
limit $1 offset $2;

-- name: GetSlaRollupsWithPagination :many
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

-- name: GetSlaNodeReportsByAddress :many
select * from sla_node_reports
where validator_address = lower($1)
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

-- name: GetValidatorsForSlaRollup :many
select
    v.*,
    snr.num_blocks_proposed
from validators v
left join sla_node_reports snr on snr.validator_address = v.comet_address and snr.sla_rollup_id = $1
where v.status = 'active'
order by v.voting_power desc;

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

-- name: GetPlaysByTxHash :many
select * from plays
where tx_hash = $1
order by played_at desc;

-- ========================================
-- manage entities queries
-- ========================================

-- name: GetManageEntity :one
select * from manage_entities where id = $1;

-- name: GetManageEntityByTxHash :one
select * from manage_entities where tx_hash = $1;

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
-- storage proofs queries
-- ========================================

-- name: GetStorageProofByTxHash :one
select * from storage_proofs where tx_hash = $1;

-- name: GetStorageProofVerificationByTxHash :one
select * from storage_proof_verifications where tx_hash = $1;

-- ========================================
-- dashboard queries
-- ========================================

-- name: GetLatestIndexedBlock :one
select coalesce(max(height), 0) from blocks;

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

-- name: GetDashboardTransactionStats :one
select
    total_transactions,
    transactions_24h,
    transactions_24h as transactions_previous_24h, -- placeholder
    transactions_7d,
    transactions_30d
from chain_stats
where id = 1;

-- name: GetDashboardTransactionTypes :many
select
    tx_type,
    total_count as transaction_count
from transaction_type_stats
order by total_count desc;

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

-- name: GetHealthyValidatorCountsForRollups :many
select
    sr.id as rollup_id,
    count(distinct snr.validator_address) filter (
        where snr.num_blocks_proposed >= (sr.block_quota * 0.8)
          and (snr.challenges_received = 0 or
               (1.0 - (snr.challenges_failed::float / snr.challenges_received)) >= 0.8)
    ) as healthy_validators
from sla_rollups sr
left join sla_node_reports snr on snr.sla_rollup_id = sr.id
where sr.id = any($1::int[])
group by sr.id;

-- name: GetChallengeStatisticsForBlockRange :many
-- Get challenge statistics for validators in a block range
select
    snr.validator_address as address,
    sum(snr.challenges_received) as challenges_received,
    sum(snr.challenges_failed) as challenges_failed
from sla_node_reports snr
join sla_rollups sr on sr.id = snr.sla_rollup_id
where sr.block_start >= $1 and sr.block_end <= $2
group by snr.validator_address;

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
