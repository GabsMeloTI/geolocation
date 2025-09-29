create table payment_hist (
                              id BIGSERIAL primary key,
                              user_id BIGINT not null,
                              email varchar(255) not null,
                              name varchar(100) not null,
                              value float not null,
                              method varchar(100) not null,
                              automatic boolean not null,
                              payment_date timestamp not null,
                              payment_expireted timestamp not null,
                              payment_status varchar(20) not null,
                              currency varchar(20) not null,
                              invoice varchar(100) not null,
                              customer varchar(100) not null,
                              interval varchar(50) not null
)