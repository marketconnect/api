CREATE TABLE public.mc_users (
    id SERIAL PRIMARY KEY,
    email varchar(255) UNIQUE NOT NULL,
    password varchar(255) NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW(),
    updated_at timestamp NOT NULL DEFAULT NOW()
);
CREATE TABLE public.phrases (id SERIAL PRIMARY KEY, content text UNIQUE);
CREATE TABLE public.mc_user_phrase (
    user_id integer NOT NULL REFERENCES public.mc_users(id),
    phrase_id integer NOT NULL REFERENCES public.phrases(id),
    PRIMARY KEY (user_id, phrase_id)
);
CREATE TABLE public.ranks (
    id SERIAL PRIMARY KEY,
    mp VARCHAR (10),
    user_id INTEGER REFERENCES mc_users(id),
    phrase_id INTEGER REFERENCES phrases(id),
    rank INTEGER,
    paid_rank INTEGER,
    created_at date NOT NULL DEFAULT CURRENT_DATE,
    updated_at timestamp NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_mp_user_id_phrase_id_created_at UNIQUE (mp, user_id, phrase_id, created_at)
);
CREATE TABLE public.mc_products (
    product INTEGER UNIQUE NOT NULL,
    user_id INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES public.mc_users(id)
);
-- CREATE USER IF NOT EXISTS mc_service WITH ENCRYPTED PASSWORD '000000';
GRANT CONNECT ON DATABASE mc TO mc_service;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO mc_service;
GRANT ALL PRIVILEGES ON TABLE public.mc_users TO mc_service;
GRANT ALL PRIVILEGES ON TABLE public.phrases TO mc_service;
GRANT ALL PRIVILEGES ON TABLE public.mc_user_phrase TO mc_service;
GRANT ALL PRIVILEGES ON TABLE public.ranks TO mc_service;