create table if not exists image
(
    hash  varchar(64) not null primary key,
    image bytea       null
);

create unique index if not exists image_hash_uix
    on image (hash);

create table if not exists category
(
    id   serial      not null
        constraint category_pk
            primary key,
    name varchar(64) not null,
    l8n  text        null
);

create unique index if not exists category_name_uix
    on category (name);

create table if not exists sku
(
    id            bigserial           not null
        constraint sku_pk
            primary key,
    ean           varchar(64)         not null,
    warehouse_id  varchar(64)         not null,
    avides_sku    varchar(64)         not null,
    category_id   int                 not null
        constraint sku_category_fk
            references category
            on update cascade on delete cascade,
    title         text                not null,
    asin          varchar(64),
    condition     text                not null,
    l8n           text                null,
    quantity      int       default 1 not null,
    unit_cost     int,
    unit_discount int,
    retail_price  int,
    image_hash    varchar(64)         null,
    timestamp     timestamp default current_timestamp
);

create unique index if not exists sku_ean_warehouse_sku_uix
    on sku (ean, warehouse_id, avides_sku);