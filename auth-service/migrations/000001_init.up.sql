CREATE TABLE ROLE_TYPES
(
    id   serial not null unique,
    name varchar unique
);


INSERT INTO ROLE_TYPES
VALUES (1, 'user'),
       (2, 'admin'),
       (3, 'university');

CREATE TABLE USERS
(
    id            serial                              not null unique,
    username      varchar(255)                        not null unique,
    email         varchar(255)                        not null unique,
    password_hash text                                not null,

    phone         varchar(255) unique,
    avatar        varchar(255),

    first_name    varchar(255),
    last_name     varchar(255),
    middle_name   varchar(255),

    role_id       int                                 not null default 1,

    is_confirm    bool      default false,

    created_at    TIMESTAMP default CURRENT_TIMESTAMP not null,
    updated_at    TIMESTAMP default CURRENT_TIMESTAMP not null,

    FOREIGN KEY (role_id) REFERENCES ROLE_TYPES (id)
);

CREATE
OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at
= now(); -- Обновить updated_at текущим значением времени
RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER update_timestamp_trigger
    BEFORE UPDATE
    ON USERS
    FOR EACH ROW
    EXECUTE PROCEDURE update_timestamp();



CREATE TABLE REFRESH_TOKENS
(
    id            serial             not null unique,
    user_id       int                not null,
    refresh_token varchar(255)       not null,
    expire_at     TIMESTAMP          not null,
    black_list    bool default false not null,
    FOREIGN KEY (user_id) REFERENCES USERS (id) ON DELETE CASCADE
);

CREATE TABLE TOKEN_TYPES
(
    id          serial       not null unique,
    description varchar(255) not null unique
);

INSERT INTO TOKEN_TYPES
VALUES (1, 'EMAIL_VERIFY'),
       (2, 'PASSWORD_RESET');

CREATE TABLE USER_TOKENS
(
    id          serial                              not null unique,
    user_id     int                                 not null,
    token_type  int                                 not null,
    token_value varchar(255)                        not null unique,

    created_at  TIMESTAMP default CURRENT_TIMESTAMP not null,
    expire_at   TIMESTAMP                           not null,
    black_list  bool      default false             not null,

    FOREIGN KEY (user_id) REFERENCES USERS (id) on DELETE CASCADE,
    FOREIGN KEY (token_type) REFERENCES TOKEN_TYPES (id) on DELETE CASCADE
);