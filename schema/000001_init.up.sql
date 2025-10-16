CREATE TABLE users (
                       id serial not null unique,
                       name varchar(255) not null,
                       username varchar(255) not null unique,
                       password_hash varchar(255) not null
);

CREATE TABLE todo_lists (
                            id serial not null unique,
                            title varchar(255) not null,
                            description varchar(255),
                            archived boolean not null default false,
                            created_at timestamp with time zone not null default current_timestamp,
                            updated_at timestamp with time zone not null default current_timestamp,
                            color varchar(50) default '#000000',
                            priority integer not null default 0
);

CREATE TABLE users_lists (
                             id serial not null unique,
                             user_id int references users (id) on delete cascade not null,
                             list_id int references todo_lists (id) on delete cascade not null
);

CREATE TABLE todo_items (
                            id serial not null unique,
                            title varchar(255) not null,
                            description varchar(255),
                            done boolean not null default false,
                            archived boolean not null default false,
                            created_at timestamp with time zone not null default current_timestamp,
                            updated_at timestamp with time zone not null default current_timestamp
);

CREATE TABLE lists_items (
                             id serial not null unique,
                             item_id int references todo_items (id) on delete cascade not null,
                             list_id int references todo_lists (id) on delete cascade not null
);