CREATE TABLE refresh_tokens 
(
    id serial primary key,
    user_id int references users (id) on delete cascade not null,
    token varchar (255) not null unique,
    expires_date timestamp not null
);