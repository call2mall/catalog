create schema if not exists public;
create schema if not exists asin;
create schema if not exists catalog;
create schema if not exists access;
create schema if not exists messaging;

do
$$
    begin
        if not exists(select 1 from pg_type where typname = 'condition') then
            create type public.condition as enum ('returned', 'refurbished', 'unchecked', 'defective', 'a-ware', 'b-ware', 'c-ware');
        end if;
    end
$$;

do
$$
    begin
        if not exists(select 1 from pg_type where typname = 'queue_state') then
            create type public.queue_state as enum ('pending', 'executing', 'done', 'fail');
        end if;
    end
$$;

do
$$
    begin
        if not exists(select 1 from pg_type where typname = 'user_role') then
            create type public.user_role as enum ('admin', 'manager', 'broker', 'customer');
        end if;
    end
$$;

do
$$
    begin
        if not exists(select 1 from pg_type where typname = 'otp_type') then
            create type public.otp_type as enum ('initiation', 'recovery');
        end if;
    end
$$;

create table if not exists catalog.grabber
(
    uid       int  not null
        constraint grabber_pk
            primary key,
    url       text not null,
    timestamp timestamp default current_timestamp
);

create unique index if not exists grabber_url_uix
    on catalog.grabber (url);

create table if not exists asin.image
(
    hash  varchar(64) not null primary key,
    bytes bytea       null
);

create unique index if not exists image_hash_uix
    on asin.image (hash);

create table if not exists asin.category
(
    id           serial      not null
        constraint category_pk
            primary key,
    name         varchar(64) not null,
    l8n          text        null,
    is_available bool        not null default false
);

create unique index if not exists category_name_uix
    on asin.category (name);

create table if not exists asin.list
(
    asin        varchar(64) not null
        constraint list_pk
            primary key,
    category_id int         null
        constraint list_category_fk
            references asin.category
            on update cascade on delete set null,
    title       text        null,
    l8n         text        null,
    image_hash  varchar(64) null
        constraint list_image_fk
            references asin.image
            on update cascade on delete set null,
    timestamp   timestamp   not null default current_timestamp,
    is_ready    bool        not null default false
);

create table asin.origin
(
    asin varchar(64)  not null
        constraint origin_list_asin_fk
            references asin.list
            on update cascade on delete cascade,
    lang char(2)      not null,
    url  varchar(1024) not null
);

create unique index origin_asin_lang_uix
    on asin.origin (asin, lang);

alter table asin.origin
    add constraint origin_pk
        primary key (asin, lang);

create index origin_asin_ix
    on asin.origin (asin);

create table if not exists asin.searcher_queue
(
    asin    varchar(64)                                  not null
        constraint searcher_asin_asin_fk
            references asin.list (asin)
            on update cascade on delete cascade
        primary key,
    state   public.queue_state default 'pending'         not null,
    added   timestamp          default current_timestamp not null,
    updated timestamp          default null
);

create table if not exists asin.enricher_queue
(
    asin    varchar(64)                                  not null
        constraint enricher_asin_asin_fk
            references asin.list (asin)
            on update cascade on delete cascade
        primary key,
    state   public.queue_state default 'pending'         not null,
    added   timestamp          default current_timestamp not null,
    updated timestamp          default null
);

create table if not exists asin.publisher_queue
(
    asin    varchar(64)                                  not null
        constraint publisher_asin_asin_fk
            references asin.list (asin)
            on update cascade on delete cascade
        primary key,
    state   public.queue_state default 'pending'         not null,
    added   timestamp          default current_timestamp not null,
    updated timestamp          default null
);

create table if not exists catalog.unit
(
    id            bigserial   not null
        constraint unit_pk
            primary key,
    warehouse_id  varchar(64) not null,
    ean           varchar(64) not null,
    asin          varchar(64) not null
        constraint unit_asin_fk
            references asin.list
            on update cascade on delete cascade,
    sku           varchar(64) null,
    condition     text        not null,
    quantity      int         not null default 1,
    unit_cost     int,
    unit_discount int         null,
    retail_price  int         null,
    timestamp     timestamp   not null default current_timestamp,
    is_published  bool        not null default false
);

create unique index if not exists unit_warehouse_ean_asin_uix
    on catalog.unit (warehouse_id, ean, asin);

create table if not exists access.user
(
    id         bigserial                      not null
        constraint user_pk
            primary key,
    email      varchar(256)                   not null,
    pass       bytea                          null,
    salt       bytea                          not null,
    name       bytea                          null,
    surname    bytea                          null,
    is_removed bool             default false not null,
    role       public.user_role default 'customer',
    check_sum  bytea                          not null
);

create unique index if not exists user_email_uix
    on access.user (email);

create table if not exists catalog.catalog
(
    id                 bigserial not null
        constraint catalog_pk
            primary key,
    unit_id            bigint    not null
        constraint catalog_unit_id_fk
            references catalog.unit
            on update cascade on delete cascade,
    reserved_timestamp timestamp default null,
    reserved_ttl       int       default null,
    reserved_by        bigint    default null
        constraint catalog_user_id_fk
            references access."user"
            on update cascade on delete cascade
);

create table if not exists access.user_otp
(
    id        serial                              not null
        constraint user_otp_pk
            primary key,
    user_id   int
        constraint user_otp_user_id_fk
            references access.user
            on update cascade on delete cascade,
    pass      bytea                               not null,
    type      public.otp_type                     not null,
    timestamp timestamp default current_timestamp not null,
    ttl       int                                 not null
);

create unique index if not exists user_otp_pass_uix
    on access.user_otp (pass);

create table if not exists messaging.email
(
    id        serial                                       not null
        constraint email_pk
            primary key,
    receiver  bytea                                        not null,
    sender    bytea                                        not null,
    subject   bytea                                        not null,
    text      bytea                                        not null,
    state     public.queue_state default 'pending'         not null,
    salt      bytea                                        not null,
    timestamp timestamp          default current_timestamp not null,
    reason    text
);

create table if not exists messaging.attachment
(
    id   serial       not null
        constraint attachment_pk
            primary key,
    name varchar(256) not null,
    data bytea        not null
);

create table if not exists messaging.email_attachment
(
    email_id      int not null
        constraint email_attachment_email_id_fk
            references messaging.email
            on update cascade on delete cascade,
    attachment_id int not null
        constraint email_attachment_attachment_email_id_fk
            references messaging.attachment (id)
            on update cascade on delete cascade
);

create unique index if not exists email_attachment_email_id_attachment_id_uix
    on messaging.email_attachment (email_id, attachment_id);

create table if not exists messaging.email_template
(
    name    varchar(256) not null
        constraint email_template_pk
            primary key,
    subject text         not null,
    text    text         not null
);
