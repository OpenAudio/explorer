-- +goose Up
-- ========================================
-- CORE BLOCKCHAIN TABLES
-- ========================================

-- Blocks table: stores block header information
create table if not exists blocks (
    height bigint primary key,
    hash text not null unique,
    block_time timestamp not null,
    proposer_address text not null,
    data bytea,
    created_at timestamp not null default now()
);

create index idx_blocks_proposer on blocks(proposer_address);
create index idx_blocks_time on blocks(block_time desc);
create index idx_blocks_hash on blocks(hash);

-- Transactions table: stores all blockchain transactions
create table if not exists transactions (
    id serial,
    tx_hash text primary key,
    block_height bigint not null references blocks(height) on delete cascade,
    tx_index integer not null,
    tx_type text not null,
    proposer text not null,
    sender text not null,
    data bytea,
    created_at timestamp not null default now()
);

create index idx_transactions_block on transactions(block_height);
create index idx_transactions_type on transactions(tx_type);
create index idx_transactions_sender on transactions(sender);
create index idx_transactions_proposer on transactions(proposer);
create index idx_transactions_time on transactions(created_at desc);
create unique index idx_transactions_block_index on transactions(block_height, tx_index);

-- ========================================
-- VALIDATOR & CONSENSUS TABLES
-- ========================================

-- Validators table: stores validator registration and status
create table if not exists validators (
    address text primary key,
    comet_address text not null,
    endpoint text not null,
    node_type text not null,
    spid text not null,
    voting_power bigint not null default 0,
    status text not null default 'active', -- active, deregistered, misbehavior_deregistered
    registered_at bigint not null default 0, -- block height
    deregistered_at bigint,
    created_at timestamp not null default now(),
    updated_at timestamp
);

create index idx_validators_status on validators(status);
create index idx_validators_comet on validators(comet_address);
create index idx_validators_endpoint on validators(endpoint);
create index idx_validators_node_type on validators(node_type);
create index idx_validators_voting_power on validators(voting_power desc);

-- Validator events: track registration/deregistration history
create table if not exists validator_events (
    id serial primary key,
    validator_address text not null references validators(address) on delete cascade,
    event_type text not null, -- registration, deregistration, misbehavior_deregistration
    block_height bigint not null references blocks(height) on delete cascade,
    tx_hash text not null references transactions(tx_hash) on delete cascade,
    created_at timestamp not null default now()
);

create index idx_validator_events_address on validator_events(validator_address);
create index idx_validator_events_type on validator_events(event_type);
create index idx_validator_events_block on validator_events(block_height);
create index idx_validator_events_time on validator_events(created_at desc);

-- ========================================
-- SLA & PERFORMANCE TABLES
-- ========================================

-- SLA Rollups: stores SLA rollup metadata
create table if not exists sla_rollups (
    id serial primary key,
    block_start bigint not null references blocks(height) on delete cascade,
    block_end bigint not null references blocks(height) on delete cascade,
    validator_count integer not null,
    block_quota integer not null,
    tx_hash text not null,
    block_height bigint not null references blocks(height) on delete cascade,
    created_at timestamp not null default now()
);

create index idx_sla_rollups_block_range on sla_rollups(block_start, block_end);
create index idx_sla_rollups_id on sla_rollups(id desc);
create index idx_sla_rollups_time on sla_rollups(created_at desc);
create index idx_sla_rollups_tx on sla_rollups(tx_hash);

-- SLA Node Reports: stores per-validator SLA performance for each rollup
create table if not exists sla_node_reports (
    id serial primary key,
    sla_rollup_id integer not null references sla_rollups(id) on delete cascade,
    validator_address text not null references validators(address) on delete cascade,
    num_blocks_proposed integer not null default 0,
    challenges_received integer not null default 0,
    challenges_failed integer not null default 0,
    created_at timestamp not null default now()
);

create index idx_sla_node_reports_rollup on sla_node_reports(sla_rollup_id);
create index idx_sla_node_reports_validator on sla_node_reports(validator_address);
create index idx_sla_node_reports_composite on sla_node_reports(validator_address, sla_rollup_id desc);
create unique index idx_sla_node_reports_unique on sla_node_reports(sla_rollup_id, validator_address);

-- ========================================
-- APPLICATION TABLES
-- ========================================

-- Accounts: maps addresses to user IDs
create table if not exists accounts (
    address text primary key,
    user_id text
);

create index idx_accounts_user_id on accounts(user_id);

-- Plays: stores music play events
create table if not exists plays (
    id serial primary key,
    user_id text not null,
    track_id text not null,
    city text not null,
    region text not null,
    country text not null,
    latitude numeric(9,6),
    longitude numeric(9,6),
    played_at timestamp not null,
    listened_at timestamp not null,
    recorded_at timestamp not null,
    block_height bigint not null references blocks(height) on delete cascade,
    tx_hash text not null references transactions(tx_hash) on delete cascade,
    created_at timestamp not null default now()
);

create index idx_plays_user on plays(user_id);
create index idx_plays_track on plays(track_id);
create index idx_plays_country on plays(country);
create index idx_plays_time on plays(played_at desc);
create index idx_plays_block on plays(block_height);
create index idx_plays_tx on plays(tx_hash);

-- Manage Entities: stores entity management transactions (create/update/delete)
create table if not exists manage_entities (
    id serial primary key,
    address text not null,
    entity_type text not null,
    entity_id bigint not null,
    action text not null, -- create, update, delete
    metadata jsonb,
    signature text not null,
    signer text not null,
    nonce text not null,
    block_height bigint not null references blocks(height) on delete cascade,
    tx_hash text not null references transactions(tx_hash) on delete cascade,
    created_at timestamp not null default now()
);

create index idx_manage_entities_address on manage_entities(address);
create index idx_manage_entities_type on manage_entities(entity_type);
create index idx_manage_entities_entity_id on manage_entities(entity_id);
create index idx_manage_entities_action on manage_entities(action);
create index idx_manage_entities_block on manage_entities(block_height);
create index idx_manage_entities_tx on manage_entities(tx_hash);
create index idx_manage_entities_time on manage_entities(created_at desc);
create index idx_manage_entities_composite on manage_entities(entity_type, entity_id);

-- Storage Proofs: stores storage proof submissions
create table if not exists storage_proofs (
    id serial primary key,
    prover_address text not null,
    challenge_id text not null,
    proof_data bytea,
    block_height bigint not null references blocks(height) on delete cascade,
    tx_hash text not null references transactions(tx_hash) on delete cascade,
    created_at timestamp not null default now()
);

create index idx_storage_proofs_prover on storage_proofs(prover_address);
create index idx_storage_proofs_challenge on storage_proofs(challenge_id);
create index idx_storage_proofs_block on storage_proofs(block_height);
create index idx_storage_proofs_tx on storage_proofs(tx_hash);

-- Storage Proof Verifications: stores storage proof verification results
create table if not exists storage_proof_verifications (
    id serial primary key,
    challenge_id text not null,
    verifier_address text not null,
    is_valid boolean not null,
    verification_data bytea,
    block_height bigint not null references blocks(height) on delete cascade,
    tx_hash text not null references transactions(tx_hash) on delete cascade,
    created_at timestamp not null default now()
);

create index idx_storage_proof_verifications_challenge on storage_proof_verifications(challenge_id);
create index idx_storage_proof_verifications_verifier on storage_proof_verifications(verifier_address);
create index idx_storage_proof_verifications_block on storage_proof_verifications(block_height);
create index idx_storage_proof_verifications_tx on storage_proof_verifications(tx_hash);

-- ========================================
-- STATS TABLES (SINGLE ROW FOR ATOMIC UPDATES)
-- ========================================

-- Chain stats: single row with running totals for dashboard
create table if not exists chain_stats (
    id integer primary key default 1 check (id = 1), -- enforce single row
    latest_block_height bigint not null default 0,
    latest_block_hash text,
    latest_block_time timestamp,
    total_transactions bigint not null default 0,
    total_validators integer not null default 0,
    active_validators integer not null default 0,
    avg_block_time_seconds numeric(10,4),
    transactions_24h bigint not null default 0,
    transactions_7d bigint not null default 0,
    transactions_30d bigint not null default 0,
    updated_at timestamp not null default now()
);

-- Insert initial row
insert into chain_stats (id) values (1) on conflict do nothing;

-- Transaction type stats: one row per transaction type with running totals
create table if not exists transaction_type_stats (
    tx_type text primary key,
    total_count bigint not null default 0,
    latest_transaction timestamp,
    updated_at timestamp not null default now()
);

-- Rolling window table: stores recent transactions for window calculations
-- Automatically cleaned up to only keep transactions within 30 days of latest block
create table if not exists transaction_windows (
    tx_hash text primary key references transactions(tx_hash) on delete cascade,
    tx_type text not null,
    block_time timestamp not null  -- use block_time, not created_at!
);

create index idx_transaction_windows_block_time on transaction_windows(block_time desc);
create index idx_transaction_windows_type_time on transaction_windows(tx_type, block_time desc);

-- Validator stats: one row per validator with cached aggregates
create table if not exists validator_stats (
    validator_address text primary key references validators(address) on delete cascade,
    total_blocks_proposed bigint not null default 0,
    total_sla_rollups integer not null default 0,
    sla_rollups_passed integer not null default 0,
    sla_rollups_failed integer not null default 0,
    last_block_proposed_height bigint,
    last_sla_rollup_id integer,
    updated_at timestamp not null default now()
);

-- ========================================
-- HELPER FUNCTIONS
-- ========================================

-- Function to update chain stats atomically
create or replace function update_chain_stats(
    p_block_height bigint,
    p_block_hash text,
    p_block_time timestamp,
    p_tx_count integer default 0
)
returns void as $$
begin
    update chain_stats set
        latest_block_height = greatest(latest_block_height, p_block_height),
        latest_block_hash = case
            when p_block_height > latest_block_height then p_block_hash
            else latest_block_hash
        end,
        latest_block_time = case
            when p_block_height > latest_block_height then p_block_time
            else latest_block_time
        end,
        total_transactions = total_transactions + p_tx_count,
        updated_at = now()
    where id = 1;
end;
$$ language plpgsql;

-- Function to increment transaction type stats
create or replace function increment_tx_type_stats(
    p_tx_type text,
    p_timestamp timestamp default now()
)
returns void as $$
begin
    insert into transaction_type_stats (tx_type, total_count, latest_transaction)
    values (p_tx_type, 1, p_timestamp)
    on conflict (tx_type) do update set
        total_count = transaction_type_stats.total_count + 1,
        latest_transaction = greatest(transaction_type_stats.latest_transaction, p_timestamp),
        updated_at = now();
end;
$$ language plpgsql;

-- Function to update validator stats after SLA rollup
create or replace function update_validator_stats(
    p_validator_address text,
    p_blocks_proposed integer,
    p_sla_rollup_id integer,
    p_passed boolean
)
returns void as $$
begin
    insert into validator_stats (validator_address, total_blocks_proposed, total_sla_rollups,
                                  sla_rollups_passed, sla_rollups_failed, last_sla_rollup_id)
    values (p_validator_address, p_blocks_proposed, 1,
            case when p_passed then 1 else 0 end,
            case when p_passed then 0 else 1 end,
            p_sla_rollup_id)
    on conflict (validator_address) do update set
        total_blocks_proposed = validator_stats.total_blocks_proposed + p_blocks_proposed,
        total_sla_rollups = validator_stats.total_sla_rollups + 1,
        sla_rollups_passed = validator_stats.sla_rollups_passed + case when p_passed then 1 else 0 end,
        sla_rollups_failed = validator_stats.sla_rollups_failed + case when p_passed then 0 else 1 end,
        last_sla_rollup_id = p_sla_rollup_id,
        updated_at = now();
end;
$$ language plpgsql;

-- Function to get windowed transaction counts (based on latest block time)
create or replace function get_transaction_window_counts()
returns table(
    window_24h bigint,
    window_7d bigint,
    window_30d bigint
) as $$
declare
    v_latest_block_time timestamp;
    v_24h_ago timestamp;
    v_7d_ago timestamp;
    v_30d_ago timestamp;
begin
    -- Get latest block time from chain_stats
    select latest_block_time into v_latest_block_time
    from chain_stats
    where id = 1;

    -- If no blocks yet, return zeros
    if v_latest_block_time is null then
        return query select 0::bigint, 0::bigint, 0::bigint;
        return;
    end if;

    -- Calculate windows based on latest block time
    v_24h_ago := v_latest_block_time - interval '24 hours';
    v_7d_ago := v_latest_block_time - interval '7 days';
    v_30d_ago := v_latest_block_time - interval '30 days';

    return query
    select
        (select count(*) from transaction_windows where block_time >= v_24h_ago)::bigint,
        (select count(*) from transaction_windows where block_time >= v_7d_ago)::bigint,
        (select count(*) from transaction_windows where block_time >= v_30d_ago)::bigint;
end;
$$ language plpgsql stable;

-- Function to cleanup old windowed transactions
-- Keeps only transactions within 30 days of latest block
create or replace function cleanup_transaction_windows()
returns bigint as $$
declare
    v_latest_block_time timestamp;
    v_cutoff_time timestamp;
    v_deleted bigint;
begin
    -- Get latest block time
    select latest_block_time into v_latest_block_time
    from chain_stats
    where id = 1;

    -- If no blocks, nothing to clean
    if v_latest_block_time is null then
        return 0;
    end if;

    -- Delete transactions older than 30 days from latest block
    v_cutoff_time := v_latest_block_time - interval '30 days';

    delete from transaction_windows
    where block_time < v_cutoff_time;

    get diagnostics v_deleted = row_count;
    return v_deleted;
end;
$$ language plpgsql;

-- Function to update validator counts in chain_stats
create or replace function recalculate_validator_counts()
returns void as $$
declare
    v_total integer;
    v_active integer;
begin
    select count(*) into v_total from validators;
    select count(*) into v_active from validators where status = 'active';

    update chain_stats set
        total_validators = v_total,
        active_validators = v_active,
        updated_at = now()
    where id = 1;
end;
$$ language plpgsql;

-- ========================================
-- TRIGGERS FOR AUTO-UPDATING STATS
-- ========================================

-- Trigger function to update chain_stats on block insert
create or replace function trigger_update_chain_stats_on_block()
returns trigger as $$
begin
    perform update_chain_stats(NEW.height, NEW.hash, NEW.block_time, 0);
    return NEW;
end;
$$ language plpgsql;

create trigger trg_block_insert_update_stats
    after insert on blocks
    for each row
    execute function trigger_update_chain_stats_on_block();

-- Trigger function to increment transaction counts
create or replace function trigger_increment_tx_stats()
returns trigger as $$
declare
    v_block_time timestamp;
begin
    -- Get block time for this transaction
    select block_time into v_block_time
    from blocks
    where height = NEW.block_height;

    -- Update chain_stats total
    update chain_stats set
        total_transactions = total_transactions + 1,
        updated_at = now()
    where id = 1;

    -- Update transaction_type_stats
    perform increment_tx_type_stats(NEW.tx_type, NEW.created_at);

    -- Add to rolling window table with block_time
    insert into transaction_windows (tx_hash, tx_type, block_time)
    values (NEW.tx_hash, NEW.tx_type, v_block_time)
    on conflict (tx_hash) do nothing;

    return NEW;
end;
$$ language plpgsql;

create trigger trg_transaction_insert_update_stats
    after insert on transactions
    for each row
    execute function trigger_increment_tx_stats();

-- Trigger function to update validator counts
create or replace function trigger_update_validator_counts()
returns trigger as $$
begin
    perform recalculate_validator_counts();
    return NEW;
end;
$$ language plpgsql;

create trigger trg_validator_insert_update_counts
    after insert on validators
    for each row
    execute function trigger_update_validator_counts();

create trigger trg_validator_update_update_counts
    after update of status on validators
    for each row
    when (OLD.status is distinct from NEW.status)
    execute function trigger_update_validator_counts();

-- +goose Down
drop trigger if exists trg_validator_update_update_counts on validators;
drop trigger if exists trg_validator_insert_update_counts on validators;
drop trigger if exists trg_transaction_insert_update_stats on transactions;
drop trigger if exists trg_block_insert_update_stats on blocks;
drop function if exists trigger_update_validator_counts();
drop function if exists trigger_increment_tx_stats();
drop function if exists trigger_update_chain_stats_on_block();
drop function if exists recalculate_validator_counts();
drop function if exists cleanup_transaction_windows();
drop function if exists get_transaction_window_counts();
drop function if exists update_validator_stats(text, integer, integer, boolean);
drop function if exists increment_tx_type_stats(text, timestamp);
drop function if exists update_chain_stats(bigint, text, timestamp, integer);
drop table if exists validator_stats cascade;
drop table if exists transaction_windows cascade;
drop table if exists transaction_type_stats cascade;
drop table if exists chain_stats cascade;
drop table if exists storage_proof_verifications cascade;
drop table if exists storage_proofs cascade;
drop table if exists manage_entities cascade;
drop table if exists plays cascade;
drop table if exists accounts cascade;
drop table if exists sla_node_reports cascade;
drop table if exists sla_rollups cascade;
drop table if exists validator_events cascade;
drop table if exists validators cascade;
drop table if exists transactions cascade;
drop table if exists blocks cascade;
