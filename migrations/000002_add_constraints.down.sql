ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_unique;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_name_unique;
ALTER TABLE activities DROP CONSTRAINT IF EXISTS user_id_fk;
ALTER TABLE answers DROP CONSTRAINT IF EXISTS question_id_fk;