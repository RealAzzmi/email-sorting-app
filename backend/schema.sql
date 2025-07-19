create table accounts (
    id bigserial primary key,
    email varchar(256) unique not null,
    name varchar(256) not null,
    access_token text not null,
    refresh_token text,
    token_expiry timestamp with time zone,
    created_at timestamp with time zone not null default now(),
    updated_at timestamp with time zone not null default now()
);

CREATE TABLE categories (
    id bigserial primary key,
    account_id bigint not null references accounts(id) on delete cascade,
    name varchar(256) not null,
    description varchar(2048),
    created_at timestamp with time zone not null default now(),
    updated_at timestamp with time zone not null default now()
);

create table emails (
    id bigserial primary key,
    account_id bigint not null references accounts(id) on delete cascade,
    category_id bigint not null references categories(id) on delete set null,
    gmail_message_id varchar(256) unique not null,
    sender text,
    subject text,
    body text,
    ai_summary text,
    received_at timestamp with time zone,
    is_archived_in_gmail boolean not null default false,
    unsubscribe_link text,
    created_at timestamp with time zone not null default now(),
    updated_at timestamp with time zone not null default now(),
    unique(account_id, gmail_message_id)
);

create index idx_emails_account_id on emails(account_id);
create index idx_emails_category_id on emails(category_id);
create index idx_categories_account_id on categories(account_id);