-- ========================================
-- blocks writes
-- ========================================

-- name: InsertBlock :exec
insert into blocks (height, hash, block_time, proposer_address, data)
values ($1, $2, $3, $4, $5);

-- name: UpsertBlock :exec
insert into blocks (height, hash, block_time, proposer_address, data)
values ($1, $2, $3, $4, $5)
on conflict (height) do update set
    hash = excluded.hash,
    block_time = excluded.block_time,
    proposer_address = excluded.proposer_address,
    data = excluded.data;

-- name: DeleteBlock :exec
delete from blocks where height = $1;

-- ========================================
-- transactions writes
-- ========================================

-- name: InsertTransaction :exec
insert into transactions (tx_hash, block_height, tx_index, tx_type, proposer, sender, data)
values ($1, $2, $3, $4, $5, $6, $7)
on conflict (tx_hash) do nothing;

-- name: UpsertTransaction :exec
insert into transactions (tx_hash, block_height, tx_index, tx_type, proposer, sender, data)
values ($1, $2, $3, $4, $5, $6, $7)
on conflict (tx_hash) do update set
    block_height = excluded.block_height,
    tx_index = excluded.tx_index,
    tx_type = excluded.tx_type,
    proposer = excluded.proposer,
    sender = excluded.sender,
    data = excluded.data;

-- name: DeleteTransaction :exec
delete from transactions where tx_hash = $1;

-- name: DeleteTransactionsByBlock :exec
delete from transactions where block_height = $1;

-- ========================================
-- validators writes
-- ========================================

-- name: InsertValidator :exec
insert into validators (address, comet_address, endpoint, node_type, spid, voting_power, status, registered_at)
values ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: UpsertValidator :exec
insert into validators (address, comet_address, endpoint, node_type, spid, voting_power, status, registered_at)
values ($1, $2, $3, $4, $5, $6, $7, $8)
on conflict (address) do update set
    comet_address = excluded.comet_address,
    endpoint = excluded.endpoint,
    node_type = excluded.node_type,
    spid = excluded.spid,
    voting_power = excluded.voting_power,
    status = excluded.status,
    registered_at = excluded.registered_at,
    updated_at = now();

-- name: UpdateValidatorStatus :exec
update validators set
    status = $2,
    deregistered_at = $3,
    updated_at = now()
where address = $1;

-- name: UpdateValidatorVotingPower :exec
update validators set
    voting_power = $2,
    updated_at = now()
where address = $1;

-- name: DeleteValidator :exec
delete from validators where address = $1;

-- ========================================
-- validator events writes
-- ========================================

-- name: InsertValidatorEvent :exec
insert into validator_events (validator_address, event_type, block_height, tx_hash)
values ($1, $2, $3, $4);

-- name: DeleteValidatorEventsByValidator :exec
delete from validator_events where validator_address = $1;

-- ========================================
-- sla rollups writes
-- ========================================

-- name: InsertSLARollup :one
insert into sla_rollups (block_start, block_end, validator_count, block_quota, tx_hash, block_height)
values ($1, $2, $3, $4, $5, $6)
returning id;

-- name: DeleteSLARollup :exec
delete from sla_rollups where id = $1;

-- ========================================
-- sla node reports writes
-- ========================================

-- name: InsertSLANodeReport :exec
insert into sla_node_reports (sla_rollup_id, validator_address, num_blocks_proposed, challenges_received, challenges_failed)
values ($1, $2, $3, $4, $5);

-- name: UpsertSLANodeReport :exec
insert into sla_node_reports (sla_rollup_id, validator_address, num_blocks_proposed, challenges_received, challenges_failed)
values ($1, $2, $3, $4, $5)
on conflict (sla_rollup_id, validator_address) do update set
    num_blocks_proposed = excluded.num_blocks_proposed,
    challenges_received = excluded.challenges_received,
    challenges_failed = excluded.challenges_failed,
    created_at = now();

-- name: DeleteSLANodeReportsByRollup :exec
delete from sla_node_reports where sla_rollup_id = $1;

-- name: DeleteSLANodeReportsByValidator :exec
delete from sla_node_reports where validator_address = $1;

-- ========================================
-- accounts writes
-- ========================================

-- name: InsertAccount :exec
insert into accounts (address, user_id)
values ($1, $2);

-- name: UpsertAccount :exec
insert into accounts (address, user_id)
values ($1, $2)
on conflict (address) do update set
    user_id = excluded.user_id;

-- name: DeleteAccount :exec
delete from accounts where address = $1;

-- ========================================
-- plays writes
-- ========================================

-- name: InsertPlay :exec
insert into plays (user_id, track_id, city, region, country, latitude, longitude, played_at, listened_at, recorded_at, block_height, tx_hash)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: DeletePlay :exec
delete from plays where id = $1;

-- name: DeletePlaysByBlock :exec
delete from plays where block_height = $1;

-- ========================================
-- manage entities writes
-- ========================================

-- name: InsertManageEntity :exec
insert into manage_entities (address, entity_type, entity_id, action, metadata, signature, signer, nonce, block_height, tx_hash)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: DeleteManageEntity :exec
delete from manage_entities where id = $1;

-- name: DeleteManageEntitiesByBlock :exec
delete from manage_entities where block_height = $1;

-- ========================================
-- stats writes (manual updates)
-- ========================================

-- name: UpdateChainStats :exec
select update_chain_stats($1, $2, $3, $4);

-- name: IncrementTxTypeStats :exec
select increment_tx_type_stats($1, $2);

-- name: UpdateValidatorStats :exec
select update_validator_stats($1, $2, $3, $4);

-- name: RecalculateValidatorCounts :exec
select recalculate_validator_counts();

-- name: CleanupTransactionWindows :one
select cleanup_transaction_windows();

-- ========================================
-- transaction windows writes
-- ========================================

-- name: InsertTransactionWindow :exec
insert into transaction_windows (tx_hash, tx_type, block_time)
values ($1, $2, $3)
on conflict (tx_hash) do nothing;

-- name: DeleteOldTransactionWindows :exec
delete from transaction_windows
where block_time < (
    select latest_block_time - interval '30 days'
    from chain_stats
    where id = 1
);

-- ========================================
-- batch operations
-- ========================================

-- name: BatchInsertTransactions :copyfrom
insert into transactions (tx_hash, block_height, tx_index, tx_type, proposer, sender, data)
values ($1, $2, $3, $4, $5, $6, $7);

-- name: BatchInsertSLANodeReports :copyfrom
insert into sla_node_reports (sla_rollup_id, validator_address, num_blocks_proposed, challenges_received, challenges_failed)
values ($1, $2, $3, $4, $5);

-- name: BatchInsertPlays :copyfrom
insert into plays (user_id, track_id, city, region, country, latitude, longitude, played_at, listened_at, recorded_at, block_height, tx_hash)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
