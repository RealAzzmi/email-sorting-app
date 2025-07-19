create table accounts (
    id bigserial primary key,
    email varchar(256) unique not null,
    name varchar(256) not null,
    access_token text not null,
    refresh_token text,
    token_expiry timestamp with time zone,
    last_sync_history_id varchar(64),
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

-- Junction table for many-to-many relationship between emails and categories
create table email_categories (
    id bigserial primary key,
    email_id bigint not null references emails(id) on delete cascade,
    category_id bigint not null references categories(id) on delete cascade,
    created_at timestamp with time zone not null default now(),
    unique(email_id, category_id)
);

create index idx_emails_account_id on emails(account_id);
create index idx_categories_account_id on categories(account_id);
create index idx_email_categories_email_id on email_categories(email_id);
create index idx_email_categories_category_id on email_categories(category_id);