BEGIN;
SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = ON;
SET check_function_bodies = FALSE;
SET client_min_messages = WARNING;
SET search_path = public,
    extensions;
SET default_tablespace = '';
SET default_with_oids = FALSE;
-- EXTENSIONS --
CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- TABLES --
CREATE TABLE public.mc_user (
    id SERIAL PRIMARY KEY,
    email varchar(255) UNIQUE NOT NULL,
    password varchar(255) NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW(),
    updated_at timestamp NOT NULL DEFAULT NOW()
);

CREATE USER mc_service WITH ENCRYPTED PASSWORD '000000';
GRANT CONNECT ON DATABASE mc TO mc_service;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO mc_service;
GRANT ALL PRIVILEGES ON TABLE public.mc_user TO mc_service;