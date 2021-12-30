-- kick all
--DELETE from customers;
-- sv_restart 1 
--ALTER SEQUENCE customers_id_seq RESTART with 1;
--update customers set active = false where id = 7;
--update customers set active = true where id = 8;

CREATE TABLE customers_tokens
(
    token text not null UNIQUE,
    customer_id bigint not null references customers,
    expire timestamp not null default CURRENT_TIMESTAMP + INTERVAL '1 hour',
    created timestamp not null default CURRENT_TIMESTAMP
);